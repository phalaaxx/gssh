package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

/* ssh client group */
type SshGroup struct {
	/* mutex */
	stMu sync.RWMutex
	prMu sync.Mutex
	/* statistics */
	Active   int
	Total    int
	Complete int
}

/* determine if output device is terminal */
func IsTerminal(fd uintptr) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&termios)))
	return err == 0
}

/* wait until there are at most "n" (or none) processes left */
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

/* clear progress line */
func (s *SshGroup) ClearProgress() {
	s.prMu.Lock()
	fmt.Fprintf(os.Stderr, "\r%*s\r",
		27,
		" ")
	s.prMu.Unlock()
}

/* print progress line */
func (s *SshGroup) PrintProgress() {
	s.stMu.RLock()
	s.prMu.Lock()
	fmt.Fprintf(os.Stderr, "[%d/%d] %.2f%% complete",
		s.Complete,
		s.Total,
		float64(s.Complete)*float64(100)/float64(s.Total))
	s.prMu.Unlock()
	s.stMu.RUnlock()
}

/* connect to remote server */
func (s *SshGroup) Command(Username, Address string, AddrPadding int, Command string) {
	defer func() {
		s.stMu.Lock()
		s.Active--
		s.Complete++
		s.stMu.Unlock()
		s.ClearProgress()
		s.PrintProgress()
	}()

	/* hostkey checking from commandline arguments */
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
		"-l", Username,
		Address,
		Command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	/* padding length */
	Padding := AddrPadding - len(Address) + 1
	Stdout := bufio.NewReader(stdout)
	Stderr := bufio.NewReader(stderr)

	/* run the command */
	cmd.Start()

	var w sync.WaitGroup
	w.Add(2)

	PrintOutput := func(OutDev *os.File, Std *bufio.Reader, Template string) {
		for {
			line, err := Std.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			s.ClearProgress()
			s.prMu.Lock()
			/* write output to stdout */
			fmt.Fprintf(
				OutDev,
				Template,
				Padding,
				" ",
				Address,
				line)
			s.prMu.Unlock()
			s.PrintProgress()
		}
		w.Done()
	}

	go PrintOutput(os.Stdout, Stdout, Template)
	go PrintOutput(os.Stderr, Stderr, ErrTemplate)

	w.Wait()

}

/* load servers list from a file */
func LoadServerList(file *os.File) (AddrPadding int, ServerList []string) {
	AppendUniq := func(ServerList []string, Server string) []string {
		for _, S := range ServerList {
			if S == Server {
				return ServerList
			}
		}
		return append(ServerList, Server)
	}
	Reader := bufio.NewReader(file)
	for Line, err := Reader.ReadString('\n'); err != io.EOF; Line, err = Reader.ReadString('\n') {
		SLine := strings.TrimSpace(Line)
		if SLine == "" || strings.HasPrefix(SLine, "#") {
			continue
		}
		if AddrPadding < len(SLine) {
			AddrPadding = len(SLine)
		}
		ServerList = AppendUniq(ServerList, SLine)
	}
	return
}

/* global variables */
var Template = ""
var ErrTemplate = ""

/* commandline arguments */
var fCommand string
var fUser string
var fDelay int
var fProcs int
var fFile string
var fStrict bool

/* initialize */
func init() {
	/* commandline arguments */
	flag.StringVar(&fUser, "user", "root", "ssh login as this username")
	flag.StringVar(&fFile, "file", "", "file with the list of hosts")
	flag.IntVar(&fDelay, "delay", 10, "delay between each ssh fork (default 10 msec)")
	flag.IntVar(&fProcs, "procs", 500, "number of parallel ssh processes (default: 500)")
	flag.BoolVar(&fStrict, "strict", true, "strict ssh fingerprint checking")

	/* initialize output template strings */
	Template = "%*s%s \033[01;32m->\033[0m %s"
	ErrTemplate = "%*s%s \033[01;31m=>\033[0m %s"

	/* disable colored output in case output is redirected */
	if !IsTerminal(os.Stdout.Fd()) {
		Template = "%*s%s -> %s"
	}
	if !IsTerminal(os.Stderr.Fd()) {
		ErrTemplate = "%*s%s => %s"
	}
}

/* main program */
func main() {
	/* parse commandline argiments */
	flag.Parse()
	if flag.NArg() < 1 {
		log.Fatal("Missing command.")
	}

	/* sanity checks */
	if fFile == "" {
		log.Fatal("No serverlist file.")
	}

	fCommand = flag.Args()[0]
	/* read server names from file */
	file, err := os.Open(fFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	AddrPadding, ServerList := LoadServerList(file)

	/* make new group */
	ssh := &SshGroup{
		Active:   0,
		Total:    len(ServerList),
		Complete: 0,
	}

	/* no point to display more processes than  */
	if fProcs > ssh.Total {
		fProcs = ssh.Total
	}

	/* print heading text */
	fmt.Fprintln(os.Stderr, "gssh - group ssh, ver. 0.4")
	fmt.Fprintln(os.Stderr, "(c)2014 Bozhin Zafirov <bozhin@deck17.com>")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "  [*] read (%d) hosts from the list\n", ssh.Total)
	fmt.Fprintf(os.Stderr, "  [*] executing '%s' as user '%s'\n", fCommand, fUser)
	fmt.Fprintf(os.Stderr, "  [*] spawning %d parallel ssh sessions\n\n", fProcs)

	/* spawn ssh processes */
	for i, Server := range ServerList {
		/* run command */
		ssh.stMu.Lock()
		ssh.Active++
		ssh.stMu.Unlock()
		go ssh.Command(
			fUser,
			Server,
			AddrPadding,
			fCommand)
		/* show progless after new process spawn */
		ssh.ClearProgress()
		ssh.PrintProgress()
		if i < ssh.Total {
			/* time delay and max procs wait between spawn */
			time.Sleep(time.Duration(fDelay) * time.Millisecond)
			ssh.Wait(fProcs)
		}
	}
	/* wait for ssh processes to exit */
	ssh.Wait(0)
	ssh.ClearProgress()

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "  Done. %d hosts processed.\n", ssh.Total)
}
