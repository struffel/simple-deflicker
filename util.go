package main

import (
	"fmt"

	"github.com/gosuri/uiprogress"
	"github.com/inancgumus/screen"
)

func forEveryPicture(pictures []picture, progressBar *uiprogress.Bar, threads int, f func(pic picture) (picture, error)) ([]picture, error) {
	tokens := make(chan error, threads)
	var err error
	for i := 0; i < threads; i++ {
		tokens <- nil
	}
	for i := range pictures {
		err = <-tokens
		if err != nil {
			return pictures, err
		}
		go func(i int) {
			var functionError error
			defer func() {
				progressBar.Incr()
				tokens <- functionError
			}()
			pictures[i], functionError = f(pictures[i])
		}(i)
	}
	for i := 0; i < threads; i++ {
		err = <-tokens
		if err != nil {
			return pictures, err
		}
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
