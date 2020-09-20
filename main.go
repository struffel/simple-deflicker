package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/disintegration/imaging"
)

func main() {
	fmt.Println("Starting...")

	files, err := ioutil.ReadDir("./input/")
	if err != nil {
		log.Fatal(err)
	}

	brightnessValues := make([]uint16, len(files))
	fmt.Println(len(brightnessValues))

	for i, f := range files {
		fmt.Print(".")
		var img, _ = imaging.Open("./input/" + f.Name())
		brightnessValues[i] = getBrightness(img, 8)
	}
	fmt.Print("\n")

	var sum uint64 = 0
	for _, value := range brightnessValues {
		sum += uint64(value)
	}
	var average uint16 = uint16(float64(sum) / float64(len(brightnessValues)))
	fmt.Printf("AVG: %v\n", average)
	for i, f := range files {
		var img, _ = imaging.Open("./input/" + f.Name())
		var imgCorrected = imaging.AdjustGamma(img, float64(average)/float64(brightnessValues[i]))
		fmt.Print(".")
		imaging.Save(imgCorrected, "./output/"+filepath.Base(f.Name()), imaging.JPEGQuality(95), imaging.PNGCompressionLevel(0))
	}
}

func getBrightness(input image.Image, precision int) uint16 {
	var sum, pixels uint64
	input = imaging.Grayscale(input)
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			col, _, _, alpha := input.At(x, y).RGBA()
			if alpha > 0 {
				sum += uint64(col)
				pixels++
			}
		}
	}
	return uint16(sum / pixels)
}
