package main

import (
	"fmt"
	"image"
	"path/filepath"

	"github.com/aarzilli/nucular/label"
	"github.com/aarzilli/nucular/style"

	"github.com/aarzilli/nucular"
	"github.com/sqweek/dialog"
)

var window = nucular.NewMasterWindowSize(0, "Simple Deflicker", image.Point{450, 450}, windowUpdateFunction)
var sourceField nucular.TextEditor
var destinationField nucular.TextEditor

func initalizeWindow() {
	window.SetStyle(style.FromTheme(style.DarkTheme, 1.0))
	sourceField.Flags = nucular.EditSelectable | nucular.EditClipboard | nucular.EditSigEnter | nucular.EditIbeamCursor
	sourceField.SingleLine = true
	sourceField.Buffer = []rune(config.sourceDirectory)

	destinationField.Flags = nucular.EditSelectable | nucular.EditClipboard | nucular.EditSigEnter | nucular.EditIbeamCursor
	destinationField.SingleLine = true
	destinationField.Buffer = []rune(config.destinationDirectory)
}

func windowUpdateFunction(w *nucular.Window) {
	//Source Directory
	w.Row(25).Dynamic(1)
	w.Label("Source Directory", "LB")
	sourceField.Edit(w)
	sourceField.Buffer = []rune(filepath.ToSlash(string(sourceField.Buffer)))
	w.Row(25).Ratio(0.333)
	if w.ButtonText("Browse") {
		directory, _ := dialog.Directory().Title("Select a source directory.").Browse()
		sourceField.Buffer = []rune(filepath.ToSlash(directory))
		if len(destinationField.Buffer) == 0 && len(sourceField.Buffer) > 0 {
			destinationField.Buffer = []rune(filepath.Join(string(sourceField.Buffer), "deflickered"))
		}
	}
	w.Row(25).Dynamic(1)
	//Destination Directory
	w.Label("Destination Directory", "LB")
	destinationField.Edit(w)
	destinationField.Buffer = []rune(filepath.ToSlash(string(destinationField.Buffer)))
	w.Row(25).Ratio(0.333)
	if w.ButtonText("Browse") {
		directory, _ := dialog.Directory().Title("Select a destination directory.").Browse()
		destinationField.Buffer = []rune(filepath.ToSlash(directory))
	}
	w.Row(25).Dynamic(1)
	w.Label("Advanced settings", "LB")
	w.PropertyInt("Rolling average", 0, &config.rollingAverage, 100, 1, 1)
	w.PropertyInt("JPEG quality", 1, &config.jpegCompression, 100, 1, 1)
	w.PropertyInt("Threads", 1, &config.threads, 128, 1, 1)
	w.Row(35).Dynamic(1)
	if w.Button(label.T("Start"), false) {
		w.Label("TEST", "LB")
		config.sourceDirectory = string(sourceField.Buffer)
		config.destinationDirectory = string(destinationField.Buffer)
		deflickeringError := runDeflickering()
		if deflickeringError != nil {
			clear()
			fmt.Println("An error occured:")
			fmt.Println(deflickeringError)
		}
	}
}
