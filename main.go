package main

import (
	"flag"
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
	currentHistogram [256]int
	targetHistogram  [256]int
}

var pictures []picture

var config struct {
	source         string
	destination    string
	rollingaverage int
	threads        int
}

func main() {

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
		log.Fatalf("'%v': %v", config.source, err)
	}
	//Prepare array of pictures
	for _, file := range files {
		var fullPath = filepath.Join(config.source, file.Name())
		var extension = strings.ToLower(filepath.Ext(file.Name()))
		var temp [256]int
		if extension == ".jpg" || extension == ".png" {
			pictures = append(pictures, picture{fullPath, temp, temp})
		} else {
			log.Printf("'%v': ignoring file with unsupported extension", fullPath)
		}
	}
	progressBars := createProgressBars()

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
				log.Fatalf("'%v': %v", pictures[i].path, err)
			}
			pictures[i].currentHistogram = calculateHistogram(img)
			//pictures[i].kelvin = getAverageImageKelvin(img, 8)
		}(i)
	}
	for i := 0; i < config.threads; i++ {
		_ = <-tokens
	}
	//Calculate global or rolling average
	if config.rollingaverage < 1 {
		var averageHistogram [256]int
		for i := range pictures {
			for j := 0; j < 256; j++ {
				averageHistogram[j] += pictures[i].currentHistogram[j]
			}
		}
		for i := 0; i < 256; i++ {
			averageHistogram[i] /= len(pictures)
		}
		//fmt.Println(formatHistogram(averageHistogram))
		for i := range pictures {
			pictures[i].targetHistogram = averageHistogram
		}
	} else {
		for i := range pictures {
			var averageHistogram [256]int
			var start = maximum(0, i-config.rollingaverage)
			var end = minimum(len(pictures)-1, i+config.rollingaverage)
			for i := start; i <= end; i++ {
				for j := 0; j < 256; j++ {
					averageHistogram[j] += pictures[i].currentHistogram[j]
				}
			}
			for i := 0; i < 256; i++ {
				averageHistogram[i] /= end - start + 1
			}
			pictures[i].targetHistogram = averageHistogram
		}
	}

	//Create token channel and fill it with inital tokens
	tokens = make(chan bool, config.threads)
	for i := 0; i < config.threads; i++ {
		tokens <- true
	}
	//printDebug()
	//Run the loop for image adjustment
	for i := range pictures {
		_ = <-tokens
		go func(i int) {
			defer func() {
				progressBars["ADJUST"].Incr()
				tokens <- true
			}()
			//fmt.Println(pictures[i].path)
			var img, _ = imaging.Open(pictures[i].path)
			lut := generateLutFromHistograms(pictures[i].currentHistogram, pictures[i].targetHistogram)
			//fmt.Println("LUT\n" + formatHistogram(lut))
			img = applyLut(img, lut)
			imaging.Save(img, filepath.Join(config.destination, filepath.Base(pictures[i].path)), imaging.JPEGQuality(95), imaging.PNGCompressionLevel(0))
		}(i)
	}
	for i := 0; i < config.threads; i++ {
		_ = <-tokens
	}
	uiprogress.Stop()
	//printDebug()
}
