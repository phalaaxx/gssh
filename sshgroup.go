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
	mu sync.Mutex
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
		s.mu.Lock()
		if s.Active == 0 || s.Active < n {
			s.mu.Unlock()
			break
		}
		s.mu.Unlock()
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
	fmt.Fprintf(os.Stderr, "[%d/%d] %.2f%% complete, %d active",
		s.Complete,
		s.Total,
		float64(s.Complete)*float64(100)/float64(s.Total),
		s.Active)
	os.Stderr.Sync()
}

// clear and reprint progress line
func (s *SshGroup) UpdateProgress() {
	s.mu.Lock()
	s.ClearProgress()
	s.PrintProgress()
	s.mu.Unlock()
}

// connect to remote server
func (s *SshGroup) Command(ssh *SshServer, AddrPadding int, Command string) {
	defer func() {
		s.mu.Lock()
		s.Active--
		s.Complete++
		s.mu.Unlock()
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
		log.Fatal(fmt.Sprintf("StdoutPipe: Error: %v", err))
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(fmt.Sprintf("StderrPipe: Error: %v", err))
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
				log.Fatal(fmt.Sprintf("PrintOutput: Error: %v", err))
			}

			s.mu.Lock()
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
			s.mu.Unlock()
		}
		w.Done()
	}

	go PrintOutput(os.Stdout, Stdout, Template, &ssh.StdoutLineCount)
	go PrintOutput(os.Stderr, Stderr, ErrTemplate, &ssh.StderrLineCount)

	w.Wait()
	cmd.Wait()
}
