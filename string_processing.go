package main

import (
	"sort"
	"strings"
)

// merge merges overlapping intervals from input.
// Non-overlapping intervals are included in the result.
// E.g.:
//
//	Input: [25,30] [2,19] [14, 23] [4,8]
//	Output: [2,23] [25,30]
//
// merge() operates in-place to efficiently manage memory.
// The underlying array will be modified.
func merge(intervals []interval) []interval {
	if len(intervals) < 2 {
		return intervals
	}

	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i].x < intervals[j].x
	})
	i := 0
	for i < len(intervals)-1 {
		merged, ok := intervals[i].mergeIfSortedAndOverlap(intervals[i+1])
		if ok {
			intervals[i] = merged
			// remove i+1 from the list
			intervals = append(intervals[:i+1], intervals[i+2:]...)
			intervals = intervals[:len(intervals):len(intervals)]
		} else {
			i++
		}
	}

	return intervals
}

// mergeIfSortedAndOverlap merges intervals a y b iff:
// they are *sorted in ascending order by left endpoint*
// **and** *overlap*.
//
// That is:
//
//   - interval a begins *before or at the same left
//     endpoint* as b .
//
//   - interval a ends *either at or after* b's left endpoint.
//
//     E.g.:
//
//     a: [x, ...->|
//     b: |<- [x, ...
//
// In that case, the function returns the merged intervals
// and true.
//
// Note that, under these circumstances, overlapping
// intervals can only fall into one of the following cases:
//
// 1) Either interval a ends before interval b:
//
//	a: ...,y]->)
//	b:    ...,y]
//
// resulting in a merged interval of [a.x, b.y], or
//
// 2) interval a ends either at the same value of interval
// b or after it:
//
//	a: ...,y]-->
//	b: ...,y]
//
// resulting in a merged interval of [a.x, a.y].
//
// For any other case, this function returns an empty
// interval and false.
//
// Use the second return value to check if the input
// intervals did indeed overlap.
func (a interval) mergeIfSortedAndOverlap(b interval) (interval, bool) {
	if a.x <= b.x && b.x <= a.y {
		if a.y < b.y {
			return interval{x: a.x, y: b.y}, true
		}
		return interval{x: a.x, y: a.y}, true
	}

	return interval{}, false
}

// processString parses a string into a slice
// of intervals and merges it.
// Upon success, it returns the merged list and
// a nil error.
// Any ocurring parsing errors will be returned
// with an empty slice.
func processString(s string) ([]interval, error) {
	list, err := parseFromReader(strings.NewReader(s))
	if err != nil {
		return nil, err
	}

	return merge(list), nil
}
