package main

import (
	"fmt"
	"log"
	"os"
	"sync"
)

/* Message defines an output message for OutputMonitor */
type Message struct {
	Server string
	Data   string
	Stdout bool
}

/* OutputMonitor prints out stats and server data to screen */
func OutputMonitor(total int, padding int, srv *sync.WaitGroup) (chan Message, chan int) {
	var cntActive int
	var cntComplete int
	var Template string
	var OutDev *os.File

	message := make(chan Message)
	active := make(chan int)

	type OutputStats map[string]bool

	StdoutStats := make(OutputStats)
	StderrStats := make(OutputStats)

	PrintToTerminal := IsTerminal(os.Stdout.Fd())

	/* initialize output template strings */
	OutTemplate := "%*s%s \033[01;32m->\033[0m %s"
	ErrTemplate := "%*s%s \033[01;31m=>\033[0m %s"

	/* disable colored output in case output is redirected */
	if !PrintToTerminal {
		PrintToTerminal = false
		OutTemplate = "%*s%s -> %s"
	}
	if !IsTerminal(os.Stderr.Fd()) {
		ErrTemplate = "%*s%s => %s"
	}

	/* ClearProgress defines a local function to clear command progress */
	ClearProgress := func() {
		if _, err := fmt.Fprintf(os.Stderr, "\r%*s\r", 41, " "); err != nil {
			log.Println(err)
		}
	}
	PrintProgress := func() {
		if _, err := fmt.Fprintf(os.Stderr, "[%d/%d] %.2f%% complete, %d active",
			cntComplete,
			total,
			float64(cntComplete)*float64(100)/float64(total),
			cntActive,
		); err != nil {
			log.Println(err)
		}
	}

	/* statistics variables */
	var StdoutServersCount int
	var StderrServersCount int
	var StdoutLinesCount int
	var StderrLinesCount int

	OutputCallback := func(message chan Message, active chan int, srv *sync.WaitGroup) {
		defer srv.Done()
		for cntComplete != total {
			select {
			case msg := <-message:
				if msg.Stdout {
					Template = OutTemplate
					if _, ok := StdoutStats[msg.Server]; !ok {
						StdoutStats[msg.Server] = true
						StdoutServersCount++
					}
					StdoutLinesCount++
					OutDev = os.Stdout
				} else {
					Template = ErrTemplate
					if _, ok := StderrStats[msg.Server]; !ok {
						StderrStats[msg.Server] = true
						StderrServersCount++
					}
					StderrLinesCount++
					OutDev = os.Stderr
				}
				if PrintToTerminal {
					/* clear progress */
					ClearProgress()
				}
				if _, err := fmt.Fprintf(OutDev, Template, padding-len(msg.Server)+1, " ", msg.Server, msg.Data); err != nil {
					log.Println(err)
				}
				if PrintToTerminal {
					/* print progress */
					PrintProgress()
				}
			case cnt := <-active:
				/* update active count and exit if cntActive is zero */
				if cnt < 0 {
					cntComplete = cntComplete - cnt
				}
				cntActive = cntActive + cnt
				ClearProgress()
				PrintProgress()
			}
		}
		/* calculate and print end stats */
		ClearProgress()
		_, err := fmt.Fprintf(os.Stderr,
			"\n  Done. Processed: %d / Output: %d (%d) / \033[01;32m->\033[0m %d (%d) / \033[01;31m=>\033[0m %d (%d)\n",
			total,
			StdoutServersCount+StderrServersCount,
			StdoutLinesCount+StderrLinesCount,
			StdoutServersCount,
			StdoutLinesCount,
			StderrServersCount,
			StderrLinesCount,
		)
		if err != nil {
			log.Println(err)
		}
	}
	go OutputCallback(message, active, srv)

	return message, active
}
