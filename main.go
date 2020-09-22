package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"
	"runtime"
	"time"

	"github.com/StruffelProductions/imaging"
)

func main() {
	fmt.Println("Starting...")
	var maxThreads = runtime.NumCPU()

	runtime.GOMAXPROCS(maxThreads)

	files, err := ioutil.ReadDir("./input/")
	if err != nil {
		log.Fatal(err)
	}

	var numberOfFiles = len(files)
	fmt.Printf("Number of Files: %v\n", numberOfFiles)

	brightnessValues := make([]uint16, numberOfFiles)
	hueValues := make([]float64, numberOfFiles)
	saturationValues := make([]float64, numberOfFiles)

	var ongoingThreads = 0
	fmt.Printf("Using %v threads...\n", maxThreads)
	for i, file := range files {
		for ongoingThreads >= maxThreads {
			time.Sleep(time.Millisecond * 5)
		}
		ongoingThreads++
		go func(i int, fileName string) {
			defer func() {
				ongoingThreads--
			}()
			if filepath.Ext(fileName) == ".JPG" {
				var img, _ = imaging.Open(fileName)
				brightnessValues[i] = getAverageImageBrightness(img, 8)
				hueValues[i] = getAverageImageHue(img, 8)
				saturationValues[i] = getAverageImageSaturation(img, 8)
				fmt.Printf("%v | %v | %v | %v\n", fileName, brightnessValues[i], hueValues[i], saturationValues[i])
			}
		}(i, "./input/"+file.Name())
	}
	for ongoingThreads > 0 {
		time.Sleep(time.Millisecond * 5)
	}
	fmt.Println("All threads finished!")

	fmt.Print("\n")

	var sumBrightness uint64 = 0
	var sumHue float64 = 0
	var sumSaturation float64 = 0
	//for _, value := range brightnessValues {
	for i := 0; i < numberOfFiles; i++ {
		sumBrightness += uint64(brightnessValues[i])
		sumHue += hueValues[i]
		sumSaturation += saturationValues[i]
	}

	var averageBrightness uint16 = uint16(float64(sumBrightness) / float64(numberOfFiles))
	var averageHue float64 = sumHue / float64(numberOfFiles)
	var averageSaturation float64 = sumSaturation / float64(numberOfFiles)

	fmt.Printf("AVG Brightness: %v\n", averageBrightness)
	fmt.Printf("AVG Hue: %v\n", averageHue)

	for i, file := range files {
		for ongoingThreads >= maxThreads {
			time.Sleep(time.Millisecond * 5)
		}
		ongoingThreads++
		go func(i int, fileName string) {
			defer func() {
				ongoingThreads--
			}()
			if filepath.Ext(fileName) == ".JPG" {
				var img, _ = imaging.Open(fileName)
				var gamma = float64(averageBrightness) / float64(brightnessValues[i])
				var hueShift = -(float64(hueValues[i]) - averageHue) * 360 * 0
				var saturationPercentage = (averageSaturation / saturationValues[i]) * 100.0
				var imgCorrected = imaging.AdjustGamma(img, gamma)
				imgCorrected = imaging.AdjustHues(imgCorrected, hueShift)
				imgCorrected = imaging.AdjustSaturation(imgCorrected, saturationPercentage)
				imaging.Save(imgCorrected, "./output/"+filepath.Base(fileName), imaging.JPEGQuality(95), imaging.PNGCompressionLevel(0))
				fmt.Printf("%v | %v | %v | %v\n", fileName, gamma, hueShift, saturationPercentage)
			}
		}(i, "./input/"+file.Name())
	}
	for ongoingThreads > 0 {
		time.Sleep(time.Millisecond * 5)
	}
	fmt.Println("All threads finished!")
}

func getAverageImageBrightness(input image.Image, precision int) uint16 {
	var sum, pixels uint64
	var inputGray = imaging.Grayscale(input)
	for y := inputGray.Bounds().Min.Y; y < inputGray.Bounds().Max.Y; y += precision {
		for x := inputGray.Bounds().Min.X; x < inputGray.Bounds().Max.X; x += precision {
			col, _, _, alpha := inputGray.At(x, y).RGBA()
			if alpha > 0 {
				sum += uint64(col)
				pixels++
			}
		}
	}
	return uint16(sum / pixels)
}

func getAverageImageHue(input image.Image, precision int) float64 {
	var pixels uint64
	var sum float64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, alpha := input.At(x, y).RGBA()
			if alpha > 0 {
				rr := float64(r) / 255
				gg := float64(g) / 255
				bb := float64(b) / 255

				max := math.Max(rr, math.Max(gg, bb))
				min := math.Min(rr, math.Min(gg, bb))

				var h float64
				d := max - min
				//fmt.Println(max)
				switch max {
				case rr:
					h = (gg - bb) / d
					if g < b {
						h += 6
					}
				case gg:
					h = (bb-rr)/d + 2
				case bb:
					h = (rr-gg)/d + 4
				}
				h /= 6
				if max == min {
					h = 0
				}
				sum += h
				pixels++
			}

		}
	}
	return sum / float64(pixels)
}

func getAverageImageSaturation(input image.Image, precision int) float64 {
	var pixels uint64
	var sum float64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, alpha := input.At(x, y).RGBA()
			if alpha > 0 {
				rr := float64(r) / 255
				gg := float64(g) / 255
				bb := float64(b) / 255

				max := math.Max(rr, math.Max(gg, bb))
				min := math.Min(rr, math.Min(gg, bb))

				l := (max + min) / 2

				d := max - min

				var s float64
				if l > 0.5 {
					s = d / (2 - max - min)
				} else if l > 0 {
					s = d / (max + min)
				} else {
					s = 0
				}
				sum += s
				pixels++
			}

		}
	}
	return sum / float64(pixels)
}
