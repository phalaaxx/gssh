package main

import (
	"github.com/pborman/getopt"
	"fmt"
	"log"
	"os"
	"time"
)


// main program
func main() {
	// local variables
	var err error

	// parse command line arguments
	OptUser := getopt.StringLong("user", 'u', "root", "ssh login as this username")
	OptFile := getopt.StringLong("file", 'f', "", "file with the list of hosts")
	OptDelay := getopt.IntLong("delay", 'd', 100, "delay between each ssh fork (default 100 msec)")
	OptProcs := getopt.IntLong("procs", 'p', 500, "number of parallel ssh processes (default: 500)")
	OptNoStrict := getopt.BoolLong("nostrict", 'n', "don't use strict ssh fingerprint checking")
	OptHelp := getopt.BoolLong("help", 'h', "show this help screen")
	getopt.Parse()

	// show help screen and exit in case of -h or --help option
	if *OptHelp {
		getopt.Usage()
		os.Exit(1)
	}

	// look for mandatory positional arguments
	if getopt.NArgs() < 1 {
		log.Fatal("Nothing to do. Use -h for help.")
	}

	// initialize output template strings
	Template := "%*s%s \033[01;32m->\033[0m %s"
	ErrTemplate := "%*s%s \033[01;31m=>\033[0m %s"

	// disable colored output in case output is redirected
	if !IsTerminal(os.Stdout.Fd()) {
		Template = "%*s%s -> %s"
	}
	if !IsTerminal(os.Stderr.Fd()) {
		ErrTemplate = "%*s%s => %s"
	}

	// by default, read server list from stdin
	ServerListFile := os.Stdin

	// read server names from file if a file name is supplied
	if *OptFile != "" {
		ServerListFile, err = os.Open(*OptFile)
		if err != nil {
			log.Fatal(fmt.Sprintf("ServerListFile: Error: %v", err))
		}
		defer ServerListFile.Close()
	}
	AddrPadding, ServerList := LoadServerList(ServerListFile)

	// command to run on servers
	OptCommand := getopt.Arg(0)

	// make new group
	group := &SshGroup{
		Active:   0,
		Total:    len(ServerList),
		Complete: 0,
	}

	// no point to display more processes than
	if *OptProcs > group.Total {
		*OptProcs = group.Total
	}

	// print heading text
	fmt.Fprintln(os.Stderr, "gssh - group ssh, ver. 0.6")
	fmt.Fprintln(os.Stderr, "(c)2014 Bozhin Zafirov <bozhin@deck17.com>")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "  [*] read (%d) hosts from the list\n", group.Total)
	fmt.Fprintf(os.Stderr, "  [*] executing '%s' as user '%s'\n", OptCommand, *OptUser)
	fmt.Fprintf(os.Stderr, "  [*] spawning %d parallel ssh sessions\n\n", *OptProcs)

	// spawn ssh processes
	for i, Server := range ServerList {
		ssh := &SshServer{
			Username: *OptUser,
			Address:  Server,
		}
		group.Servers = append(group.Servers, ssh)
		// run command
		group.mu.Lock()
		group.Active++
		group.mu.Unlock()
		go group.Command(ssh, AddrPadding, OptCommand, *OptNoStrict, Template, ErrTemplate)
		// show progress after new process spawn
		group.UpdateProgress()
		if i < group.Total {
			// time delay and max procs wait between spawn
			time.Sleep(time.Duration(*OptDelay) * time.Millisecond)
			group.Wait(*OptProcs)
		}
	}
	// wait for ssh processes to exit
	group.Wait(0)
	group.mu.Lock()
	group.ClearProgress()
	group.mu.Unlock()

	// calculate stats
	var StdoutServersCount int
	var StderrServersCount int
	var AllServersCount int
	var StdoutLinesCount int
	var StderrLinesCount int
	var AllLinesCount int
	for _, ssh := range group.Servers {
		if ssh.StdoutLineCount > 0 {
			StdoutLinesCount += ssh.StdoutLineCount
			StdoutServersCount++
		}
		if ssh.StderrLineCount > 0 {
			StderrLinesCount += ssh.StderrLineCount
			StderrServersCount++
		}
		if ssh.StdoutLineCount > 0 || ssh.StderrLineCount > 0 {
			AllLinesCount += ssh.StdoutLineCount + ssh.StderrLineCount
			AllServersCount++
		}
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "  Done. Processed: %d / Output: %d (%d) / \033[01;32m->\033[0m %d (%d) / \033[01;31m=>\033[0m %d (%d)\n",
		group.Total,
		AllServersCount,
		AllLinesCount,
		StdoutServersCount,
		StdoutLinesCount,
		StderrServersCount,
		StderrLinesCount,
	)
}
