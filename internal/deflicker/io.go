package deflicker

import (
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

func directoryExists(path string) bool {
	stat, err := os.Stat(path)
	if err == nil && stat.IsDir() {
		return true
	}
	return false
}

func listImagesInDirectory(path string) ([]string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var images []string
	for _, file := range files {
		extension := strings.ToLower(filepath.Ext(file.Name()))
		if extension == ".jpg" || extension == ".png" {
			images = append(images, filepath.Join(path, file.Name()))
		}
	}
	return images, nil
}

func readImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func saveImage(img image.Image, path string, format OutputFormat, jpegQuality int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	switch format {
	case FormatJpeg:
		err = jpeg.Encode(file, img, &jpeg.Options{Quality: jpegQuality})
	case FormatPng:
		err = png.Encode(file, img)
	default:
		return errors.New("unsupported output format")
	}

	return nil
}
