package main
import (
	"github.com/aarzilli/nucular/style"
	"fmt"
	"path/filepath"
	"strconv"
	"image"

	"github.com/aarzilli/nucular"
	"github.com/sqweek/dialog"

)

var guiComponents struct {
	sourceField          nucular.TextEditor
	destinationField     nucular.TextEditor
	rollingAverageField  nucular.TextEditor
	threadsField         nucular.TextEditor
	jpegCompressionField nucular.TextEditor
}
var window = nucular.NewMasterWindowSize(0, "Simple Deflicker", image.Point{450, 450}, windowUpdateFunction)

func initalizeWindow(config configuration){
	window.SetStyle(style.FromTheme(style.DarkTheme, 1.0))
	guiComponents.sourceField.Flags = nucular.EditSelectable | nucular.EditClipboard | nucular.EditSigEnter | nucular.EditIbeamCursor
	guiComponents.sourceField.SingleLine = true
	guiComponents.sourceField.Buffer = []rune(config.sourceDirectory)

	guiComponents.destinationField.Flags = nucular.EditSelectable | nucular.EditClipboard | nucular.EditSigEnter | nucular.EditIbeamCursor
	guiComponents.destinationField.SingleLine = true
	guiComponents.destinationField.Buffer = []rune(config.destinationDirectory)

	guiComponents.rollingAverageField.Flags = nucular.EditSelectable | nucular.EditClipboard | nucular.EditSigEnter
	guiComponents.rollingAverageField.SingleLine = true
	guiComponents.rollingAverageField.Filter = nucular.FilterDecimal
	guiComponents.rollingAverageField.Buffer = []rune(fmt.Sprint(config.rollingAverage))

	guiComponents.jpegCompressionField.Flags = nucular.EditSelectable | nucular.EditClipboard | nucular.EditSigEnter
	guiComponents.jpegCompressionField.SingleLine = true
	guiComponents.jpegCompressionField.Filter = nucular.FilterDecimal
	guiComponents.jpegCompressionField.Buffer = []rune(fmt.Sprint(config.jpegCompression))

	guiComponents.threadsField.Flags = nucular.EditSelectable | nucular.EditClipboard | nucular.EditSigEnter
	guiComponents.threadsField.SingleLine = true
	guiComponents.threadsField.Filter = nucular.FilterDecimal
	guiComponents.threadsField.Buffer = []rune(fmt.Sprint(config.threads))
}

func windowUpdateFunction(w *nucular.Window) {
	w.Row(25).Dynamic(1)
	w.Label("Source Directory", "LB")
	guiComponents.sourceField.Edit(w)
	guiComponents.sourceField.Buffer = []rune(filepath.ToSlash(string(guiComponents.sourceField.Buffer)))
	w.Row(25).Ratio(0.333)
	if w.ButtonText("Browse") {
		directory, _ := dialog.Directory().Title("Select a source directory.").Browse()
		guiComponents.sourceField.Buffer = []rune(filepath.ToSlash(directory))
		if len(guiComponents.destinationField.Buffer) == 0 && len(guiComponents.sourceField.Buffer) > 0 {
			guiComponents.destinationField.Buffer = []rune(filepath.Join(string(guiComponents.sourceField.Buffer), "deflickered"))
		}
	}
	w.Row(5).Dynamic(1)
	w.Row(25).Dynamic(1)
	w.Label("Destination Directory", "LB")
	guiComponents.destinationField.Edit(w)
	guiComponents.destinationField.Buffer = []rune(filepath.ToSlash(string(guiComponents.destinationField.Buffer)))
	w.Row(25).Ratio(0.333)
	if w.ButtonText("Browse") {
		directory, _ := dialog.Directory().Title("Select a destination directory.").Browse()
		guiComponents.destinationField.Buffer = []rune(filepath.ToSlash(directory))
	}
	w.Row(25).Dynamic(1)
	w.Label("Rolling average", "LB")
	guiComponents.rollingAverageField.Edit(w)
	w.Row(25).Dynamic(1)
	w.Label("JPEG Compression", "LB")
	guiComponents.jpegCompressionField.Edit(w)
	w.Row(25).Dynamic(1)
	w.Label("Threads", "LB")
	guiComponents.threadsField.Edit(w)
	w.Row(30).Dynamic(1)
	if w.ButtonText("Start") {
		var config configuration
		config.sourceDirectory = string(guiComponents.sourceField.Buffer)
		config.destinationDirectory = string(guiComponents.destinationField.Buffer)
		config.rollingAverage, _ = strconv.Atoi(string(guiComponents.rollingAverageField.Buffer))
		config.jpegCompression, _ = strconv.Atoi(string(guiComponents.jpegCompressionField.Buffer))
		config.threads, _ = strconv.Atoi(string(guiComponents.threadsField.Buffer))
		isValid, description := validateConfigInformation(config)
		if isValid {
			runDeflickering(config)
		} else {
			fmt.Print(description)
			dialog.Message("Invalid settings:\n%s", description).Title("Invalid settings").Error()
		}

	}
}