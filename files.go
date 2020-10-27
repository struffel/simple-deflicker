package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func createPictureSliceFromDirectory(currentDirectory string, targetDirectory string) []picture {
	var pictures []picture
	//Get list of files
	files, err := ioutil.ReadDir(currentDirectory)
	if err != nil {
		fmt.Printf("'%v': %v\n", currentDirectory, err)
		os.Exit(1)
	}
	//Prepare slice of pictures
	for _, file := range files {
		var fullSourcePath = filepath.Join(currentDirectory, file.Name())
		var fullTargetPath = filepath.Join(targetDirectory, file.Name())
		var extension = strings.ToLower(filepath.Ext(file.Name()))
		var temp histogram
		if extension == ".jpg" || extension == ".png" {
			pictures = append(pictures, picture{fullSourcePath, fullTargetPath, temp, temp})
		} else {
			fmt.Printf("'%v': ignoring file with unsupported extension\n", fullSourcePath)
		}
	}
	return pictures
}

func makeDirectoryIfNotExists(directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return os.Mkdir(directory, os.ModeDir|0755)
	}
	return nil
}
