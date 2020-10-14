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
	brightness       uint16
	targetBrightness uint16
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
			pictures = append(pictures, picture{filepath.Join(config.source, file.Name()), 0, 0, 0.0})
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
			pictures[i].brightness = getAverageImageBrightness(img, 16)
			//pictures[i].kelvin = getAverageImageKelvin(img, 8)
		}(i)
	}
	for i := 0; i < config.threads; i++ {
		_ = <-tokens
	}

	//Calculate global or rolling average
	var targetBrightness uint64 = 0
	if config.rollingaverage < 1 {
		for i := range pictures {
			targetBrightness += uint64(pictures[i].brightness)
		}
		targetBrightness /= uint64(len(pictures))
		for i := range pictures {
			pictures[i].targetBrightness = uint16(targetBrightness)
		}
	} else {
		for i := range pictures {
			targetBrightness = 0
			var start = maximum(i-config.rollingaverage, 0)
			var end = minimum(i+config.rollingaverage, len(pictures)-1)
			for j := start; j <= end; j++ {
				targetBrightness += uint64(pictures[j].brightness)
			}
			targetBrightness /= uint64(end - start + 1)
			pictures[i].targetBrightness = uint16(targetBrightness)
		}
	}

	/*//Create token channel and fill it with inital tokens
	tokens = make(chan bool, config.threads)
	for i := 0; i < config.threads; i++ {
		tokens <- true
	}

	//Run the loop for gamma calculation
	for i := range pictures {
		_ = <-tokens
		go func(i int) {
			defer func() {
				progressBars["PREPARE"].Incr()
				tokens <- true
			}()
			var img, _ = imaging.Open(pictures[i].path)
			pictures[i].requiredGamma = getRequiredGamma(img, pictures[i].targetBrightness, 16)
		}(i)
	}
	for i := 0; i < config.threads; i++ {
		_ = <-tokens
	}*/

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
			//var gamma = float64(pictures[i].targetBrightness) / float64(pictures[i].brightness)
			//var gamma = math.Log(float64(pictures[i].brightness)/65536.0) / math.Log(float64(pictures[i].brightness)/65536.0)
			var gamma = getRequiredGamma(img, pictures[i].targetBrightness, 16)
			var imgCorrected = imaging.AdjustGamma(img, gamma)
			//fmt.Printf("%v|%v|%v\n", pictures[i].targetBrightness, getAverageImageBrightness(imgCorrected, 8), int64(pictures[i].targetBrightness)-int64(getAverageImageBrightness(imgCorrected, 8)))
			imaging.Save(imgCorrected, filepath.Join(config.destination, filepath.Base(pictures[i].path)), imaging.JPEGQuality(95), imaging.PNGCompressionLevel(0))
		}(i)
	}
	for i := 0; i < config.threads; i++ {
		_ = <-tokens
	}
	uiprogress.Stop()
}
