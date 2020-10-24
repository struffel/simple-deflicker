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
	fmt.Printf("%-40v%-25v%-25v%-25v%-25v%-25v%-25v%-25v\n", "Path", "CurrentIntensity", "TargetIntensity", "CurrentContrast", "TargetContrast", "RequiredGammaChange", "RequiredConstrastChange", "RequiredIntensityChange")
	for _, pic := range pictures {
		fmt.Printf("%-40v%-25v%-25v%-25v%-25v%-25v%-25v%-25v\n", pic.path, pic.currentIntensity, pic.targetIntensity, pic.currentContrast, pic.targetContrast, pic.requiredGammaChange, pic.requiredContrastChange, pic.requiredIntensityChange)
	}
}
