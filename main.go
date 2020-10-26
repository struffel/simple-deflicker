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

var config struct {
	source         string
	destination    string
	rollingaverage int
	threads        int
}

func main() {
	var pictures []picture

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

	//Prepare token channel
	tokens := make(chan bool, config.threads)
	//Fill token channel with initial values and start the analysis loop
	for i := 0; i < config.threads; i++ {
		tokens <- true
	}
	for i := range pictures {
		_ = <-tokens
		go func(i int) {
			defer func() {
				progressBars["INITIALIZE"].Incr()
				tokens <- true
			}()
			var img, err = imaging.Open(pictures[i].path)
			if err != nil {
				fmt.Printf("'%v': %v\n", pictures[i].path, err)
				os.Exit(2)
			}
			pictures[i].currentHistogram = generateHistogramFromImage(img)
			//pictures[i].kelvin = getAverageImageKelvin(img, 8)
		}(i)
	}
	for i := 0; i < config.threads; i++ {
		_ = <-tokens
	}
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
			var end = clamp(i-config.rollingaverage, 0, len(pictures)-1)
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

	//Create token channel and fill it with inital tokens
	tokens = make(chan bool, config.threads)
	for i := 0; i < config.threads; i++ {
		tokens <- true
	}
	//Run the loop for image adjustment
	for i := range pictures {
		_ = <-tokens
		go func(i int) {
			defer func() {
				progressBars["ADJUST"].Incr()
				tokens <- true
			}()
			var img, _ = imaging.Open(pictures[i].path)
			lut := generateLutFromHistograms(pictures[i].currentHistogram, pictures[i].targetHistogram)
			img = applyLutToImage(img, lut)
			imaging.Save(img, filepath.Join(config.destination, filepath.Base(pictures[i].path)), imaging.JPEGQuality(95), imaging.PNGCompressionLevel(0))
		}(i)
	}
	for i := 0; i < config.threads; i++ {
		_ = <-tokens
	}
	uiprogress.Stop()
}
