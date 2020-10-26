package main

import (
	"github.com/gosuri/uiprogress"
)

func clamp(a int, min int, max int) int {
	if a < min {
		return min
	}
	if a > max {
		return max
	}
	return a
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
