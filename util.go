package main

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
