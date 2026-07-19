package ui

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/struffel/simple-deflicker/internal/deflicker"
)

func layoutGui(gtx layout.Context, th *material.Theme, state *uiState) layout.Dimensions {
	return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Title

			// Source directory
			layout.Rigid(material.Body1(th, "Source Directory").Layout),
			layout.Rigid(fullWidthBorderedEditor(th, &state.sourceEditor, false)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
			layout.Rigid(material.Button(th, &state.browseSourceBtn, "Browse").Layout),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			// Destination directory
			layout.Rigid(material.Body1(th, "Destination Directory").Layout),
			layout.Rigid(fullWidthBorderedEditor(th, &state.destinationEditor, false)),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
			layout.Rigid(material.Button(th, &state.browseDestinationBtn, "Browse").Layout),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			// Advanced settings
			layout.Rigid(material.Body1(th, "Advanced settings").Layout),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				jpegQualityDisabled := state.Settings.OutFormat != deflicker.FormatJpeg
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(1, labeledEditor(th, "Rolling average", &state.rollingAvgEditor, false)),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Flexed(1, labeledEditor(th, "JPEG quality", &state.jpegQualityEditor, jpegQualityDisabled)),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Flexed(1, formatSelector(th, state)),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

			// Start button and progress bar
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := "Start"
				if state.processing {
					text = "Processing..."
					if _, progressText := state.progress(); progressText != "" {
						text = progressText
					}
					gtx = gtx.Disabled()
				}
				return material.Button(th, &state.startBtn, text).Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				fraction, _ := state.progress()
				bar := material.ProgressBar(th, fraction)
				bar.Height = unit.Dp(8)
				return bar.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				displayText := state.statusText
				if state.processing {
					displayText = ""
				}
				return material.Body2(th, displayText).Layout(gtx)
			}),
		)
	})
}

func formatSelector(th *material.Theme, state *uiState) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(material.Caption(th, "Output format").Layout),
			layout.Rigid(material.RadioButton(th, &state.formatEnum, string(deflicker.FormatPng), "PNG").Layout),
			layout.Rigid(material.RadioButton(th, &state.formatEnum, string(deflicker.FormatJpeg), "JPEG").Layout),
		)
	}
}

func labeledEditor(th *material.Theme, label string, ed *widget.Editor, disabled bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		caption := material.Caption(th, label)
		if disabled {
			caption.Color = disabledColor(th)
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(caption.Layout),
			layout.Rigid(fullWidthBorderedEditor(th, ed, disabled)),
		)
	}
}

// disabledColor returns the theme foreground color at half opacity, used to
// visually gray out disabled controls.
func disabledColor(th *material.Theme) color.NRGBA {
	c := th.Fg
	c.A = c.A / 2
	return c
}

func fullWidthBorderedEditor(th *material.Theme, ed *widget.Editor, disabled bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		borderColor := th.Fg
		if disabled {
			gtx = gtx.Disabled()
			borderColor = disabledColor(th)
		}
		border := widget.Border{Color: borderColor, Width: unit.Dp(1), CornerRadius: unit.Dp(4)}
		return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			editorStyle := material.Editor(th, ed, "")
			if disabled {
				editorStyle.Color = disabledColor(th)
				editorStyle.HintColor = disabledColor(th)
			}
			return layout.UniformInset(unit.Dp(6)).Layout(gtx, editorStyle.Layout)
		})
	}
}
