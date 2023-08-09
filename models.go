package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	errBadInput = errors.New("bad input")
)

type interval struct {
	x int
	y int
}

// parse scans through the reader one interval at a time,
// parsing it into an interval type.
// A slice with all intervals in the reader and an empty
// error will be returned upon success.
// Any ocurring parsing errors will be returned
// with an empty slice.
func parseFromReader(r io.Reader) ([]interval, error) {
	scanner := bufio.NewScanner(r)

	// scan inputinterval by interval
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// find *first* closing bracket
		closingIdx := bytes.IndexRune(data, ']')
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

	res := make([]interval, 0)
	for scanner.Scan() {
		t := scanner.Text()
		commaIdx := strings.IndexRune(t, ',')
		if commaIdx < 0 {
			return nil, fmt.Errorf("failed to parse interval %q: %w", t, errBadInput)
		}

		trimmedX := strings.Trim(t[:commaIdx], "[] ")

		x, err := strconv.Atoi(trimmedX)
		if err != nil {
			return nil, fmt.Errorf("failed to parse interval %q: failed to convert %q to number: %w", t, trimmedX, errBadInput)
		}

		trimmedY := strings.Trim(t[commaIdx+1:], "[] ")

		y, err := strconv.Atoi(trimmedY)
		if err != nil {
			return nil, fmt.Errorf("failed to parse interval %q: failed to convert %q to number: %w", t, trimmedY, errBadInput)
		}

		res = append(res, interval{x: x, y: y})
	}

	return res, nil
}

// IntervalListToString converts a list of intervals into
// a string of white-space separated intervals.
func IntervalListToString(list []interval) string {
	var b strings.Builder
	for i := 0; i < len(list)-1; i++ {
		b.WriteString(fmt.Sprintf("[%d,%d] ", list[i].x, list[i].y))
	}
	last := list[len(list)-1]
	b.WriteString(fmt.Sprintf("[%d,%d]", last.x, last.y))

	return b.String()
}
