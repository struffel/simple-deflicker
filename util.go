package main

import (
	"fmt"

	"github.com/gosuri/uiprogress"
	"github.com/inancgumus/screen"
)

func forEveryPicture(pictures []picture, progressBar *uiprogress.Bar, threads int, f func(pic picture) (picture, error)) ([]picture, error) {
	tokens := make(chan bool, threads)
	errors := make(chan error)
	for i := 0; i < threads; i++ {
		tokens <- true
	}
	for i := range pictures {
		select {
		case <-tokens:

		case err := <-errors:
			return pictures, err
		}
		go func(i int) {
			defer func() {
				progressBar.Incr()
				tokens <- true
			}()
			var functionError error
			pictures[i], functionError = f(pictures[i])
			if functionError != nil {
				errors <- functionError
			}
		}(i)
	}
	for i := 0; i < threads; i++ {
		_ = <-tokens
	}
	return pictures, nil
}
func printInfo() {
	fmt.Println("SIMPLE DEFLICKER")
	fmt.Println("v0.2.0 / github.com/StruffelProductions/simple-deflicker")
}
func clear() {
	screen.MoveTopLeft()
	screen.Clear()
}
