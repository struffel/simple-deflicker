package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/sqweek/dialog"
)

func collectConfigInformation() config {
	var config config
	flag.StringVar(&config.sourceDirectory, "source", "", "Directory with the images to process.")
	flag.StringVar(&config.destinationDirectory, "destination", "", "Directory to put the processed images in.")
	flag.IntVar(&config.rollingaverage, "rollingaverage", 15, "Number of frames to use for rolling average. 0 disables it.")
	flag.IntVar(&config.threads, "threads", runtime.NumCPU(), "Number of threads to use")
	flag.Parse()
	if config.sourceDirectory == "" {
		dialog.Message("%s", "No source directory has been specified.\nPlease specify a directory now.").Title("Specify source directory").Info()
		config.sourceDirectory, _ = dialog.Directory().Title("Specify source directory.").Browse()
	}
	if config.destinationDirectory == "" {
		if dialog.Message("No destination directory has been specified.\nWould you like to use '%s' as a destination directoy?", filepath.Join(config.sourceDirectory, "deflickered")).Title("Specify destination directory").YesNo() {
			config.destinationDirectory = filepath.Join(config.sourceDirectory, "deflickered")
		} else {
			config.destinationDirectory, _ = dialog.Directory().Title("Specify destination directory.").Browse()
		}
	}
	fmt.Println(config)
	return config
}
