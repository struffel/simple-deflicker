package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/ncruces/zenity"

	"github.com/struffel/simple-deflicker/internal/deflicker"
)

func StartGUI() error {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("Simple Deflicker"), app.Size(unit.Dp(400), unit.Dp(500)))
		if err := runWindow(w); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
	return nil
}

func runWindow(w *app.Window) error {
	theme := material.NewTheme()
	state := newUiState(deflicker.DefaultSettings())

	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			receiveGuiResults(state)
			handleGuiEvents(gtx, w, state)
			layoutGui(gtx, theme, state)
			e.Frame(gtx.Ops)
		}
	}
}

// receiveGuiResults applies any results delivered by background goroutines
// (directory pickers, deflickering) to the widget state.
func receiveGuiResults(state *uiState) {
	select {
	case dir := <-state.sourceResult:
		state.sourceEditor.SetText(filepath.ToSlash(dir))
	default:
	}
	select {
	case dir := <-state.destinationResult:
		state.destinationEditor.SetText(filepath.ToSlash(dir))
	default:
	}
	select {
	case err := <-state.deflickerResult:
		state.processing = false
		state.setProgress(0, "")
		if err != nil {
			state.statusText = "An error occurred: " + err.Error()
			go zenity.Error(err.Error(), zenity.Title("Simple Deflicker - Error"))
		} else {
			state.statusText = "Saved pictures into " + state.Settings.DestinationDirectory
			go zenity.Info(state.statusText, zenity.Title("Simple Deflicker"))
		}
	default:
	}
}

func handleGuiEvents(gtx layout.Context, w *app.Window, state *uiState) {
	if state.processing {
		return
	}

	state.formatEnum.Update(gtx)
	state.Settings.OutFormat = deflicker.OutputFormat(state.formatEnum.Value)

	if state.browseSourceBtn.Clicked(gtx) {
		go func() {
			dir, err := zenity.SelectFile(zenity.Directory(), zenity.Title("Select a source directory."))
			if err == nil {
				state.sourceResult <- dir
				w.Invalidate()
			}
		}()
	}
	if state.browseDestinationBtn.Clicked(gtx) {
		go func() {
			dir, err := zenity.SelectFile(zenity.Directory(), zenity.Title("Select a destination directory."))
			if err == nil {
				state.destinationResult <- dir
				w.Invalidate()
			}
		}()
	}
	if state.startBtn.Clicked(gtx) {
		startDeflickering(w, state)
	}
}

// startDeflickering reads the current widget values into the settings, then
// runs the deflickering process in the background while the native progress
// bar and start button give the user feedback.
func startDeflickering(w *app.Window, state *uiState) {
	state.Settings.SourceDirectory = filepath.ToSlash(state.sourceEditor.Text())
	state.Settings.DestinationDirectory = filepath.ToSlash(state.destinationEditor.Text())
	if v, err := strconv.Atoi(state.rollingAvgEditor.Text()); err == nil {
		state.Settings.RollingAverage = v
	}
	if v, err := strconv.Atoi(state.jpegQualityEditor.Text()); err == nil {
		state.Settings.JpegQuality = v
	}

	if validationErrors := state.Settings.Validate(); len(validationErrors) > 0 {
		msg := ""
		for _, validationError := range validationErrors {
			msg += validationError.Error() + "\n"
		}
		go zenity.Error(msg, zenity.Title("Simple Deflicker - Invalid settings"))
		return
	}

	state.processing = true
	state.statusText = "Processing..."
	settings := state.Settings

	go func() {
		updater := &guiUpdater{state: state, win: w}
		err := deflicker.Run(settings, updater)
		state.deflickerResult <- err
		w.Invalidate()
	}()
}

// guiUpdater implements progress.Updater by writing progress into the UiState
// and invalidating the window so the native progress bar redraws.
type guiUpdater struct {
	state *uiState
	win   *app.Window
}

func (u *guiUpdater) Start() {
	u.state.setProgress(0, "Starting...")
	u.win.Invalidate()
}

func (u *guiUpdater) Increment(msg string, phase string, completed int, ofTotal int) {
	var fraction float32
	if ofTotal > 0 {
		fraction = float32(completed) / float32(ofTotal)
	}
	u.state.setProgress(fraction, fmt.Sprintf("%s: %s (%d/%d)", phase, msg, completed, ofTotal))
	u.win.Invalidate()
}

func (u *guiUpdater) Finish() {
	u.state.setProgress(1, "Finishing...")
	u.win.Invalidate()
}
