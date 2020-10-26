package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gosuri/uiprogress"
)

type lut [256]uint8
type histogram [256]uint32

type picture struct {
	path             string
	currentHistogram histogram
	targetHistogram  histogram
}

func main() {
	var pictures []picture

	var config struct {
		source         string
		destination    string
		rollingaverage int
		threads        int
	}

	flag.StringVar(&config.source, "source", ".", "Source folder")
	flag.StringVar(&config.destination, "destination", ".", "Destination folder")
	flag.IntVar(&config.rollingaverage, "rollingaverage", 10, "Number of frames to use for rolling average. 0 disables it.")
	flag.IntVar(&config.threads, "threads", runtime.NumCPU(), "Number of threads to use")
	flag.Parse()

	uiprogress.Start() // start rendering

	//Set number of CPU cores to use
	runtime.GOMAXPROCS(config.threads)

	//Get list of files
	files, err := ioutil.ReadDir(config.source)
	if err != nil {
		fmt.Printf("'%v': %v\n", config.source, err)
		os.Exit(1)
	}
	//Prepare slice of pictures
	for _, file := range files {
		var fullPath = filepath.Join(config.source, file.Name())
		var extension = strings.ToLower(filepath.Ext(file.Name()))
		var temp histogram
		if extension == ".jpg" || extension == ".png" {
			pictures = append(pictures, picture{fullPath, temp, temp})
		} else {
			fmt.Printf("'%v': ignoring file with unsupported extension\n", fullPath)
		}
	}
	progressBars := createProgressBars(len(pictures))

	//Analyze and create Histograms
	pictures = forEveryPicture(pictures, progressBars.analyze, config.threads, func(pic picture) picture {
		var img, err = imaging.Open(pic.path)
		if err != nil {
			fmt.Printf("'%v': %v\n", pic.path, err)
			os.Exit(2)
		}
		pic.currentHistogram = generateHistogramFromImage(img)
		return pic
	})

	//Calculate global or rolling average
	if config.rollingaverage < 1 {
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
			var start = clamp(i-config.rollingaverage, 0, len(pictures)-1)
			var end = clamp(i+config.rollingaverage, 0, len(pictures)-1)
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

	pictures = forEveryPicture(pictures, progressBars.analyze, config.threads, func(pic picture) picture {
		var img, _ = imaging.Open(pic.path)
		lut := generateLutFromHistograms(pic.currentHistogram, pic.targetHistogram)
		img = applyLutToImage(img, lut)
		imaging.Save(img, filepath.Join(config.destination, filepath.Base(pic.path)), imaging.JPEGQuality(95), imaging.PNGCompressionLevel(0))
		return pic
	})
	uiprogress.Stop()
}
