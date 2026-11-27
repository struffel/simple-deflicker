package progress

import (
	"fmt"
	"time"
)

type Updater interface {
	Start()
	Increment(msg string, completed int, ofTotal int)
	Finish()
}

// Default implementation for printing to the console.
type ConsoleUpdater struct {
	startTime time.Time
}

func (c *ConsoleUpdater) Start() {
	fmt.Println("Processing started...")
	c.startTime = time.Now()
}

func (c *ConsoleUpdater) Increment(msg string, completed int, ofTotal int) {
	fmt.Printf("[%d/%d] %s\n", completed, ofTotal, msg)
}

func (c *ConsoleUpdater) Finish() {
	fmt.Printf("Processing finished in %s\n", time.Since(c.startTime))
}
