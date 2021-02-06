package main

import (
	"fmt"
	"math"

	"github.com/gosuri/uiprogress"
)

type progressInfo struct {
	container *uiprogress.Progress
	bars      map[string]*uiprogress.Bar
}

func createProgressBars(numberOfPictures int) progressInfo {
	var tmpProgress progressInfo
	tmpProgress.container = uiprogress.New()
	tmpProgress.bars = make(map[string]*uiprogress.Bar)
	tmpProgress.bars["analyze"] = tmpProgress.container.AddBar(numberOfPictures).PrependCompleted().PrependElapsed()
	tmpProgress.bars["adjust"] = tmpProgress.container.AddBar(numberOfPictures).PrependCompleted().PrependElapsed()

	tmpProgress.bars["analyze"].Width = 20
	tmpProgress.bars["adjust"].Width = 20

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

	tmpProgress.bars["adjust"].AppendFunc(progressBarFunctionAnalyze)
	tmpProgress.bars["analyze"].AppendFunc(progressBarFunctionAdjust)

	return tmpProgress
}
