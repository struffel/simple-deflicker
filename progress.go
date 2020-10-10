package main

import (
	"path/filepath"

	"github.com/gosuri/uiprogress"
)

func createProgressBars() map[string]*uiprogress.Bar {
	progressBars := make(map[string]*uiprogress.Bar)

	progressBars["ANALYZE"] = uiprogress.AddBar(len(pictures)).PrependCompleted().PrependElapsed()
	progressBars["PREPARE"] = uiprogress.AddBar(len(pictures)).PrependCompleted().PrependElapsed()
	progressBars["ADJUST"] = uiprogress.AddBar(len(pictures)).PrependCompleted().PrependElapsed()

	progressBars["ANALYZE"].Width = 10
	progressBars["PREPARE"].Width = 10
	progressBars["ADJUST"].Width = 10

	progressBarFunctionAnalyze := func(b *uiprogress.Bar) string {
		if b.Current() == 0 {
			return "Analyzing"
		}
		return "Analyzing " + filepath.Base(pictures[b.Current()-1].path)
	}

	progressBarFunctionPrepare := func(b *uiprogress.Bar) string {
		if b.Current() == 0 {
			return "Preparing"
		}
		return "Preparing " + filepath.Base(pictures[b.Current()-1].path)
	}

	progressBarFunctionAdjust := func(b *uiprogress.Bar) string {
		if b.Current() == 0 {
			return "Adjusting"
		}
		return "Adjusting " + filepath.Base(pictures[b.Current()-1].path)
	}

	progressBars["ANALYZE"].AppendFunc(progressBarFunctionAnalyze)
	progressBars["PREPARE"].AppendFunc(progressBarFunctionPrepare)
	progressBars["ADJUST"].AppendFunc(progressBarFunctionAdjust)

	return progressBars
}
