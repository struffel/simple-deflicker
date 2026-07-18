//go:build !cli
// +build !cli

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/ncruces/zenity"
)

// guiState holds all the widgets and channels used to communicate results
// from background goroutines (file dialogs, processing) back to the single
// goroutine that owns the Gio window and its widgets.
type guiState struct {
	sourceEditor      widget.Editor
	destinationEditor widget.Editor
	rollingAvgEditor  widget.Editor
	jpegQualityEditor widget.Editor
	threadsEditor     widget.Editor

	browseSourceBtn      widget.Clickable
	browseDestinationBtn widget.Clickable
	startBtn             widget.Clickable

	sourceResult      chan string
	destinationResult chan string
	deflickerResult   chan error

	processing bool
	statusText string
}

func newGuiState() *guiState {
	state := &guiState{
		sourceResult:      make(chan string, 1),
		destinationResult: make(chan string, 1),
		deflickerResult:   make(chan error, 1),
	}
	state.sourceEditor.SingleLine = true
	state.sourceEditor.SetText(config.sourceDirectory)

	state.destinationEditor.SingleLine = true
	state.destinationEditor.SetText(config.destinationDirectory)

	state.rollingAvgEditor.SingleLine = true
	state.rollingAvgEditor.Filter = "0123456789"
	state.rollingAvgEditor.SetText(strconv.Itoa(config.rollingAverage))

	state.jpegQualityEditor.SingleLine = true
	state.jpegQualityEditor.Filter = "0123456789"
	state.jpegQualityEditor.SetText(strconv.Itoa(config.jpegCompression))

	state.threadsEditor.SingleLine = true
	state.threadsEditor.Filter = "0123456789"
	state.threadsEditor.SetText(strconv.Itoa(config.threads))
	return state
}

func startGUI() error {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("Simple Deflicker"), app.Size(unit.Dp(480), unit.Dp(420)))
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
	state := newGuiState()

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
// (directory pickers, deflickering) to the widget state. It must only be
// called from the goroutine that owns the window.
func receiveGuiResults(state *guiState) {
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
		if err != nil {
			state.statusText = "An error occurred: " + err.Error()
			go zenity.Error(err.Error(), zenity.Title("Simple Deflicker - Error"))
		} else {
			state.statusText = "Saved pictures into " + config.destinationDirectory
			go zenity.Info(state.statusText, zenity.Title("Simple Deflicker"))
		}
	default:
	}
}

func handleGuiEvents(gtx layout.Context, w *app.Window, state *guiState) {
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
	if state.startBtn.Clicked(gtx) && !state.processing {
		startDeflickering(w, state)
	}
}

// startDeflickering reads the current widget values into config, then runs
// the deflickering process in the background while a zenity progress dialog
// gives the user feedback.
func startDeflickering(w *app.Window, state *guiState) {
	config.sourceDirectory = filepath.ToSlash(state.sourceEditor.Text())
	config.destinationDirectory = filepath.ToSlash(state.destinationEditor.Text())
	if v, err := strconv.Atoi(state.rollingAvgEditor.Text()); err == nil {
		config.rollingAverage = v
	}
	if v, err := strconv.Atoi(state.jpegQualityEditor.Text()); err == nil {
		config.jpegCompression = v
	}
	if v, err := strconv.Atoi(state.threadsEditor.Text()); err == nil {
		config.threads = v
	}

	state.processing = true
	state.statusText = "Processing..."

	go func() {
		dlg, dlgErr := zenity.Progress(
			zenity.Title("Simple Deflicker"),
			zenity.Pulsate(),
			zenity.NoCancel(),
		)
		if dlgErr == nil {
			dlg.Text("Processing images, this can take a while...")
		}

		err := runDeflickering()

		if dlgErr == nil {
			dlg.Complete()
			dlg.Close()
		}

		state.deflickerResult <- err
		w.Invalidate()
	}()
}

func layoutGui(gtx layout.Context, th *material.Theme, state *guiState) layout.Dimensions {
	return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(material.Body1(th, "Source Directory").Layout),
			layout.Rigid(fullWidthBorderedEditor(th, &state.sourceEditor)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
			layout.Rigid(material.Button(th, &state.browseSourceBtn, "Browse").Layout),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			layout.Rigid(material.Body1(th, "Destination Directory").Layout),
			layout.Rigid(fullWidthBorderedEditor(th, &state.destinationEditor)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
			layout.Rigid(material.Button(th, &state.browseDestinationBtn, "Browse").Layout),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			layout.Rigid(material.Body1(th, "Advanced settings").Layout),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(1, labeledEditor(th, "Rolling average", &state.rollingAvgEditor)),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Flexed(1, labeledEditor(th, "JPEG quality", &state.jpegQualityEditor)),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Flexed(1, labeledEditor(th, "Threads", &state.threadsEditor)),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := "Start"
				if state.processing {
					text = "Processing..."
				}
				return material.Button(th, &state.startBtn, text).Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(material.Body2(th, state.statusText).Layout),
		)
	})
}

func labeledEditor(th *material.Theme, label string, ed *widget.Editor) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(material.Caption(th, label).Layout),
			layout.Rigid(fullWidthBorderedEditor(th, ed)),
		)
	}
}

func fullWidthBorderedEditor(th *material.Theme, ed *widget.Editor) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		border := widget.Border{Color: th.Fg, Width: unit.Dp(1), CornerRadius: unit.Dp(4)}
		return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(6)).Layout(gtx, material.Editor(th, ed, "").Layout)
		})
	}
}
