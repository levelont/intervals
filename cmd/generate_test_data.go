package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

const outputFile = "test_data.txt"

func main() {
	generateIntervals(30, 10000)
}

// generateIntervals produces numBuffers x intervalsPerBuffer
// intervals with a random number of non-overlapping intervals.
// Output is written to outputFile, the number of non-overlapping
// intervals printed to standard output.
func generateIntervals(numBuffers int, intervalsPerBuffer int) {
	builders := make([]strings.Builder, numBuffers)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	x := 1
	y := 2
	nonOverlappingCount := 1
	for i := 0; i < numBuffers*intervalsPerBuffer; i++ {
		builders[r.Intn(numBuffers)].WriteString(fmt.Sprintf("[%d,%d]", x, y))

		if r.Intn(100) < 50 {
			//overlap
			x = y
			y = y + 1
		} else if i < (numBuffers*intervalsPerBuffer)-1 {
			// don't overlap
			x = y + 1
			y = x + 1
			nonOverlappingCount++
		}

	}

	f, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for i := 0; i < numBuffers; i++ {
		f.WriteString(builders[i].String())
	}
	fmt.Printf("Output written to %q\n", outputFile)
	fmt.Printf("Number of non-overlapping intervals: %d\n", nonOverlappingCount)
}
