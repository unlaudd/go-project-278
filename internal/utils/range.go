// Package utils provides shared helper functions for the application.
package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseRange parses a range parameter in the format "[start,end]" and returns
// the inclusive start and end values. It supports optional whitespace inside
// the brackets (e.g., "[5, 10]" is valid).
//
// Returns an error if:
//   - the input is empty
//   - the format does not match "[int,int]"
//   - start or end cannot be parsed as integers
//   - start is negative or end is less than start
//
// Example: ParseRange("[0,10]") returns (0, 10, nil).
func ParseRange(rangeParam string) (start, end int, err error) {
	if rangeParam == "" {
		return 0, 0, fmt.Errorf("empty range")
	}

	// Remove all spaces to simplify pattern matching: "[5, 10]" → "[5,10]".
	rangeParam = strings.ReplaceAll(rangeParam, " ", "")

	// Match exactly "[digits,digits]" with start and end captured in groups.
	re := regexp.MustCompile(`^\[(\d+),(\d+)\]$`)
	matches := re.FindStringSubmatch(rangeParam)
	if len(matches) != 3 {
		return 0, 0, fmt.Errorf("invalid range format")
	}

	// Parse the captured start and end values.
	start, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, err
	}
	end, err = strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, err
	}

	// Validate logical constraints: start must be non-negative, end must be >= start.
	if start < 0 || end < start {
		return 0, 0, fmt.Errorf("invalid range values")
	}

	return start, end, nil
}
