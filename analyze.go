package main

import (
	"image"
	"math"
)

func getAverageImageBrightness(input image.Image, precision int) uint16 {
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

func getRequiredGamma(input image.Image, targetBrightness uint16, precision int) float64 {
	var sum float64
	var pixels uint64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, a := input.At(x, y).RGBA()
			if a > 0 {
				brightness := uint64(0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b))
				sum += math.Log(float64(brightness)/65536.0) / math.Log(float64(targetBrightness)/65536.0)
				pixels++
			}
		}
	}
	return sum / float64(pixels)
}

func getAverageImageKelvin(input image.Image, precision int) uint16 {
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
