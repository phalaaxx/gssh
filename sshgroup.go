package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

// ssh client group
type SshGroup struct {
	// mutex
	stMu sync.RWMutex
	prMu sync.Mutex
	// statistics
	Active   int
	Total    int
	Complete int
	// server data
	Servers []*SshServer
}

// wait until there are at most "n" (or none) processes left
func (s *SshGroup) Wait(n int) {
	for {
		s.stMu.RLock()
		if s.Active == 0 || s.Active < n {
			s.stMu.RUnlock()
			break
		}
		s.stMu.RUnlock()
		time.Sleep(100 * time.Millisecond)
	}
}

// clear progress line
func (s *SshGroup) ClearProgress() {
	fmt.Fprintf(os.Stderr, "\r%*s\r",
		41,
		" ")
}

// print progress line
func (s *SshGroup) PrintProgress() {
	s.stMu.RLock()
	fmt.Fprintf(os.Stderr, "[%d/%d] %.2f%% complete, %d active",
		s.Complete,
		s.Total,
		float64(s.Complete)*float64(100)/float64(s.Total),
		s.Active)
	s.stMu.RUnlock()
}

// clear and reprint progress line
func (s *SshGroup) UpdateProgress() {
	s.prMu.Lock()
	s.ClearProgress()
	s.PrintProgress()
	s.prMu.Unlock()
}

// connect to remote server
func (s *SshGroup) Command(ssh *SshServer, AddrPadding int, Command string) {
	defer func() {
		s.stMu.Lock()
		s.Active--
		s.Complete++
		s.stMu.Unlock()
		s.UpdateProgress()
	}()

	// hostkey checking from commandline arguments
	StrictHostKeyChecking := "StrictHostKeyChecking=yes"
	if !fStrict {
		StrictHostKeyChecking = "StrictHostKeyChecking=no"
	}

	cmd := exec.Command("env",
		"ssh",
		"-A",
		"-o", "PasswordAuthentication=no",
		"-o", StrictHostKeyChecking,
		"-o", "GSSAPIAuthentication=no",
		"-o", "HostbasedAuthentication=no",
		"-l", ssh.Username,
		ssh.Address,
		Command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	// padding length
	Padding := AddrPadding - len(ssh.Address) + 1
	Stdout := bufio.NewReader(stdout)
	Stderr := bufio.NewReader(stderr)

	// run the command
	cmd.Start()

	var w sync.WaitGroup
	w.Add(2)

	PrintOutput := func(OutDev *os.File, Std *bufio.Reader, Template string, LineCount *int) {
		PrintToTerminal := IsTerminal(OutDev.Fd())
		for {
			line, err := Std.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			s.prMu.Lock()
			if PrintToTerminal {
				s.ClearProgress()
			}
			// print output
			fmt.Fprintf(
				OutDev,
				Template,
				Padding,
				" ",
				ssh.Address,
				line)
			if PrintToTerminal {
				s.PrintProgress()
			}
			*LineCount++
			s.prMu.Unlock()
		}
		w.Done()
	}

	go PrintOutput(os.Stdout, Stdout, Template, &ssh.StdoutLineCount)
	go PrintOutput(os.Stderr, Stderr, ErrTemplate, &ssh.StderrLineCount)

	w.Wait()
	cmd.Wait()
}
