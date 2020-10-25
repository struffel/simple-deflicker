package main

import (
	"image"
	"math"
)

func measureIntensity(input image.Image, precision int) float64 {
	var sum float64
	var pixels uint64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, a := input.At(x, y).RGBA()
			if a > 0 {
				sum += 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
				pixels++
			}
		}
	}
	return sum / float64(pixels)
}

func measureContrast(input image.Image, averageIntensity float64, precision int) float64 {
	var sum float64
	var pixels uint64
	var averageIntensityNormalized float64 = averageIntensity / 65536.0
	var intensityNormalized float64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, a := input.At(x, y).RGBA()
			if a > 0 {
				intensityNormalized = (0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)) / 65536.0
				sum += math.Pow((intensityNormalized-averageIntensityNormalized)/averageIntensityNormalized, 2.0)
				pixels++
			}
		}
	}
	return math.Sqrt(sum / float64(pixels))
}

func measureKelvin(input image.Image, precision int) float64 {
	var sum float64
	var pixels uint64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, a := input.At(x, y).RGBA()
			if a > 0 {
				n := (0.23881*float64(r) + 0.25499*float64(g) - 0.58291*float64(b)) / (0.11109*float64(r) - 0.85406*float64(g) + 0.52289*float64(b))
				k := 449*(n*n*n) + 3525*(n*n) + 6823.3*n + 5520.33
				sum += k
				pixels++
			}
		}
	}
	return sum / float64(pixels)
}

func calculateIntensityDifference(input image.Image, targetIntensity float64, precision int) float64 {
	var sum float64
	var pixels uint64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, a := input.At(x, y).RGBA()
			if a > 0 {
				intensity := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
				sum += targetIntensity - intensity
				pixels++
			}
		}
	}
	return sum / float64(pixels)
}

func calculateGammaDifference(input image.Image, targetIntensity float64, precision int) float64 {
	var sum float64
	var pixels uint64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, a := input.At(x, y).RGBA()
			if a > 0 {
				intensity := float64(0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b))
				//if intensity > 0.001*65536.0 && intensity < 0.999*65536.0 {
				gamma := math.Log(clamp(intensity, 1.0, 65535.0)/65536.0) / math.Log(clamp(targetIntensity, 1.0, 65535.0)/65536.0)
				sum += gamma
				pixels++
				//fmt.Printf("G: %v\n", gamma)
				//}
			}
		}
	}
	return sum / float64(pixels)
}
func calculateSimpleGammaDifference(currentIntensity float64, targetIntensity float64) float64 {
	return math.Log(clamp(float64(currentIntensity), 1.0, 65535.0)/65536.0) / math.Log(clamp(float64(targetIntensity), 1.0, 65535.0)/65536.0)
}

func calculateHistogram(input image.Image, precision int) [256]int {
	var histogram [256]int
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, a := input.At(x, y).RGBA()
			if a > 0 {
				intensity := float64(0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b))
				histogram[uint8(intensity)]++
			}
		}
	}
	return histogram
}
