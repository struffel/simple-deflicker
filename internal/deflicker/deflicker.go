package deflicker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/struffel/simple-deflicker/internal/progress"
	"golang.org/x/sync/errgroup"
)

type pictureInfo struct {
	Name                 string
	OriginalRgbHistogram rgbHistogram
	DesiredRgbHistogram  rgbHistogram
}

func Run(settings Settings, updater progress.Updater) error {

	// Validate config information
	settingsErrors := settings.Validate()
	if len(settingsErrors) > 0 {
		return fmt.Errorf("Settings validation failed: %v", settingsErrors)
	}

	updater.Start()

	// Read the entire source directory and create Picture structs with histograms
	pictures, err := getSourcePictureInfo(settings.SourceDirectory, updater)
	if err != nil {
		return err
	}

	err = fillDesiredHistograms(&pictures, updater, settings.RollingAverage)
	if err != nil {
		return err
	}

	adjustImages(&pictures, settings, updater)

	updater.Finish()
	return nil
}

func adjustImages(pictures *[]pictureInfo, settings Settings, updater progress.Updater) error {
	total := len(*pictures)
	var completed atomic.Int32

	var g errgroup.Group
	g.SetLimit(runtime.NumCPU())

	for i := range *pictures {
		g.Go(func() error {
			sourcePath := filepath.Join(settings.SourceDirectory, (*pictures)[i].Name)

			destinationExtension := settings.OutFormat.Extension()
			destinationFileName := strings.TrimSuffix((*pictures)[i].Name, filepath.Ext((*pictures)[i].Name)) + destinationExtension
			destinationPath := filepath.Join(settings.DestinationDirectory, destinationFileName)

			sourceImage, err := readImage(sourcePath)
			if err != nil {
				return err
			}

			destinationLut := generateRgbLutFromRgbHistograms((*pictures)[i].OriginalRgbHistogram, (*pictures)[i].DesiredRgbHistogram)
			destinationImage := applyRgbLutToImage(sourceImage, destinationLut)

			if err := saveImage(destinationImage, destinationPath, settings.OutFormat, settings.JpegQuality); err != nil {
				return err
			}

			updater.Increment((*pictures)[i].Name, "Adjusting image", int(completed.Add(1)), total)
			return nil
		})
	}
	return g.Wait()
}

func fillDesiredHistograms(pictures *[]pictureInfo, updater progress.Updater, rollingAverage int) error {

	if rollingAverage < 1 {
		// Simply calculate the global average histogram
		var averageRgbHistogram rgbHistogram
		for i := range *pictures {
			for j := 0; j < 256; j++ {
				averageRgbHistogram.R[j] += (*pictures)[i].OriginalRgbHistogram.R[j]
				averageRgbHistogram.G[j] += (*pictures)[i].OriginalRgbHistogram.G[j]
				averageRgbHistogram.B[j] += (*pictures)[i].OriginalRgbHistogram.B[j]
			}
			//updater.Increment((*pictures)[i].Name, "Calculating average histogram", i+1, len(*pictures))
		}
		for i := 0; i < 256; i++ {
			averageRgbHistogram.R[i] /= uint32(len(*pictures))
			averageRgbHistogram.G[i] /= uint32(len(*pictures))
			averageRgbHistogram.B[i] /= uint32(len(*pictures))
		}
		for i := range *pictures {
			(*pictures)[i].DesiredRgbHistogram = averageRgbHistogram
		}
	} else {
		// Calculate the rolling average histogram for each image
		for i := range *pictures {
			var averageRgbHistogram rgbHistogram
			var start = i - rollingAverage
			if start < 0 {
				start = 0
			}
			var end = i + rollingAverage
			if end > len(*pictures)-1 {
				end = len(*pictures) - 1
			}
			for j := start; j <= end; j++ {
				for k := 0; k < 256; k++ {
					averageRgbHistogram.R[k] += (*pictures)[j].OriginalRgbHistogram.R[k]
					averageRgbHistogram.G[k] += (*pictures)[j].OriginalRgbHistogram.G[k]
					averageRgbHistogram.B[k] += (*pictures)[j].OriginalRgbHistogram.B[k]
				}
			}
			for k := 0; k < 256; k++ {
				averageRgbHistogram.R[k] /= uint32(end - start + 1)
				averageRgbHistogram.G[k] /= uint32(end - start + 1)
				averageRgbHistogram.B[k] /= uint32(end - start + 1)
			}
			(*pictures)[i].DesiredRgbHistogram = averageRgbHistogram
			//updater.Increment((*pictures)[i].Name, "Calculating rolling average histogram", i+1, len(*pictures))
		}
	}

	return nil
}

func getSourcePictureInfo(directory string, updater progress.Updater) ([]pictureInfo, error) {

	// Get raw list of files
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	// Filter down to the compatible image file names
	var imageNames []string
	for _, file := range files {
		extension := strings.ToLower(filepath.Ext(file.Name()))
		if extension == ".jpg" || extension == ".png" {
			imageNames = append(imageNames, file.Name())
		}
	}
	if len(imageNames) < 1 {
		return nil, errors.New("the source directory does not contain any compatible images (JPG or PNG)")
	}

	totalFiles := len(imageNames)
	pictures := make([]pictureInfo, totalFiles)
	var completed atomic.Int32

	var g errgroup.Group
	g.SetLimit(runtime.NumCPU())

	// Calculate histograms concurrently
	for index, name := range imageNames {
		g.Go(func() error {
			image, err := readImage(filepath.Join(directory, name))
			if err != nil {
				return err
			}
			histogram := generateRgbHistogramFromImage(image)
			pictures[index] = pictureInfo{Name: name, OriginalRgbHistogram: histogram, DesiredRgbHistogram: rgbHistogram{}}

			updater.Increment(name, "Calculating histogram", int(completed.Add(1)), totalFiles)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return pictures, nil
}
