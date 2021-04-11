package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sqweek/dialog"
)

func readDirectory(currentDirectory string, targetDirectory string) ([]picture, error) {
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
		return pictures, errors.New("The source directory does not contain any compatible images (JPG or PNG).")
	}
	return pictures, nil
}

func makeDirectoryIfNotExists(directory string, askUser bool) (bool, error) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		ok := true
		if askUser {
			ok = dialog.Message("'%s' does not exist. Do you want to create it?", directory).Title("Create new directory?").YesNo()
		}
		if ok {
			return true, os.Mkdir(directory, os.ModeDir|0755)
		}
		return false, nil
	}
	return true, nil
}

func testForDirectory(directory string) bool {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return false
	}
	return true
}
