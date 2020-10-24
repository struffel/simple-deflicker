package main

import (
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/disintegration/imaging"
)

type picture struct {
	path                    string
	currentIntensity        float64
	currentContrast         float64
	targetIntensity         float64
	targetContrast          float64
	requiredGammaChange     float64
	requiredIntensityChange float64
	requiredContrastChange  float64
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

	//uiprogress.Start() // start rendering

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
		if extension == ".jpg" || extension == ".png" {
			pictures = append(pictures, picture{fullPath, 0, 0, 0, 0, 0, 0, 0})
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
			pictures[i].currentIntensity = measureIntensity(img, 16)
			pictures[i].currentContrast = measureContrast(img, pictures[i].currentIntensity, 16)
			//pictures[i].kelvin = getAverageImageKelvin(img, 8)
		}(i)
	}
	for i := 0; i < config.threads; i++ {
		_ = <-tokens
	}

	//Calculate global or rolling average
	var targetIntensity float64 = 0
	var targetContrast float64 = 0
	if config.rollingaverage < 1 {
		for i := range pictures {
			targetIntensity += pictures[i].currentIntensity
			targetContrast += pictures[i].currentContrast
		}
		targetIntensity /= float64(len(pictures))
		targetContrast /= float64(len(pictures))
		for i := range pictures {
			pictures[i].targetIntensity = targetIntensity
			pictures[i].targetContrast = targetContrast
		}
	} else {
		for i := range pictures {
			targetIntensity = 0
			targetContrast = 0.0
			var start = maximum(i-config.rollingaverage, 0)
			var end = minimum(i+config.rollingaverage, len(pictures)-1)
			for j := start; j <= end; j++ {
				targetIntensity += pictures[j].currentIntensity
				targetContrast += pictures[j].currentContrast
			}
			targetIntensity /= float64(end - start + 2)
			targetContrast /= float64(end - start + 2)
			pictures[i].targetIntensity = targetIntensity
			pictures[i].targetContrast = targetContrast
		}
	}
	//Create token channel and fill it with inital tokens
	tokens = make(chan bool, config.threads)
	for i := 0; i < config.threads; i++ {
		tokens <- true
	}
	printDebug()
	//Run the loop for image adjustment
	for i := range pictures {
		_ = <-tokens
		go func(i int) {
			defer func() {
				progressBars["ADJUST"].Incr()
				tokens <- true
			}()
			var img, _ = imaging.Open(pictures[i].path)

			pictures[i].requiredGammaChange = calculateGammaDifference(img, pictures[i].targetIntensity, 2)
			img = imaging.AdjustGamma(img, pictures[i].requiredGammaChange)

			pictures[i].currentIntensity = measureIntensity(img, 2)
			pictures[i].currentContrast = measureContrast(img, pictures[i].currentIntensity, 2)
			pictures[i].requiredContrastChange = 100 * (pictures[i].targetContrast/pictures[i].currentContrast - 1) * 0
			img = imaging.AdjustContrast(img, pictures[i].requiredContrastChange)

			pictures[i].requiredIntensityChange = calculateIntensityDifference(img, pictures[i].targetIntensity, 2)
			img = imaging.AdjustBrightness(img, pictures[i].requiredIntensityChange/65536*100)

			imaging.Save(img, filepath.Join(config.destination, filepath.Base(pictures[i].path)), imaging.JPEGQuality(95), imaging.PNGCompressionLevel(0))
		}(i)
	}
	for i := 0; i < config.threads; i++ {
		_ = <-tokens
	}
	//uiprogress.Stop()
	printDebug()
}
