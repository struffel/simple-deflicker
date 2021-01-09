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

type configuration struct {
	sourceDirectory      string
	destinationDirectory string
	rollingaverage       int
	jpegcompression      int
	threads              int
}

func collectConfigInformation() configuration {
	var config configuration
	flag.StringVar(&config.sourceDirectory, "source", "", "Directory with the images to process.")
	flag.StringVar(&config.destinationDirectory, "destination", "", "Directory to put the processed images in.")
	flag.IntVar(&config.rollingaverage, "rollingaverage", 15, "Number of frames to use for rolling average. 0 disables it.")
	flag.IntVar(&config.jpegcompression, "jpegcompression", 95, "Level of JPEG compression. Must be between 1 - 100. Default is 95.")
	flag.IntVar(&config.threads, "threads", runtime.NumCPU(), "Number of threads to use. Default is the detected number of cores.")
	flag.Parse()

	//Test for illegal inputs
	if config.jpegcompression < 1 || config.jpegcompression > 100 {
		log.Fatalln("'jpegcompression' must be a value between 1 and 100")
	}
	if config.threads < 1 {
		log.Fatalln("'threads' must be greater than 0")
	}
	//Test for missing directory inputs
	if config.sourceDirectory == "" {
		dialog.Message("%s", "No source directory has been specified.\nPlease specify a directory now.").Title("Specify source directory").Info()
		sourceDirectory, err := dialog.Directory().Title("Specify source directory.").Browse()
		if err != nil {
			if errors.Is(err, dialog.ErrCancelled) {
				dialog.Message("%s", "Source directory selection was canceled. The program will now close.").Title("No Source Selected").Info()
				os.Exit(1)
			} else {
				dialog.Message("%s: %v", "Unexpected error", err).Title("Unexpected Error").Error()
				log.Fatalln(err)
			}
		}

		config.sourceDirectory = sourceDirectory
	}
	if config.destinationDirectory == "" {
		if dialog.Message("No destination directory has been specified.\nWould you like to use '%s' as a destination directoy?", filepath.Join(config.sourceDirectory, "deflickered")).Title("Specify destination directory").YesNo() {
			config.destinationDirectory = filepath.Join(config.sourceDirectory, "deflickered")
		} else {
			destinationDirectory, err := dialog.Directory().Title("Specify destination directory.").Browse()
			if err != nil {
				if errors.Is(err, dialog.ErrCancelled) {
					dialog.Message("%s", "Destination directory selection was canceled. The program will now close.").Title("No Destination Selected").Info()
					os.Exit(1)
				} else {
					dialog.Message("%s: %v", "Unexpected error", err).Title("Unexpected Error").Error()
					log.Fatalln(err)
				}
			}
			config.destinationDirectory = destinationDirectory
		}
	}
	return config
}
