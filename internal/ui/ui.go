package ui

import (
	"strconv"

	"gioui.org/widget"
	"github.com/struffel/simple-deflicker/internal"
)

type UiState struct {
	Settings internal.Settings

	sourceResult      chan string
	destinationResult chan string
	deflickerResult   chan error

	processing bool
	statusText string

	sourceEditor      widget.Editor
	destinationEditor widget.Editor
	rollingAvgEditor  widget.Editor
	jpegQualityEditor widget.Editor
	threadsEditor     widget.Editor

	browseSourceBtn      widget.Clickable
	browseDestinationBtn widget.Clickable
	startBtn             widget.Clickable
}

func NewUiState(settings internal.Settings) *UiState {
	state := &UiState{
		sourceResult:      make(chan string, 1),
		destinationResult: make(chan string, 1),
		deflickerResult:   make(chan error, 1),
		processing:        false,
		statusText:        "",
	}
	state.sourceEditor.SingleLine = true
	state.sourceEditor.SetText(settings.SourceDirectory)

	state.destinationEditor.SingleLine = true
	state.destinationEditor.SetText(settings.DestinationDirectory)

	state.rollingAvgEditor.SingleLine = true
	state.rollingAvgEditor.Filter = "0123456789"
	state.rollingAvgEditor.SetText(strconv.Itoa(settings.RollingAverage))

	state.jpegQualityEditor.SingleLine = true
	state.jpegQualityEditor.Filter = "0123456789"
	state.jpegQualityEditor.SetText(strconv.Itoa(settings.JpegCompression))

	state.threadsEditor.SingleLine = true
	state.threadsEditor.Filter = "0123456789"
	state.threadsEditor.SetText(strconv.Itoa(settings.Threads))
	return state
}
