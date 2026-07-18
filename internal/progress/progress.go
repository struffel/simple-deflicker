package progress

import "fmt"

type Updater interface {
	Start()
	Increment(msg string, phase string, completed int, ofTotal int)
	Finish()
}

// NullUpdater is a dummy implementation of the Updater interface that does nothing.
type NullUpdater struct{}

func (n *NullUpdater) Start()                                                         {}
func (n *NullUpdater) Increment(msg string, phase string, completed int, ofTotal int) {}
func (n *NullUpdater) Finish()                                                        {}

// ConsoleUpdater is an implementation of the Updater interface that prints progress to the console.
type ConsoleUpdater struct{}

func (c *ConsoleUpdater) Start() {
	println("Processing started...")
}

func (c *ConsoleUpdater) Increment(msg string, phase string, completed int, ofTotal int) {
	fmt.Printf("[%s] %s (%d/%d)\n", phase, msg, completed, ofTotal)
}

func (c *ConsoleUpdater) Finish() {
	println("Processing finished.")
}
