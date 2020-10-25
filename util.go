package main

import "fmt"

func minimum(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maximum(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(a int, min int, max int) int {
	if a < min {
		return min
	}
	if a > max {
		return max
	}
	return a
}
func formatHistogram(lut [256]uint8) string {
	output := ""
	for i, v := range lut {
		output += fmt.Sprintf("%v: %v\n", i, v)
	}
	return output
}
