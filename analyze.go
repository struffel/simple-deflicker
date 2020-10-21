package main

import (
	"image"
	"math"
)

func measureIntensity(input image.Image, precision int) uint16 {
	var sum, pixels uint64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, a := input.At(x, y).RGBA()
			if a > 0 {
				sum += uint64(0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b))
				pixels++
			}
		}
	}
	return uint16(sum / pixels)
}

func measureContrast(input image.Image, averageIntensity uint16, precision int) float64 {
	var sum float64
	var pixels uint64
	var averageIntensityNormalized = float64(averageIntensity) / 65536.0
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, a := input.At(x, y).RGBA()
			if a > 0 {
				intensityNormalized := float64((0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b))) / 65536.0
				sum += math.Pow((intensityNormalized-averageIntensityNormalized)/averageIntensityNormalized, 2.0)
				pixels++
			}
		}
	}
	return math.Sqrt(sum / float64(pixels))
}

func measureKelvin(input image.Image, precision int) uint16 {
	var sum, pixels uint64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, a := input.At(x, y).RGBA()
			if a > 0 {
				n := (0.23881*float32(r) + 0.25499*float32(g) - 0.58291*float32(b)) / (0.11109*float32(r) - 0.85406*float32(g) + 0.52289*float32(b))
				k := 449*(n*n*n) + 3525*(n*n) + 6823.3*n + 5520.33
				sum += uint64(k)
				pixels++
			}
		}
	}
	return uint16(sum / pixels)
}

func calculateIntensityDifference(input image.Image, targetIntensity uint16, precision int) float64 {
	var sum int64
	var pixels uint64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, a := input.At(x, y).RGBA()
			if a > 0 {
				intensity := uint64(0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b))
				if intensity != 0 && intensity != 65536 {
					sum += int64(targetIntensity) - int64(intensity)
					pixels++
				}
			}
		}
	}
	return float64(sum) / float64(pixels)
}

func calculateGammaDifference(input image.Image, targetIntensity uint16, precision int) float64 {
	var sum float64
	var pixels uint64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, a := input.At(x, y).RGBA()
			if a > 0 {
				intensity := uint64(0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b))
				if intensity != 0 && intensity != 65536 {
					sum += math.Log(clamp(float64(intensity), 1.0, 65535.0)/65536.0) / math.Log(clamp(float64(targetIntensity), 1.0, 65535.0)/65536.0)
					pixels++
				}
			}
		}
	}
	return sum / float64(pixels)
}
