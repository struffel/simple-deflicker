package main

import (
	"fmt"
	"math"

	"github.com/gosuri/uiprogress"
)

func createProgressBars(numberOfPictures int) struct {
	analyze *uiprogress.Bar
	adjust  *uiprogress.Bar
} {
	var progressBars struct {
		analyze *uiprogress.Bar
		adjust  *uiprogress.Bar
	}
	progressBars.analyze = uiprogress.AddBar(numberOfPictures).PrependCompleted().PrependElapsed()
	progressBars.adjust = uiprogress.AddBar(numberOfPictures).PrependCompleted().PrependElapsed()

	progressBars.analyze.Width = 20
	progressBars.adjust.Width = 20

	progressBarFunction := func(b *uiprogress.Bar, step string) string {
		//Calculate the number of digits to display
		n := math.Floor(math.Log10(float64(b.Total)) + 1)
		f := fmt.Sprintf("%%-15v %%-%vv/%%-%vv", n, n)
		return fmt.Sprintf(f, step, b.Current(), b.Total)
	}

	progressBarFunctionAnalyze := func(b *uiprogress.Bar) string {
		return progressBarFunction(b, "Analyzing")
	}

	progressBarFunctionAdjust := func(b *uiprogress.Bar) string {
		return progressBarFunction(b, "Adjusting")
	}

	progressBars.analyze.AppendFunc(progressBarFunctionAnalyze)
	progressBars.adjust.AppendFunc(progressBarFunctionAdjust)

	return progressBars
}
