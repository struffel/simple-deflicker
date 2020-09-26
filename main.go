package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/disintegration/imaging"
)

type picture struct {
	path       string
	brightness uint16
}

func main() {
	fmt.Println("Starting...")

	//Set number of CPU cores to use
	var maxThreads = runtime.NumCPU()
	runtime.GOMAXPROCS(maxThreads)

	var pictures []picture

	//Get list of files
	var inputFolder = "./input/"
	files, err := ioutil.ReadDir(inputFolder)
	if err != nil {
		log.Fatal(err)
	}
	for i, file := range files {
		var extension = strings.ToLower(filepath.Ext(file.Name()))
		if extension == ".jpg" || extension == ".png" {
			pictures = append(pictures, picture{filepath.Join(inputFolder, file.Name()), 0})
			fmt.Printf("%v\n", pictures[i].path)
		}
	}
	//var numberOfPictures=len(pictures)
	//fmt.Printf("Number of Pictures: %v\n", numberOfPictures)

	//Prepare array for brightness values and token channel
	fmt.Printf("Using %v threads...\n", maxThreads)
	tokens := make(chan bool, maxThreads)

	//Fill token channel with initial values and start the analysis loop
	for i := 0; i < maxThreads; i++ {
		tokens <- true
	}
	for i := range pictures {
		_ = <-tokens
		go func(i int) {
			defer func() {
				tokens <- true
			}()
			var img, _ = imaging.Open(pictures[i].path)
			pictures[i].brightness = getAverageImageBrightness(img, 16)
			fmt.Printf("%v | %v\n", pictures[i].path, pictures[i].brightness)
		}(i)

	}
	for i := 0; i < maxThreads; i++ {
		_ = <-tokens
	}
	fmt.Println("All threads finished!")

	//Calculate the average brightness
	var sum uint64 = 0
	for i := range pictures {
		sum += uint64(pictures[i].brightness)
	}
	var averageBrightness uint16 = uint16(float64(sum) / float64(len(pictures)))
	fmt.Printf("Average Brightness: %v\n", averageBrightness)

	//Create token channel and fill it with inital tokens
	tokens = make(chan bool, maxThreads)
	for i := 0; i < maxThreads; i++ {
		tokens <- true
	}

	//Run the loop for image adjustment
	for i := range pictures {
		_ = <-tokens
		go func(i int) {
			defer func() {
				tokens <- true
			}()
			var img, _ = imaging.Open(pictures[i].path)
			var gamma = float64(averageBrightness) / float64(pictures[i].brightness)
			var imgCorrected = imaging.AdjustGamma(img, gamma)
			imaging.Save(imgCorrected, "./output/"+filepath.Base(pictures[i].path), imaging.JPEGQuality(95), imaging.PNGCompressionLevel(0))
			fmt.Printf("%v | %v\n", pictures[i].path, gamma)
		}(i)
	}
	for i := 0; i < maxThreads; i++ {
		_ = <-tokens
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
