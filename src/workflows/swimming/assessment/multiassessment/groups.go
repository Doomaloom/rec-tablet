package multiassessment

import (
	"fmt"
)

func BalancedSizes(students, groups int) []int {
	if students < 1 || groups < 1 || groups > students {
		return nil
	}
	sizes := make([]int, groups)
	for i := range sizes {
		sizes[i] = students / groups
		if i < students%groups {
			sizes[i]++
		}
	}
	return sizes
}

func SplitLabel(students, groups int) string {
	sizes := BalancedSizes(students, groups)
	if len(sizes) == 0 {
		return ""
	}
	groupLabel := func(n, size int) string {
		noun := "groups"
		if n == 1 {
			noun = "group"
		}
		return fmt.Sprintf("%d %s of %d", n, noun, size)
	}
	larger := students % groups
	if larger == 0 {
		return groupLabel(groups, sizes[0])
	}
	return groupLabel(larger, sizes[0]) + ", " + groupLabel(groups-larger, sizes[len(sizes)-1])
}
