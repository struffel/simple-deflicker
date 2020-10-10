package main

import (
	"image"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gosuri/uiprogress"
)

type picture struct {
	path             string
	brightness       uint16
	targetBrightness uint16
}

var pictures []picture

func main() {

	uiprogress.Start() // start rendering

	//Set number of CPU cores to use
	var maxThreads = runtime.NumCPU()
	runtime.GOMAXPROCS(maxThreads)

	//Get list of files
	var inputFolder = "./input/"
	files, err := ioutil.ReadDir(inputFolder)
	if err != nil {
		log.Fatal(err)
	}
	//Prepare array of pictures
	for _, file := range files {
		var extension = strings.ToLower(filepath.Ext(file.Name()))
		if extension == ".jpg" || extension == ".png" {
			pictures = append(pictures, picture{filepath.Join(inputFolder, file.Name()), 0, 0})
		}
	}
	progressBars := createProgressBars()
	//Prepare token channel
	tokens := make(chan bool, maxThreads)
	//Fill token channel with initial values and start the analysis loop
	for i := 0; i < maxThreads; i++ {
		tokens <- true
	}
	for i := range pictures {
		_ = <-tokens
		go func(i int) {
			defer func() {
				progressBars["ANALYZE"].Incr()
				tokens <- true
			}()
			var img, _ = imaging.Open(pictures[i].path)
			pictures[i].brightness = getAverageImageBrightness(img, 16)
		}(i)
	}
	for i := 0; i < maxThreads; i++ {
		_ = <-tokens
	}

	//Calculate global or rolling average
	var rollingAverageFrames = 0
	var targetBrightness uint64 = 0
	if rollingAverageFrames < 1 {
		for i := range pictures {
			targetBrightness += uint64(pictures[i].brightness)
		}
		targetBrightness /= uint64(len(pictures))
		for i := range pictures {
			pictures[i].targetBrightness = uint16(targetBrightness)
			progressBars["PREPARE"].Incr()
		}
	} else {
		for i := range pictures {
			targetBrightness = 0
			var start = maximum(i-rollingAverageFrames, 0)
			var end = minimum(i+rollingAverageFrames, len(pictures)-1)
			for j := start; j <= end; j++ {
				targetBrightness += uint64(pictures[j].brightness)
			}
			targetBrightness /= uint64(end - start + 1) //Throws odd results
			pictures[i].targetBrightness = uint16(targetBrightness)
			progressBars["PREPARE"].Incr()
		}
	}

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
				progressBars["ADJUST"].Incr()
				tokens <- true
			}()
			var img, _ = imaging.Open(pictures[i].path)
			var gamma = float64(pictures[i].targetBrightness) / float64(pictures[i].brightness)
			var imgCorrected = imaging.AdjustGamma(img, gamma)
			imaging.Save(imgCorrected, "./output/"+filepath.Base(pictures[i].path), imaging.JPEGQuality(95), imaging.PNGCompressionLevel(0))
		}(i)
	}
	for i := 0; i < maxThreads; i++ {
		_ = <-tokens
	}
	uiprogress.Stop()
}

func getAverageImageBrightness(input image.Image, precision int) uint16 {
	var sum, pixels uint64
	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y += precision {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x += precision {
			r, g, b, a := input.At(x, y).RGBA()
			if a > 0 {
				sum += uint64(0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b))
				pixels++
			}
		}
	}
	return uint16(sum / pixels)
}
