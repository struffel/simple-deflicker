package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"time"

	"github.com/disintegration/imaging"
)

func main() {
	fmt.Println("Starting...")
	var maxThreads = runtime.NumCPU()

	runtime.GOMAXPROCS(maxThreads)

	files, err := ioutil.ReadDir("./input/")
	if err != nil {
		log.Fatal(err)
	}

	var numberOfFiles = len(files)
	fmt.Printf("Number of Files: %v\n", numberOfFiles)

	brightnessValues := make([]uint16, numberOfFiles)

	var ongoingThreads = 0
	fmt.Printf("Using %v threads...\n", maxThreads)

	for i, file := range files {
		for ongoingThreads >= maxThreads {
			time.Sleep(time.Millisecond * 5)
		}
		ongoingThreads++
		go func(i int, fileName string) {
			defer func() {
				ongoingThreads--
			}()
			if filepath.Ext(fileName) == ".JPG" {
				var img, _ = imaging.Open(fileName)
				brightnessValues[i] = getAverageImageBrightness(img, 8)
				fmt.Printf("%v | %v\n", fileName, brightnessValues[i])
			}
		}(i, "./input/"+file.Name())
	}
	for ongoingThreads > 0 {
		time.Sleep(time.Millisecond * 5)
	}
	fmt.Println("All threads finished!")

	fmt.Print("\n")

	var sumBrightness uint64 = 0
	//for _, value := range brightnessValues {
	for i := 0; i < numberOfFiles; i++ {
		sumBrightness += uint64(brightnessValues[i])
	}

	var averageBrightness uint16 = uint16(float64(sumBrightness) / float64(numberOfFiles))

	fmt.Printf("AVG Brightness: %v\n", averageBrightness)

	for i, file := range files {
		for ongoingThreads >= maxThreads {
			time.Sleep(time.Millisecond * 5)
		}
		ongoingThreads++
		go func(i int, fileName string) {
			defer func() {
				ongoingThreads--
			}()
			if filepath.Ext(fileName) == ".JPG" {
				var img, _ = imaging.Open(fileName)
				var gamma = float64(averageBrightness) / float64(brightnessValues[i])
				var imgCorrected = imaging.AdjustGamma(img, gamma)
				imaging.Save(imgCorrected, "./output/"+filepath.Base(fileName), imaging.JPEGQuality(95), imaging.PNGCompressionLevel(0))
				fmt.Printf("%v | %v\n", fileName, gamma)
			}
		}(i, "./input/"+file.Name())
	}
	for ongoingThreads > 0 {
		time.Sleep(time.Millisecond * 5)
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
