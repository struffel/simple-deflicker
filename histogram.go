package main

import (
	"image"
	"image/color"

	"github.com/disintegration/imaging"
)

func generateLuminanceHistogramFromImage(input image.Image) histogram {
	var histogram histogram
	var pixels uint64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y++ {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x++ {
			r, g, b, _ := input.At(x, y).RGBA()
			intensity := float64(0.2126*float64(r)+0.7152*float64(g)+0.0722*float64(b)) / 256.0
			histogram[int(intensity)]++
			pixels++
		}
	}
	return histogram
}

func generateRGBHistogramFromImage(input image.Image) rgbHistogram {
	var histogram rgbHistogram
	var pixels uint64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y++ {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x++ {
			r, g, b, _ := input.At(x, y).RGBA()
			histogram.r[r/256]++
			histogram.g[g/256]++
			histogram.b[b/256]++
			pixels++
		}
	}
	return histogram
}

func convertToCumulativeHistogram(input histogram) histogram {
	var targetHistogram histogram
	targetHistogram[0] = input[0]
	for i := 1; i < 256; i++ {
		targetHistogram[i] = targetHistogram[i-1] + input[i]
	}
	return targetHistogram
}

func convertToCumulativeRGBHistogram(input rgbHistogram) rgbHistogram {
	var targetHistogram rgbHistogram
	targetHistogram.r[0] = input.r[0]
	targetHistogram.g[0] = input.g[0]
	targetHistogram.b[0] = input.b[0]
	for i := 1; i < 256; i++ {
		targetHistogram.r[i] = targetHistogram.r[i-1] + input.r[i]
		targetHistogram.g[i] = targetHistogram.g[i-1] + input.g[i]
		targetHistogram.b[i] = targetHistogram.b[i-1] + input.b[i]
	}
	return targetHistogram
}

func generateLutFromHistograms(current histogram, target histogram) lut {
	currentCumulativeHistogram := convertToCumulativeHistogram(current)
	targetCumulativeHistogram := convertToCumulativeHistogram(target)

	ratio := float64(currentCumulativeHistogram[255]) / float64(targetCumulativeHistogram[255])
	for i := 0; i < 256; i++ {
		targetCumulativeHistogram[i] = uint32(0.5 + float64(targetCumulativeHistogram[i])*ratio)
	}

	//Generate LUT
	var lut lut
	var p uint8 = 0
	for i := 0; i < 256; i++ {
		for targetCumulativeHistogram[p] < currentCumulativeHistogram[i] {
			p++
		}
		lut[i] = p
	}
	return lut
}

func generateRGBLutFromHistograms(current rgbHistogram, target rgbHistogram) rgbLut {
	currentCumulativeHistogram := convertToCumulativeRGBHistogram(current)
	targetCumulativeHistogram := convertToCumulativeRGBHistogram(target)
	var ratio [3]float64
	ratio[0] = float64(currentCumulativeHistogram.r[255]) / float64(targetCumulativeHistogram.r[255])
	ratio[1] = float64(currentCumulativeHistogram.g[255]) / float64(targetCumulativeHistogram.g[255])
	ratio[2] = float64(currentCumulativeHistogram.b[255]) / float64(targetCumulativeHistogram.b[255])
	for i := 0; i < 256; i++ {
		targetCumulativeHistogram.r[i] = uint32(0.5 + float64(targetCumulativeHistogram.r[i])*ratio[0])
		targetCumulativeHistogram.g[i] = uint32(0.5 + float64(targetCumulativeHistogram.g[i])*ratio[1])
		targetCumulativeHistogram.b[i] = uint32(0.5 + float64(targetCumulativeHistogram.b[i])*ratio[2])
	}

	//Generate LUT
	var lut rgbLut
	var p [3]uint8
	for i := 0; i < 256; i++ {
		for targetCumulativeHistogram.r[p[0]] < currentCumulativeHistogram.r[i] {
			p[0]++
		}
		for targetCumulativeHistogram.g[p[1]] < currentCumulativeHistogram.g[i] {
			p[1]++
		}
		for targetCumulativeHistogram.r[p[2]] < currentCumulativeHistogram.b[i] {
			p[2]++
		}
		lut.r[i] = p[0]
		lut.g[i] = p[1]
		lut.b[i] = p[2]
	}
	return lut
}

func applyLutToImage(input image.Image, lut lut) image.Image {
	result := imaging.AdjustFunc(input, func(c color.NRGBA) color.NRGBA {
		c.R = uint8(lut[c.R])
		c.G = uint8(lut[c.G])
		c.B = uint8(lut[c.B])
		return c
	})
	return result
}

func applyRGBLutToImage(input image.Image, lut rgbLut) image.Image {
	result := imaging.AdjustFunc(input, func(c color.NRGBA) color.NRGBA {
		c.R = uint8(lut.r[c.R])
		c.G = uint8(lut.g[c.G])
		c.B = uint8(lut.b[c.B])
		return c
	})
	return result
}
