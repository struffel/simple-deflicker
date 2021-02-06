package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/style"
	"github.com/sqweek/dialog"

	"github.com/disintegration/imaging"
)

type picture struct {
	currentPath         string
	targetPath          string
	currentRgbHistogram rgbHistogram
	targetRgbHistogram  rgbHistogram
}

var guiComponents struct {
	sourceField          nucular.TextEditor
	destinationField     nucular.TextEditor
	rollingAverageField  nucular.TextEditor
	threadsField         nucular.TextEditor
	jpegCompressionField nucular.TextEditor
}

func main() {

	//Initial console output
	printInfo()

	//Read parameters from console
	config := collectConfigInformation()

	window := nucular.NewMasterWindowSize(0, "Simple Deflicker", image.Point{450, 450}, windowUpdateFunction)
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
	window.Main()
	os.Exit(3)

	//Preparations
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

func runDeflickering(config configuration) {
	fmt.Println("Starting...")
	runtime.GOMAXPROCS(config.threads)
	pictures := readDirectory(config.sourceDirectory, config.destinationDirectory)
	progress := createProgressBars(len(pictures))
	progress.container.Start()
	//fmt.Printf("%+v\n", pictures)

	//Analyze and create Histograms
	pictures = forEveryPicture(pictures, progress.bars["analyze"], config.threads, func(pic picture) picture {
		var img, err = imaging.Open(pic.currentPath)
		if err != nil {
			fmt.Printf("'%v': %v\n", pic.targetPath, err)
			os.Exit(2)
		}
		pic.currentRgbHistogram = generateRgbHistogramFromImage(img)
		return pic
	})
	//Calculate global or rolling average
	if config.rollingAverage < 1 {
		var averageRgbHistogram rgbHistogram
		for i := range pictures {
			for j := 0; j < 256; j++ {
				averageRgbHistogram.r[j] += pictures[i].currentRgbHistogram.r[j]
				averageRgbHistogram.g[j] += pictures[i].currentRgbHistogram.g[j]
				averageRgbHistogram.b[j] += pictures[i].currentRgbHistogram.b[j]
			}
		}
		for i := 0; i < 256; i++ {
			averageRgbHistogram.r[i] /= uint32(len(pictures))
			averageRgbHistogram.g[i] /= uint32(len(pictures))
			averageRgbHistogram.b[i] /= uint32(len(pictures))
		}
		for i := range pictures {
			pictures[i].targetRgbHistogram = averageRgbHistogram
		}
	} else {
		for i := range pictures {
			var averageRgbHistogram rgbHistogram
			var start = i - config.rollingAverage
			if start < 0 {
				start = 0
			}
			var end = i + config.rollingAverage
			if end > len(pictures)-1 {
				end = len(pictures) - 1
			}
			for i := start; i <= end; i++ {
				for j := 0; j < 256; j++ {
					averageRgbHistogram.r[j] += pictures[i].currentRgbHistogram.r[j]
					averageRgbHistogram.g[j] += pictures[i].currentRgbHistogram.g[j]
					averageRgbHistogram.b[j] += pictures[i].currentRgbHistogram.b[j]
				}
			}
			for i := 0; i < 256; i++ {
				averageRgbHistogram.r[i] /= uint32(end - start + 1)
				averageRgbHistogram.g[i] /= uint32(end - start + 1)
				averageRgbHistogram.b[i] /= uint32(end - start + 1)
			}
			pictures[i].targetRgbHistogram = averageRgbHistogram
		}
	}

	pictures = forEveryPicture(pictures, progress.bars["adjust"], config.threads, func(pic picture) picture {
		var img, _ = imaging.Open(pic.currentPath)
		lut := generateRgbLutFromRgbHistograms(pic.currentRgbHistogram, pic.targetRgbHistogram)
		img = applyRgbLutToImage(img, lut)
		imaging.Save(img, pic.targetPath, imaging.JPEGQuality(config.jpegCompression), imaging.PNGCompressionLevel(0))
		return pic
	})
	progress.container.Stop()
	fmt.Println("Finished.")
}
