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

func clamp(a float64, min float64, max float64) float64 {
	if a < min {
		return min
	}
	if a > max {
		return max
	}
	return a
}
func formatHistogram(lut [256]int) string {
	output := ""
	for i, v := range lut {
		output += fmt.Sprintf("%v: %v\n", i, v)
	}
	return output
}
