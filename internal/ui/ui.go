package ui

import (
	"strconv"

	"gioui.org/widget"
	"github.com/struffel/simple-deflicker/internal/deflicker"
)

type UiState struct {
	Settings deflicker.Settings

	sourceResult      chan string
	destinationResult chan string
	deflickerResult   chan error

	processing bool
	statusText string

	sourceEditor      widget.Editor
	destinationEditor widget.Editor
	rollingAvgEditor  widget.Editor
	jpegQualityEditor widget.Editor
	formatEnum        widget.Enum

	browseSourceBtn      widget.Clickable
	browseDestinationBtn widget.Clickable
	startBtn             widget.Clickable
}

func NewUiState(settings deflicker.Settings) *UiState {
	state := &UiState{
		Settings:          settings,
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
	state.jpegQualityEditor.SetText(strconv.Itoa(settings.JpegQuality))

	state.formatEnum.Value = string(settings.OutFormat)
	return state
}

// DefaultSettings returns the settings the GUI is pre-populated with.
func DefaultSettings() deflicker.Settings {
	return deflicker.Settings{
		RollingAverage: 15,
		OutFormat:      deflicker.FormatPng,
		JpegQuality:    95,
	}
}
