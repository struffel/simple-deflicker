package ui

import (
	"strconv"
	"sync"

	"gioui.org/widget"
	"github.com/struffel/simple-deflicker/internal/deflicker"
)

type uiState struct {
	Settings deflicker.Settings

	sourceResult      chan string
	destinationResult chan string
	deflickerResult   chan error

	processing bool
	statusText string

	progressMu       sync.Mutex
	progressFraction float32
	progressText     string

	sourceEditor      widget.Editor
	destinationEditor widget.Editor
	rollingAvgEditor  widget.Editor
	jpegQualityEditor widget.Editor
	formatEnum        widget.Enum

	browseSourceBtn      widget.Clickable
	browseDestinationBtn widget.Clickable
	startBtn             widget.Clickable
}

func newUiState(settings deflicker.Settings) *uiState {
	state := &uiState{
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

// setProgress updates the current progress fraction (0..1) and status text.
// It is safe to call from any goroutine.
func (s *uiState) setProgress(fraction float32, text string) {
	s.progressMu.Lock()
	s.progressFraction = fraction
	s.progressText = text
	s.progressMu.Unlock()
}

// progress returns the current progress fraction (0..1) and status text. It
// is safe to call from any goroutine.
func (s *uiState) progress() (float32, string) {
	s.progressMu.Lock()
	defer s.progressMu.Unlock()
	return s.progressFraction, s.progressText
}
