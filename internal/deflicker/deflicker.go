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

type PictureInfo struct {
	Name                 string
	OriginalRgbHistogram RgbHistogram
	DesiredRgbHistogram  RgbHistogram
}

func Run(settings Settings, updater progress.Updater) error {

	// Validate config information
	settingsErrors := settings.Validate()
	if len(settingsErrors) > 0 {
		return fmt.Errorf("Settings validation failed: %v", settingsErrors)
	}

	updater.Start()

	// Read the entire source directory and create Picture structs with histograms
	pictures, err := GetSourcePictureInfo(settings.SourceDirectory, updater)
	if err != nil {
		return err
	}

	err = FillDesiredHistograms(&pictures, updater, settings.RollingAverage)
	if err != nil {
		return err
	}

	AdjustImages(&pictures, settings, updater)

	updater.Finish()
	return nil
}

func AdjustImages(pictures *[]PictureInfo, settings Settings, updater progress.Updater) error {
	total := len(*pictures)
	var completed atomic.Int32

	var g errgroup.Group
	g.SetLimit(runtime.NumCPU())

	for i := range *pictures {
		g.Go(func() error {
			sourcePath := filepath.Join(settings.SourceDirectory, (*pictures)[i].Name)
			destinationPath := filepath.Join(settings.DestinationDirectory, (*pictures)[i].Name)

			sourceImage, err := ReadImage(sourcePath)
			if err != nil {
				return err
			}

			destinationLut := GenerateRgbLutFromRgbHistograms((*pictures)[i].OriginalRgbHistogram, (*pictures)[i].DesiredRgbHistogram)
			destinationImage := ApplyRgbLutToImage(sourceImage, destinationLut)

			if err := SaveImage(destinationImage, destinationPath, settings.OutFormat, settings.JpegQuality); err != nil {
				return err
			}

			updater.Increment((*pictures)[i].Name, "Adjusting image", int(completed.Add(1)), total)
			return nil
		})
	}
	return g.Wait()
}

func FillDesiredHistograms(pictures *[]PictureInfo, updater progress.Updater, rollingAverage int) error {

	if rollingAverage < 1 {
		// Simply calculate the global average histogram
		var averageRgbHistogram RgbHistogram
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
			var averageRgbHistogram RgbHistogram
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

func GetSourcePictureInfo(directory string, updater progress.Updater) ([]PictureInfo, error) {

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
	pictures := make([]PictureInfo, totalFiles)
	var completed atomic.Int32

	var g errgroup.Group
	g.SetLimit(runtime.NumCPU())

	// Calculate histograms concurrently
	for index, name := range imageNames {
		g.Go(func() error {
			image, err := ReadImage(filepath.Join(directory, name))
			if err != nil {
				return err
			}
			histogram := GenerateRgbHistogramFromImage(image)
			pictures[index] = PictureInfo{Name: name, OriginalRgbHistogram: histogram, DesiredRgbHistogram: RgbHistogram{}}

			updater.Increment(name, "Calculating histogram", int(completed.Add(1)), totalFiles)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return pictures, nil
}
