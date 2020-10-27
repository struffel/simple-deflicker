package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/skratchdot/open-golang/open"

	"github.com/disintegration/imaging"
	"github.com/gosuri/uiprogress"
)

type lut [256]uint8
type histogram [256]uint32

type picture struct {
	currentPath      string
	targetPath       string
	currentHistogram histogram
	targetHistogram  histogram
}
type config struct {
	sourceDirectory      string
	destinationDirectory string
	rollingaverage       int
	threads              int
}

func main() {

	config := collectConfigInformation()

	makeDirectoryIfNotExists(config.destinationDirectory)

	//Set number of CPU cores to use
	runtime.GOMAXPROCS(config.threads)

	pictures := createPictureSliceFromDirectory(config.sourceDirectory, config.destinationDirectory)
	runDeflickering(pictures, config.rollingaverage, config.threads)
	open.Start(config.destinationDirectory)
	fmt.Println("Finished. This window will close itself in 5 seconds")
	time.Sleep(time.Second * 5)
	os.Exit(0)
}

func createPictureSliceFromDirectory(currentDirectory string, targetDirectory string) []picture {
	var pictures []picture
	//Get list of files
	files, err := ioutil.ReadDir(currentDirectory)
	if err != nil {
		fmt.Printf("'%v': %v\n", currentDirectory, err)
		os.Exit(1)
	}
	//Prepare slice of pictures
	for _, file := range files {
		var fullSourcePath = filepath.Join(currentDirectory, file.Name())
		var fullTargetPath = filepath.Join(targetDirectory, file.Name())
		var extension = strings.ToLower(filepath.Ext(file.Name()))
		var temp histogram
		if extension == ".jpg" || extension == ".png" {
			pictures = append(pictures, picture{fullSourcePath, fullTargetPath, temp, temp})
		} else {
			fmt.Printf("'%v': ignoring file with unsupported extension\n", fullSourcePath)
		}
	}
	return pictures
}

func runDeflickering(pictures []picture, rollingaverage int, threads int) {
	uiprogress.Start() // start rendering
	progressBars := createProgressBars(len(pictures))

	//Analyze and create Histograms
	pictures = forEveryPicture(pictures, progressBars.analyze, threads, func(pic picture) picture {
		var img, err = imaging.Open(pic.currentPath)
		if err != nil {
			fmt.Printf("'%v': %v\n", pic.targetPath, err)
			os.Exit(2)
		}
		pic.currentHistogram = generateHistogramFromImage(img)
		return pic
	})

	//Calculate global or rolling average
	if rollingaverage < 1 {
		var averageHistogram histogram
		for i := range pictures {
			for j := 0; j < 256; j++ {
				averageHistogram[j] += pictures[i].currentHistogram[j]
			}
		}
		for i := 0; i < 256; i++ {
			averageHistogram[i] /= uint32(len(pictures))
		}
		for i := range pictures {
			pictures[i].targetHistogram = averageHistogram
		}
	} else {
		for i := range pictures {
			var averageHistogram histogram
			var start = clamp(i-rollingaverage, 0, len(pictures)-1)
			var end = clamp(i+rollingaverage, 0, len(pictures)-1)
			for i := start; i <= end; i++ {
				for j := 0; j < 256; j++ {
					averageHistogram[j] += pictures[i].currentHistogram[j]
				}
			}
			for i := 0; i < 256; i++ {
				averageHistogram[i] /= uint32(end - start + 1)
			}
			pictures[i].targetHistogram = averageHistogram
		}
	}

	pictures = forEveryPicture(pictures, progressBars.adjust, threads, func(pic picture) picture {
		var img, _ = imaging.Open(pic.currentPath)
		lut := generateLutFromHistograms(pic.currentHistogram, pic.targetHistogram)
		img = applyLutToImage(img, lut)
		imaging.Save(img, pic.targetPath, imaging.JPEGQuality(95), imaging.PNGCompressionLevel(0))
		return pic
	})
	uiprogress.Stop()
}
