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
	var sum, images, average uint64
	fmt.Println("Starting...")

	images = 0
	sum = 0

	files, err := ioutil.ReadDir("./input/")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Print(".")
		var img, _ = imaging.Open("./input/" + f.Name())
		sum += getAverageLuminance(img)
		images++
	}
	fmt.Print("\n")

	average = sum / images
	fmt.Printf("AVG: %v\n", average)

	var current uint64 = 0

	for _, f := range files {
		var img, _ = imaging.Open("./input/" + f.Name())
		current = getAverageLuminance(img)
		var imgCorrected = imaging.AdjustGamma(img, float64(average)/float64(current))
		fmt.Print(".")
		imaging.Save(imgCorrected, "./output/"+filepath.Base(f.Name()))
	}
}

func getAverageLuminance(input image.Image) uint64 {
	var sum, pixels uint64
	var precision int
	precision = 4
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
	return sum / pixels
}
