package main

import (
	"fmt"

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

func forEveryPicture(pictures []picture, progressBar *uiprogress.Bar, threads int, f func(pic picture) picture) []picture {
	tokens := make(chan bool, threads)
	for i := 0; i < threads; i++ {
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
	for i := 0; i < threads; i++ {
		_ = <-tokens
	}
	return pictures
}
func printInfo() {
	fmt.Println("SIMPLE DEFLICKER")
	fmt.Println("v0.1.0 / github.com/StruffelProductions/simple-deflicker")
}
