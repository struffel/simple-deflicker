package progress

import "fmt"

type Updater interface {
	Start()
	Increment(msg string, phase string, completed int, ofTotal int)
	Finish()
}

// Default implementation for printing to the console.
type ConsoleUpdater struct{}

func (c *ConsoleUpdater) Start() {
	fmt.Println("Processing started...")
}

func (c *ConsoleUpdater) Increment(msg string, phase string, completed int, ofTotal int) {
	fmt.Printf("%s: %s (%d/%d)\n", phase, msg, completed, ofTotal)
}

func (c *ConsoleUpdater) Finish() {
	fmt.Println("Processing finished.")
}
