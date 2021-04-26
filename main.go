package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/disintegration/imaging"
)

type picture struct {
	currentPath         string
	targetPath          string
	currentRgbHistogram rgbHistogram
	targetRgbHistogram  rgbHistogram
}

func main() {
	//Initial console output
	printInfo()
	//Read parameters from console
	config = collectConfigInformation()
	//Initialize Window from config and start GUI
	initalizeWindow()
	window.Main()
	os.Exit(0)
}

func runDeflickering() error {

	//Prepare
	configError := validateConfigInformation()
	if configError != nil {
		return configError
	}
	clear()
	runtime.GOMAXPROCS(config.threads)
	pictures, picturesError := readDirectory(config.sourceDirectory, config.destinationDirectory)
	if picturesError != nil {
		return picturesError
	}
	progress := createProgressBars(len(pictures))
	progress.container.Start()

	//Analyze and create Histograms
	var analyzeError error
	pictures, analyzeError = forEveryPicture(pictures, progress.bars["analyze"], config.threads, func(pic picture) (picture, error) {
		img, err := imaging.Open(pic.currentPath)
		if err != nil {
			return pic, errors.New(pic.currentPath + " | " + err.Error())
		}
		pic.currentRgbHistogram = generateRgbHistogramFromImage(img)
		return pic, nil
	})
	if analyzeError != nil {
		progress.container.Stop()
		return analyzeError
	}

	pictures = addTargetHistograms(pictures)

	var adjustError error
	pictures, adjustError = forEveryPicture(pictures, progress.bars["adjust"], config.threads, func(pic picture) (picture, error) {
		var img, _ = imaging.Open(pic.currentPath)
		lut := generateRgbLutFromRgbHistograms(pic.currentRgbHistogram, pic.targetRgbHistogram)
		img = applyRgbLutToImage(img, lut)
		err := imaging.Save(img, pic.targetPath, imaging.JPEGQuality(config.jpegCompression), imaging.PNGCompressionLevel(0))
		if err != nil {
			return pic, errors.New(pic.currentPath + " | " + err.Error())
		}
		return pic, nil
	})
	if adjustError != nil {
		progress.container.Stop()
		return adjustError
	}
	progress.container.Stop()
	clear()
	fmt.Printf("Saved %v pictures into %v", len(pictures), config.destinationDirectory)
	return nil
}
