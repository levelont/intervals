package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"sort"
)

const (
	// 2 x 20 digit integers + enough space for brackets,
	// comma and whitespace characters
	minBufferSize = 20*2 + 24

	tempDirPattern = "tmp.*"
	resultFileName = "result.txt"
)

// processFile is the entry point for file processing:
//
//   - split file in chunks of maxChunkFileSize bytes.
//   - map each file to its getMaxWidth().
//   - sort 'width->file' index in ascending order.
//   - process widths as if they were intervals.
//   - in case of an overlap, merge the files.
//   - otherwise, append them.
//   - continue until a single width is left.
//
// Upon success, the result will be written to a file,
// and its path returned, together with a nil error.
// Input parsing errors or I/O errors will interrupt
// processing and be returned accordingly with an empty string.
func processFile(filePath string, maxChunkFileSize int) (string, error) {
	tempDir, err := os.MkdirTemp(".", tempDirPattern)
	if err != nil {
		return "", err
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			log.Printf("failed to cleanup temp directory %q\n", tempDir)
		}
	}()

	index, err := splitFile(filePath, tempDir, maxChunkFileSize)
	if err != nil {
		return "", err
	}

	sort.Slice(index, func(i, j int) bool {
		return index[i].key.x < index[j].key.x
	})

	for len(index) > 1 {
		_, ok := index[0].key.mergeIfSortedAndOverlap(index[1].key)
		if ok {
			err := mergeIntervalsFromFiles(index[0].file, index[1].file)
			if err != nil {
				return "", err
			}
		} else {
			appendFiles(index[0].file, index[1].file)
		}

		// update key
		index[0].key = getMaxWidth(index[0].key, index[1].key)

		// cleanup index[1]
		err := index[1].file.Close()
		if err != nil {
			return "", err
		}

		err = os.Remove(index[1].file.Name())
		if err != nil {
			return "", err
		}

		index[1].file = nil
		index = append(index[:1], index[2:]...)
	}

	_, err = index[0].file.WriteString("\n")
	if err != nil {
		return "", err
	}

	err = os.Rename(index[0].file.Name(), resultFileName)
	if err != nil {
		return "", err
	}

	index[0].file.Close()
	if err != nil {
		return "", err
	}

	return resultFileName, nil
}

// splitFile splits interval data in multiple files of
// maxChukFileSize bytes. Intervals within each file will
// already be sorted and merged.
//
// A file index that maps intervals in a file to their
// getMaxWidth() value will be returned upon success
// with a nil error.
//
// It is the responsibility of the caller of this function
// to cleanup the files and close the file descriptors
// referenced by the index returned.
//
// Input parsing errors or I/O errors will interrupt
// processing and be returned with an empty index.
func splitFile(filePath string, tempDir string, maxChunkFileSize int) ([]fileIndex, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	buf := make([]byte, maxChunkFileSize)
	// ensure enough buffer space for scenarios including whitespace characters
	scanner.Buffer(buf, minBufferSize)

	// scan input such that the *maximum possible amount of intervals* fit in the buffer
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// find the *last* closing bracket
		closingIdx := bytes.LastIndexByte(data, ']')
		if closingIdx > 0 {
			// return remaining data past the closing bracket
			buffer := data[:closingIdx+1]

			// advance to the first rune past the closing bracket
			return closingIdx + 1, buffer, nil
		}

		// return remaining data if it's the end of the file
		if atEOF && len(data) > 0 {
			return len(data), data, nil
		}

		// continue reading
		return 0, nil, nil
	})

	var index []fileIndex
	for scanner.Scan() {
		intervals, err := parseFromReader(bytes.NewReader(scanner.Bytes()))
		if err != nil {
			return nil, err
		}

		intervals = merge(intervals)
		key := getMaxWidth(intervals...)

		f, err := os.CreateTemp(tempDir, "*")
		if err != nil {
			return nil, err
		}

		_, err = f.WriteString(IntervalListToString(intervals))
		if err != nil {
			return nil, err
		}
		index = append(index, fileIndex{key: key, file: f})
	}

	return index, nil
}

// getMaxWidth takes a list of intervals and
// returns an interval representing the maximum span of
// its values - the maximum width of the list.
//
// The width of two lists of intervals can be used
// to determine if they contain potential overlaps.
func getMaxWidth(intervals ...interval) interval {
	if len(intervals) == 0 {
		return interval{}
	}

	smallestX := intervals[0].x
	largestY := intervals[0].y
	for i := 1; i < len(intervals); i++ {
		if intervals[i].x < smallestX {
			smallestX = intervals[i].x
		}

		if intervals[i].y > largestY {
			largestY = intervals[i].y
		}
	}

	return interval{x: smallestX, y: largestY}
}

// mergeIntervalsFromFiles merges the lists of intervals
// contained in files a and b and writes them to a.
// Any ocurring I/O errors will be returned.
func mergeIntervalsFromFiles(a io.ReadWriteSeeker, b io.ReadSeeker) error {
	_, err := a.Seek(0, 0)
	if err != nil {
		return err
	}

	intervals, err := parseFromReader(a)
	if err != nil {
		return err
	}

	_, err = b.Seek(0, 0)
	if err != nil {
		return err
	}

	intervalsB, err := parseFromReader(b)
	if err != nil {
		return err
	}

	intervals = append(intervals, intervalsB...)
	intervalsB = nil
	intervals = merge(intervals)

	_, err = a.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = a.Write([]byte(IntervalListToString(intervals)))
	if err != nil {
		return err
	}

	return nil
}

// appendFiles writes the contents of b at the end of b
// with a whitespace character as a separator.
// Any ocurring I/O errors will be returned.
func appendFiles(a io.WriteSeeker, b io.ReadSeeker) error {
	_, err := b.Seek(0, 0)
	if err != nil {
		return err
	}

	bytes, err := io.ReadAll(b)
	if err != nil {
		return err
	}

	_, err = a.Seek(0, 2)
	if err != nil {
		return err
	}

	_, err = a.Write([]byte(" "))
	if err != nil {
		return err
	}

	_, err = a.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
