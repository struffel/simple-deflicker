package deflicker

import (
	"image"
	"image/color"
	"math"

	"github.com/disintegration/imaging"
)

type lut [256]uint8
type rgbLut struct {
	R lut
	G lut
	B lut
}
type histogram [256]uint32
type rgbHistogram struct {
	R histogram
	G histogram
	B histogram
}

const (
	lutCorrectionStrength = 0.8
	lutSmoothingRadius    = 5
)

// generateRgbHistogramFromImage generates an RGB histogram from the given image.
func generateRgbHistogramFromImage(input image.Image) rgbHistogram {
	var rgbHistogram rgbHistogram
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y++ {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x++ {
			r, g, b, _ := input.At(x, y).RGBA()
			r = r >> 8
			g = g >> 8
			b = b >> 8
			rgbHistogram.R[r]++
			rgbHistogram.G[g]++
			rgbHistogram.B[b]++
		}
	}
	return rgbHistogram
}

// convertToCumulativeRgbHistogram converts a given RGB histogram into a cumulative histogram.
func convertToCumulativeRgbHistogram(input rgbHistogram) rgbHistogram {
	var targetRgbHistogram rgbHistogram
	targetRgbHistogram.R[0] = input.R[0]
	targetRgbHistogram.G[0] = input.G[0]
	targetRgbHistogram.B[0] = input.B[0]
	for i := 1; i < 256; i++ {
		targetRgbHistogram.R[i] = targetRgbHistogram.R[i-1] + input.R[i]
		targetRgbHistogram.G[i] = targetRgbHistogram.G[i-1] + input.G[i]
		targetRgbHistogram.B[i] = targetRgbHistogram.B[i-1] + input.B[i]
	}
	return targetRgbHistogram
}

// generateRgbLutFromRgbHistograms generates a lookup table (LUT) for each color channel (R, G, B) based on the current and target RGB histograms.
// The LUT is used to map pixel values from the current image to the target image, effectively adjusting the colors to match the desired histogram.
func generateRgbLutFromRgbHistograms(current rgbHistogram, target rgbHistogram) rgbLut {
	currentCumulativeRgbHistogram := convertToCumulativeRgbHistogram(current)
	targetCumulativeRgbHistogram := convertToCumulativeRgbHistogram(target)

	var ratio [3]float64
	ratio[0] = float64(currentCumulativeRgbHistogram.R[255]) / float64(targetCumulativeRgbHistogram.R[255])
	ratio[1] = float64(currentCumulativeRgbHistogram.G[255]) / float64(targetCumulativeRgbHistogram.G[255])
	ratio[2] = float64(currentCumulativeRgbHistogram.B[255]) / float64(targetCumulativeRgbHistogram.B[255])
	for i := 0; i < 256; i++ {
		targetCumulativeRgbHistogram.R[i] = uint32(0.5 + float64(targetCumulativeRgbHistogram.R[i])*ratio[0])
		targetCumulativeRgbHistogram.G[i] = uint32(0.5 + float64(targetCumulativeRgbHistogram.G[i])*ratio[1])
		targetCumulativeRgbHistogram.B[i] = uint32(0.5 + float64(targetCumulativeRgbHistogram.B[i])*ratio[2])
	}

	//Generate LUT
	var lut rgbLut
	var p [3]uint8
	for i := 0; i < 256; i++ {
		for targetCumulativeRgbHistogram.R[p[0]] < currentCumulativeRgbHistogram.R[i] {
			p[0]++
		}
		for targetCumulativeRgbHistogram.G[p[1]] < currentCumulativeRgbHistogram.G[i] {
			p[1]++
		}
		for targetCumulativeRgbHistogram.B[p[2]] < currentCumulativeRgbHistogram.B[i] {
			p[2]++
		}
		lut.R[i] = p[0]
		lut.G[i] = p[1]
		lut.B[i] = p[2]
	}
	lut.R = regularizeLut(lut.R)
	lut.G = regularizeLut(lut.G)
	lut.B = regularizeLut(lut.B)
	return lut
}

// applyRgbLutToImage applies the given RGB lookup table (LUT) to the input image, adjusting its pixel values according to the LUT.
func applyRgbLutToImage(input image.Image, lut rgbLut) image.Image {
	result := imaging.AdjustFunc(input, func(c color.NRGBA) color.NRGBA {
		c.R = uint8(lut.R[c.R])
		c.G = uint8(lut.G[c.G])
		c.B = uint8(lut.B[c.B])
		return c
	})
	return result
}

// regularizeLut smooths and corrects the given lookup table (LUT) to ensure that it is monotonically increasing and applies a correction strength to the values.
// This helps in preventing abrupt changes in pixel values when applying the LUT to an image.
func regularizeLut(input lut) lut {
	var smoothed lut
	for i := 0; i < 256; i++ {
		sum := 0
		count := 0
		for j := i - lutSmoothingRadius; j <= i+lutSmoothingRadius; j++ {
			if j < 0 || j > 255 {
				continue
			}
			sum += int(input[j])
			count++
		}
		smoothed[i] = uint8(sum / count)
	}
	for i := 1; i < 256; i++ {
		if smoothed[i] < smoothed[i-1] {
			smoothed[i] = smoothed[i-1]
		}
	}

	var result lut
	for i := 0; i < 256; i++ {
		corrected := float64(smoothed[i])*lutCorrectionStrength + float64(i)*(1-lutCorrectionStrength)
		result[i] = clampUint8(corrected)
	}
	return result
}

// clampUint8 clamps a float64 value to the range of 0 to 255 and returns it as a uint8.
func clampUint8(value float64) uint8 {
	value = math.Round(value)
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return uint8(value)
}
