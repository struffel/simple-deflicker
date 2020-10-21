package main

import (
	"fmt"

	"github.com/gosuri/uiprogress"
)

func createProgressBars() map[string]*uiprogress.Bar {
	progressBars := make(map[string]*uiprogress.Bar)

	progressBars["INITIALIZE"] = uiprogress.AddBar(len(pictures)).PrependCompleted().PrependElapsed()
	progressBars["ADJUST"] = uiprogress.AddBar(len(pictures)).PrependCompleted().PrependElapsed()

	progressBars["INITIALIZE"].Width = 20
	progressBars["ADJUST"].Width = 20

	progressBarFunction := func(b *uiprogress.Bar, step string) string {
		return fmt.Sprintf("%-15v %-5v/%-5v", step, b.Current(), b.Total)
	}

	progressBarFunctionAnalyze := func(b *uiprogress.Bar) string {
		return progressBarFunction(b, "Initializing")
	}

	progressBarFunctionAdjust := func(b *uiprogress.Bar) string {
		return progressBarFunction(b, "Adjusting")
	}

	progressBars["INITIALIZE"].AppendFunc(progressBarFunctionAnalyze)
	progressBars["ADJUST"].AppendFunc(progressBarFunctionAdjust)

	return progressBars
}
