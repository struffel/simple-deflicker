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
	brightness       float64
	contrast         float64
	targetBrightness float64
	targetContrast   float64
	requiredGamma    float64
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
	//var inputFolder = "./input/"
	files, err := ioutil.ReadDir(config.source)
	if err != nil {
		log.Fatal(err)
	}
	//Prepare array of pictures
	for _, file := range files {
		var extension = strings.ToLower(filepath.Ext(file.Name()))
		if extension == ".jpg" || extension == ".png" {
			pictures = append(pictures, picture{filepath.Join(config.source, file.Name()), 0, 0, 0, 0, 0})
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
				progressBars["ANALYZE"].Incr()
				tokens <- true
			}()
			var img, _ = imaging.Open(pictures[i].path)
			pictures[i].brightness = measureIntensity(img, 16)
			pictures[i].contrast = measureContrast(img, pictures[i].brightness, 16)
			//pictures[i].kelvin = getAverageImageKelvin(img, 8)
		}(i)
	}
	for i := 0; i < config.threads; i++ {
		_ = <-tokens
	}

	//Calculate global or rolling average
	var targetBrightness float64 = 0
	var targetContrast float64 = 0
	if config.rollingaverage < 1 {
		for i := range pictures {
			targetBrightness += pictures[i].brightness
			targetContrast += pictures[i].contrast
		}
		targetBrightness /= float64(len(pictures))
		targetContrast /= float64(len(pictures))
		for i := range pictures {
			pictures[i].targetBrightness = targetBrightness
			pictures[i].targetContrast = targetContrast
		}
	} else {
		for i := range pictures {
			targetBrightness = 0
			targetContrast = 0.0
			var start = maximum(i-config.rollingaverage, 0)
			var end = minimum(i+config.rollingaverage, len(pictures)-1)
			for j := start; j <= end; j++ {
				targetBrightness += pictures[j].brightness
				targetContrast += pictures[j].contrast
			}
			targetBrightness /= float64(end - start + 2)
			targetContrast /= float64(end - start + 2)
			pictures[i].targetBrightness = targetBrightness
			pictures[i].targetContrast = targetContrast
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

			//------------
			var gamma = calculateGammaDifference(img, pictures[i].targetBrightness, 2)
			img = imaging.AdjustGamma(img, gamma)

			pictures[i].brightness = measureIntensity(img, 2)
			pictures[i].contrast = measureContrast(img, pictures[i].brightness, 2)
			img = imaging.AdjustContrast(img, 100*(pictures[i].targetContrast/pictures[i].contrast-1))
			//fmt.Printf("%v|%v|%v\n", pictures[i].contrast, pictures[i].targetContrast, 100*(pictures[i].targetContrast/pictures[i].contrast-1))
			//img = imaging.AdjustSigmoid(img, float64(pictures[i].brightness)/65536.0, math.E*(pictures[i].targetContrast/pictures[i].contrast-1))

			var brightness = calculateIntensityDifference(img, pictures[i].targetBrightness, 2)
			img = imaging.AdjustBrightness(img, brightness/65536*100)

			//fmt.Printf("%v|%v\n", pictures[i].contrast, 100*(1.0+pictures[i].contrast-pictures[i].targetContrast))
			//imgCorrected = imaging.AdjustContrast(imgCorrected, (1-math.Pow(2.0, exposure))*100.0)
			//imgCorrected = imaging.AdjustSigmoid(imgCorrected, 1.0-gamma, -gamma)
			//fmt.Printf("%v|%v|%v\n", pictures[i].targetBrightness, measureIntensity(imgCorrected, 8), int64(pictures[i].targetBrightness)-int64(measureIntensity(imgCorrected, 8)))
			imaging.Save(img, filepath.Join(config.destination, filepath.Base(pictures[i].path)), imaging.JPEGQuality(95), imaging.PNGCompressionLevel(0))
		}(i)
	}
	for i := 0; i < config.threads; i++ {
		_ = <-tokens
	}
	uiprogress.Stop()
}
