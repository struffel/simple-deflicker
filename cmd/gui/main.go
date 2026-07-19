package main

import (
	"fmt"
	"os"

	"github.com/struffel/simple-deflicker/internal/ui"
)

func main() {
	if err := ui.StartGUI(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
