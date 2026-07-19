package deflicker

import (
	"flag"
	"fmt"
)

type OutputFormat string

const (
	FormatJpeg OutputFormat = "jpeg"
	FormatPng  OutputFormat = "png"
)

func (f OutputFormat) Extension() string {
	switch f {
	case FormatJpeg:
		return ".jpg"
	case FormatPng:
		return ".png"
	default:
		panic("Unknown output format")
	}
}

type Settings struct {
	SourceDirectory      string
	DestinationDirectory string
	RollingAverage       int
	OutFormat            OutputFormat
	JpegQuality          int
}

// DefaultSettings returns the settings the GUI is pre-populated and that the CLI uses if no arguments are provided.
func DefaultSettings() Settings {
	return Settings{
		RollingAverage: 15,
		OutFormat:      FormatPng,
		JpegQuality:    95,
	}
}

func NewSettingsFromArgs() Settings {
	var defaultSettings = DefaultSettings()
	var settings Settings
	var tmpFormat string
	flag.StringVar(&settings.SourceDirectory, "source", "", "Directory with the images to process.")
	flag.StringVar(&settings.DestinationDirectory, "destination", "", "Directory to put the processed images in.")
	flag.IntVar(&settings.RollingAverage, "rollingAverage", defaultSettings.RollingAverage, "Number of frames to use for rolling average. 0 disables it.")
	flag.StringVar(&tmpFormat, "format", string(defaultSettings.OutFormat), "Output format. Options are jpeg png.")
	flag.IntVar(&settings.JpegQuality, "jpegQuality", defaultSettings.JpegQuality, "Level of JPEG compression. Must be between 1 - 100.")
	flag.Parse()
	settings.OutFormat = OutputFormat(tmpFormat)
	return settings
}

func (s *Settings) Validate() []error {
	errors := []error{}

	if s.JpegQuality < 1 || s.JpegQuality > 100 {
		errors = append(errors, fmt.Errorf("Invalid JPEG compression setting. Value must be between 1 and 100 (inclusive)."))
	}
	if s.RollingAverage < 0 {
		errors = append(errors, fmt.Errorf("Invalid rolling average. Value must be equal to or greater than 0, with 0 disabling it."))
	}
	if s.OutFormat != FormatJpeg && s.OutFormat != FormatPng {
		errors = append(errors, fmt.Errorf("Invalid output format. Options are jpeg png."))
	}

	if s.SourceDirectory == "" {
		errors = append(errors, fmt.Errorf("No source directory specified."))
	} else if !directoryExists(s.SourceDirectory) {
		errors = append(errors, fmt.Errorf("The source directory could not be found."))
	}
	if s.DestinationDirectory == "" {
		errors = append(errors, fmt.Errorf("No destination directory specified."))
	} else if !directoryExists(s.DestinationDirectory) {
		errors = append(errors, fmt.Errorf("The destination directory could not be found."))
	}
	return errors
}
