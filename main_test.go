package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMerge(t *testing.T) {
	testcases := []struct {
		intervals []interval
		expected  []interval
	}{
		// Empty input
		{
			intervals: []interval{},
			expected:  []interval{},
		},
		// Single interval input
		{
			intervals: []interval{
				{x: 1, y: 2},
			},
			expected: []interval{
				{x: 1, y: 2},
			},
		},
		// Repeated interval
		{
			intervals: []interval{
				{x: 1, y: 2},
				{x: 1, y: 2},
			},
			expected: []interval{
				{x: 1, y: 2},
			},
		},
		// Two intervals that merge
		{
			intervals: []interval{
				{x: 1, y: 2},
				{x: 2, y: 3},
			},
			expected: []interval{
				{x: 1, y: 3},
			},
		},
		// Two intervals that do not merge
		{
			intervals: []interval{
				{x: 1, y: 2},
				{x: 3, y: 4},
			},
			expected: []interval{
				{x: 1, y: 2},
				{x: 3, y: 4},
			},
		},
		// Three intervals: first and second merge
		{
			intervals: []interval{
				{x: 1, y: 2},
				{x: 2, y: 3},
				{x: 4, y: 5},
			},
			expected: []interval{
				{x: 1, y: 3},
				{x: 4, y: 5},
			},
		},
		// Three intervals: second and third merge
		{
			intervals: []interval{
				{x: 1, y: 2},
				{x: 3, y: 4},
				{x: 4, y: 5},
			},
			expected: []interval{
				{x: 1, y: 2},
				{x: 3, y: 5},
			},
		},
		// Three intervals: no merge
		{
			intervals: []interval{
				{x: 1, y: 2},
				{x: 3, y: 4},
				{x: 5, y: 6},
			},
			expected: []interval{
				{x: 1, y: 2},
				{x: 3, y: 4},
				{x: 5, y: 6},
			},
		},
		// Alternate merge | no merge
		{
			intervals: []interval{
				{x: 1, y: 2},
				{x: 2, y: 3},
				{x: 4, y: 5},
				{x: 6, y: 7},
				{x: 7, y: 8},
				{x: 9, y: 10},
			},
			expected: []interval{
				{x: 1, y: 3},
				{x: 4, y: 5},
				{x: 6, y: 8},
				{x: 9, y: 10},
			},
		},
		// Dot - no merge
		{
			intervals: []interval{
				{x: 1, y: 2},
				{x: 3, y: 3},
				{x: 4, y: 5},
			},
			expected: []interval{
				{x: 1, y: 2},
				{x: 3, y: 3},
				{x: 4, y: 5},
			},
		},
		// Dot - merge
		{
			intervals: []interval{
				{x: 1, y: 3},
				{x: 3, y: 3},
				{x: 3, y: 5},
			},
			expected: []interval{
				{x: 1, y: 5},
			},
		},
		// Testcase from assignment
		{
			intervals: []interval{
				{x: 25, y: 30},
				{x: 2, y: 19},
				{x: 14, y: 23},
				{x: 4, y: 8},
			},
			expected: []interval{
				{x: 2, y: 23},
				{x: 25, y: 30},
			},
		},
	}

	for _, test := range testcases {
		assert.Equal(t, test.expected, merge(test.intervals), fmt.Sprintf("testcase %+v", test))
	}
}

func TestMergeIfSortedAndOverlap(t *testing.T) {
	testcases := []struct {
		a       interval
		b       interval
		overlap bool
		merged  interval
	}{
		// A [x,   y]
		// B [x,y]
		{
			a:       interval{x: 1, y: 3},
			b:       interval{x: 1, y: 2},
			overlap: true,
			merged:  interval{x: 1, y: 3},
		},
		// A [x,y]
		// B [x,y]
		{
			a:       interval{x: 1, y: 2},
			b:       interval{x: 1, y: 2},
			overlap: true,
			merged:  interval{x: 1, y: 2},
		},
		// A [x,y]
		// B [x,   y]
		{
			a:       interval{x: 1, y: 2},
			b:       interval{x: 1, y: 3},
			overlap: true,
			merged:  interval{x: 1, y: 3},
		},
		// A [x,   y]
		// B   [x,   y]
		{
			a:       interval{x: 1, y: 3},
			b:       interval{x: 2, y: 4},
			overlap: true,
			merged:  interval{x: 1, y: 4},
		},
		// A [x,   y]
		// B    [x,y]
		{
			a:       interval{x: 1, y: 3},
			b:       interval{x: 2, y: 3},
			overlap: true,
			merged:  interval{x: 1, y: 3},
		},
		// A [x,     y]
		// B   [x,y]
		{
			a:       interval{x: 1, y: 4},
			b:       interval{x: 2, y: 3},
			overlap: true,
			merged:  interval{x: 1, y: 4},
		},
		// A: [x,y]
		// B:       [x,y]
		{
			a:       interval{x: 1, y: 2},
			b:       interval{x: 3, y: 4},
			overlap: false,
			merged:  interval{},
		},
		// A:   [x,y]
		// B: [x,y]
		{
			a:       interval{x: 2, y: 3},
			b:       interval{x: 1, y: 2},
			overlap: false,
			merged:  interval{},
		},
		// A:       [x,y]
		// B: [x,y]
		{
			a:       interval{x: 3, y: 4},
			b:       interval{x: 1, y: 2},
			overlap: false,
			merged:  interval{},
		},
	}

	for _, test := range testcases {
		merged, ok := test.a.mergeIfSortedAndOverlap(test.b)
		assert.Equal(t, test.overlap, ok, fmt.Sprintf("testcase: %+v", test))
		assert.Equal(t, test.merged, merged, fmt.Sprintf("testcase: %+v", test))
	}
}

func TestParse(t *testing.T) {
	testcases := []struct {
		input    string
		expected []interval
		error    error
	}{
		{
			input: "[1,2] [3, 4] [ 5,6] [   7   ,   8   ][9,10][11,12]",
			expected: []interval{
				{x: 1, y: 2},
				{x: 3, y: 4},
				{x: 5, y: 6},
				{x: 7, y: 8},
				{x: 9, y: 10},
				{x: 11, y: 12},
			},
		},
		{
			input: "input in bad format",
			error: errBadInput,
		},
		{
			input: "[1,2][,]",
			error: errBadInput,
		},
		{
			input: "[1,]",
			error: errBadInput,
		},
		{
			input: "[,2]",
			error: errBadInput,
		},
	}

	for _, test := range testcases {
		res, err := parseFromReader(strings.NewReader(test.input))
		assert.Equal(t, test.expected, res, fmt.Sprintf("testcase: %v", test))
		assert.ErrorIs(t, err, test.error, fmt.Sprintf("testcase: %v", test))
	}
}

func TestIntervalListToString(t *testing.T) {
	input := []interval{
		{x: 1, y: 2},
		{x: 3, y: 4},
		{x: 5, y: 6},
	}
	expected := "[1,2] [3,4] [5,6]"

	assert.Equal(t, expected, IntervalListToString(input))
}

func TestGetMaxWidth(t *testing.T) {
	testcases := []struct {
		intervals []interval
		expected  interval
	}{
		{
			intervals: []interval{},
			expected:  interval{},
		},
		{
			intervals: []interval{
				{x: 1, y: 2},
			},
			expected: interval{x: 1, y: 2},
		},
		{
			intervals: []interval{
				{x: 1, y: 2},
				{x: 1, y: 2},
			},
			expected: interval{x: 1, y: 2},
		},
		{
			intervals: []interval{
				{x: 1, y: 2},
				{x: 3, y: 4},
			},
			expected: interval{x: 1, y: 4},
		},
		{
			intervals: []interval{
				{x: 1, y: 5},
				{x: 3, y: 4},
			},
			expected: interval{x: 1, y: 5},
		},
		{
			intervals: []interval{
				{x: 3, y: 4},
				{x: 1, y: 5},
			},
			expected: interval{x: 1, y: 5},
		},
	}

	for _, test := range testcases {
		assert.Equal(t, test.expected, getMaxWidth(test.intervals...), fmt.Sprintf("testcase: %v", test))
	}
}

func TestProcessFile(t *testing.T) {
	testcases := []struct {
		expected    string
		inputFile   string
		maxFileSize int
	}{
		{
			expected:    "[1,3] [4,6] [7,8]",
			inputFile:   "data/simple_example.txt",
			maxFileSize: 3,
		},
		{
			expected:    "[1,3] [4,6] [7,8]",
			inputFile:   "data/simple_example.txt",
			maxFileSize: 5,
		},
		{
			expected:    "[1,3] [4,6] [7,8]",
			inputFile:   "data/simple_example.txt",
			maxFileSize: 12,
		},
		{
			expected:    "[1,3] [4,6] [7,8]",
			inputFile:   "data/simple_example.txt",
			maxFileSize: 15,
		},
		{
			expected:  "[2,23] [25,30]",
			inputFile: "data/coding_challenge.txt",
		},
	}

	for _, test := range testcases {
		resFile, err := processFile(test.inputFile, 5)
		assert.NoError(t, err)

		f, err := os.Open(resFile)
		assert.NoError(t, err)
		defer f.Close()

		b, err := io.ReadAll(f)
		assert.NoError(t, err)

		b = bytes.Trim(b, "\n")

		assert.Equal(t, test.expected, string(b))
	}

	t.Cleanup(func() {
		assert.NoError(t, os.Remove(resultFileName))
	})
}
