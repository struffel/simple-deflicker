package main

import (
	"errors"
	"flag"
	"log"
	"os"
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
		srcDir, err := dialog.Directory().Title("Specify source directory.").Browse()
		if err != nil {
			if errors.Is(err, dialog.ErrCancelled) {
				dialog.Message("%s", "Source directory selection was canceled.  Exiting.").Title("No Source Selected").Info()
				os.Exit(1)
			} else {
				dialog.Message("%s: %v", "Unexpected error", err).Title("Unexpected Error").Error()
				log.Fatalln(err)
			}
		}

		config.sourceDirectory = srcDir
	}
	if config.destinationDirectory == "" {
		if dialog.Message("No destination directory has been specified.\nWould you like to use '%s' as a destination directoy?", filepath.Join(config.sourceDirectory, "deflickered")).Title("Specify destination directory").YesNo() {
			config.destinationDirectory = filepath.Join(config.sourceDirectory, "deflickered")
		} else {
			config.destinationDirectory, _ = dialog.Directory().Title("Specify destination directory.").Browse()
		}
	}
	return config
}
