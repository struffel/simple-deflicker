package main

import (
	"fmt"

	"github.com/gosuri/uiprogress"
)

func minimum(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maximum(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(a int, min int, max int) int {
	if a < min {
		return min
	}
	if a > max {
		return max
	}
	return a
}
func formatHistogram(lut [256]uint8) string {
	output := ""
	for i, v := range lut {
		output += fmt.Sprintf("%v: %v\n", i, v)
	}
	return output
}

func forEveryPicture(pictures []picture, progressBar *uiprogress.Bar, f func(pic picture) picture) []picture {
	tokens := make(chan bool, config.threads)
	for i := 0; i < config.threads; i++ {
		tokens <- true
	}
	for i := range pictures {
		_ = <-tokens
		go func(i int) {
			defer func() {
				progressBar.Incr()
				tokens <- true
			}()
			pictures[i] = f(pictures[i])
		}(i)
	}
	for i := 0; i < config.threads; i++ {
		_ = <-tokens
	}
	return pictures
}
