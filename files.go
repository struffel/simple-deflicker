package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sqweek/dialog"
)

func readDirectory(currentDirectory string, targetDirectory string) []picture {
	var pictures []picture
	//Get list of files
	files, err := ioutil.ReadDir(currentDirectory)
	if err != nil {
		fmt.Printf("'%v': %v\n", currentDirectory, err)
		dialog.Message("%s", "The source directory could not be opened.").Title("Source Directory could not be Loaded").Error()
		os.Exit(1)
	}
	//Prepare slice of pictures
	for _, file := range files {
		var fullSourcePath = filepath.Join(currentDirectory, file.Name())
		var fullTargetPath = filepath.Join(targetDirectory, file.Name())
		var extension = strings.ToLower(filepath.Ext(file.Name()))
		var temp rgbHistogram
		if extension == ".jpg" || extension == ".png" {
			pictures = append(pictures, picture{fullSourcePath, fullTargetPath, temp, temp})
		} else {
			fmt.Printf("'%v': ignoring file with unsupported extension\n", fullSourcePath)
		}
	}
	if len(pictures) < 1 {
		dialog.Message("%s", "The source directory does not contain any compatible images (JPG or PNG). The program will now close.").Title("No Images in Source Directory").Error()
		os.Exit(1)
	}
	return pictures
}

func makeDirectoryIfNotExists(directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return os.Mkdir(directory, os.ModeDir|0755)
	}
	return nil
}
