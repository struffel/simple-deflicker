package main

import (
	"errors"
	"flag"
	"runtime"
)

type configuration struct {
	sourceDirectory      string
	destinationDirectory string
	rollingAverage       int
	jpegCompression      int
	threads              int
}

func collectConfigInformation() configuration {
	var config configuration
	flag.StringVar(&config.sourceDirectory, "source", "", "Directory with the images to process.")
	flag.StringVar(&config.destinationDirectory, "destination", "", "Directory to put the processed images in.")
	flag.IntVar(&config.rollingAverage, "rollingAverage", 15, "Number of frames to use for rolling average. 0 disables it.")
	flag.IntVar(&config.jpegCompression, "jpegCompression", 95, "Level of JPEG compression. Must be between 1 - 100. Default is 95.")
	flag.IntVar(&config.threads, "threads", runtime.NumCPU(), "Number of threads to use. Default is the detected number of cores.")
	flag.Parse()
	return config
}
func validateConfigInformation() error {
	description := ""
	//Test for illegal inputs
	if config.jpegCompression < 1 || config.jpegCompression > 100 {
		description += "Invalid JPEG compression setting. Value must be between 1 and 100 (inclusive).\n"
	}
	if config.threads < 1 {
		description += "Invalid number of threads. There must be at least one thread.\n"
	}
	if config.rollingAverage < 1 {
		description += "Invalid rolling average. Value must be equal to or greater than 1.\n"
	}
	if config.sourceDirectory == "" {
		description += "No source directory specified.\n"
	} else if !testForDirectory(config.sourceDirectory) {
		description += "The source directory could not be found.\n"
	}
	if config.destinationDirectory == "" {
		description += "No destination directory specified.\n"
	} else if !testForDirectory(config.destinationDirectory) {
		description += "The destination directory could not be found.\n"
	}
	if description != "" {
		return errors.New(description)
	} else {
		return nil
	}
}
