package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/struffel/simple-deflicker/internal/deflicker"
	"github.com/struffel/simple-deflicker/internal/progress"
)

func main() {

	// Read CLI parameters.
	settings := deflicker.NewSettingsFromArgs()

	// In CLI mode, check the settings immediately
	validationErrors := settings.Validate()
	if len(validationErrors) > 0 {
		fmt.Fprintln(os.Stderr, "Invalid settings:")
		for _, err := range validationErrors {
			fmt.Fprintln(os.Stderr, err)
		}
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	// Run the actual deflickering
	deflickeringError := deflicker.Run(settings, &progress.ConsoleUpdater{})
	if deflickeringError != nil {
		fmt.Println("An error occured:")
		fmt.Println(deflickeringError)
		os.Exit(1)
	}
	os.Exit(0)

}
