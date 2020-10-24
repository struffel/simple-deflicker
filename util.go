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

func printDebug() {
	fmt.Printf("%-40v%-20v%-20v%-20v%-20v%-20v\n", "Path", "CurrentBrightness", "TargetBrightness", "RequiredGamma", "RequiredConstrast", "RequiredIntensity")
	for _, pic := range pictures {
		fmt.Printf("%-40v%-20v%-20v%-20v%-20v%-20v\n", pic.path, pic.currentIntensity, pic.targetIntensity, pic.requiredGammaChange, pic.requiredContrastChange, pic.requiredIntensityChange)
	}
}
