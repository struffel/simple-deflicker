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
				fmt.Println(fileName)
			}
		}(i, "./input/"+file.Name())
	}
	for ongoingThreads > 0 {
		time.Sleep(time.Millisecond * 5)
	}
	fmt.Println("All threads finished!")

	fmt.Print("\n")

	var sum uint64 = 0
	for _, value := range brightnessValues {
		sum += uint64(value)
	}
	var average uint16 = uint16(float64(sum) / float64(len(brightnessValues)))
	fmt.Printf("AVG: %v\n", average)

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
				var imgCorrected = imaging.AdjustGamma(img, float64(average)/float64(brightnessValues[i]))
				imaging.Save(imgCorrected, "./output/"+filepath.Base(fileName), imaging.JPEGQuality(95), imaging.PNGCompressionLevel(0))
				fmt.Println(fileName)
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

/* WIP
getAverageImageHue(input image.Image, precision int) uint16 {
	var sum, pixels uint64
	var inputGray = imaging.Grayscale(input)
	for y := inputGray.Bounds().Min.Y; y < inputGray.Bounds().Max.Y; y += precision {
		for x := inputGray.Bounds().Min.X; x < inputGray.Bounds().Max.X; x += precision {
			r, g, b, alpha := inputGray.At(x, y).RGBA()
			if alpha > 0 {
				rr := float64(r) / 255
				gg := float64(g) / 255
				bb := float64(b) / 255

				max := math.Max(rr, math.Max(gg, bb))
				min := math.Min(rr, math.Min(gg, bb))

				l := (max + min) / 2

				var h, s float64
				d := max - min
				if l > 0.5 {
					s = d / (2 - max - min)
				} else {
					s = d / (max + min)
				}
			}
		}
	}
	return uint16(sum / pixels)
}*/
