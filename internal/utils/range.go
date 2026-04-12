package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseRange парсит строку вида "[0,10]" в start=0, end=10
func ParseRange(rangeParam string) (start, end int, err error) {
	if rangeParam == "" {
		return 0, 0, fmt.Errorf("empty range")
	}
	rangeParam = strings.ReplaceAll(rangeParam, " ", "")
	re := regexp.MustCompile(`^\[(\d+),(\d+)\]$`)
	matches := re.FindStringSubmatch(rangeParam)
	if len(matches) != 3 {
		return 0, 0, fmt.Errorf("invalid range format")
	}
	start, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, err
	}
	end, err = strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, err
	}
	if start < 0 || end < start {
		return 0, 0, fmt.Errorf("invalid range values")
	}
	return start, end, nil
}
