package deflicker

import (
	"errors"
	"fmt"
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

	fillDesiredHistograms(pictures, settings.RollingAverage)

	if err := adjustImages(&pictures, settings, updater); err != nil {
		return err
	}

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

			histogramComparison := compareHistograms((*pictures)[i].OriginalRgbHistogram, (*pictures)[i].DesiredRgbHistogram)

			updater.Increment(
				fmt.Sprintf("Adjusted image %s (Shift: %+.5f / EMD: %.5f / Consistency: %.2f)",
					(*pictures)[i].Name, histogramComparison.Shift, histogramComparison.Distance, histogramComparison.Consistency), int(completed.Add(1)), total)
			return nil
		})
	}
	return g.Wait()
}

// fillDesiredHistograms sets the desired histogram of each picture to the average of the original
// histograms within the rolling window. A rollingAverage below 1 uses all pictures as the window.
func fillDesiredHistograms(pictures []pictureInfo, rollingAverage int) {
	last := len(pictures) - 1

	// The global average is computed once and shared, since the window is identical for every picture.
	var globalAverage rgbHistogram
	if rollingAverage < 1 {
		globalAverage = averageOriginalHistograms(pictures, 0, last)
	}

	for i := range pictures {
		if rollingAverage < 1 {
			pictures[i].DesiredRgbHistogram = globalAverage
		} else {
			start := max(i-rollingAverage, 0)
			end := min(i+rollingAverage, last)
			pictures[i].DesiredRgbHistogram = averageOriginalHistograms(pictures, start, end)
		}
	}
}

// averageOriginalHistograms returns the average of the original histograms of pictures[start..end] (inclusive).
func averageOriginalHistograms(pictures []pictureInfo, start, end int) rgbHistogram {
	var average rgbHistogram
	for j := start; j <= end; j++ {
		for k := 0; k < 256; k++ {
			average.R[k] += pictures[j].OriginalRgbHistogram.R[k]
			average.G[k] += pictures[j].OriginalRgbHistogram.G[k]
			average.B[k] += pictures[j].OriginalRgbHistogram.B[k]
		}
	}
	count := uint32(end - start + 1)
	for k := 0; k < 256; k++ {
		average.R[k] /= count
		average.G[k] /= count
		average.B[k] /= count
	}
	return average
}

func getSourcePictureInfo(directory string, updater progress.Updater) ([]pictureInfo, error) {

	imageNames, err := listImagesInDirectory(directory)
	if err != nil {
		return nil, err
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

			updater.Increment(fmt.Sprintf("Calculated histogram for %s", name), int(completed.Add(1)), totalFiles)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return pictures, nil
}
