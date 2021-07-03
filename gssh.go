package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

/* Global gssh version string */
var GsshVersion = `gssh - group ssh, ver. 2.0
(c)2014-2021 Bozhin Zafirov <bozhin@deck17.com>
`

// main program
func main() {
	// local variables
	var err error

	// parse command line arguments
	OptUser := flag.String("u", "root", "ssh login as this username")
	OptFile := flag.String("f", "", "file with the list of hosts")
	OptDelay := flag.Int("d", 100, "delay between each ssh fork (default 100 msec)")
	OptSection := flag.String("s", "", "name of ini section containing servers list")
	OptProcesses := flag.Int("p", 500, "number of parallel ssh processes (default: 500)")
	OptNoStrict := flag.Bool("n", false, "don't use strict ssh fingerprint checking")
	OptAnsible := flag.Bool("a", false, "Read ansible hosts file at /etc/ansible/hosts")
	OptVersion := flag.Bool("v", false, "Privt version and exit")
	OptHelp := flag.Bool("h", false, "show this help screen")
	flag.Parse()

	// show help screen and exit in case of -h or --help option
	if *OptHelp {
		flag.Usage()
		os.Exit(1)
	}

	/*  */
	if *OptVersion {
		if _, err = fmt.Fprintf(os.Stderr, GsshVersion); err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	// look for mandatory positional arguments
	if flag.NArg() < 1 {
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

	if *OptAnsible {
		*OptFile = "/etc/ansible/hosts"
	}

	// read server names from file if a file name is supplied
	if *OptFile != "" {
		ServerListFile, err = os.Open(*OptFile)
		if err != nil {
			log.Fatal(fmt.Sprintf("ServerListFile: Error: %v", err))
		}
		ServerListFileClose := func() {
			if err := ServerListFile.Close(); err != nil {
				log.Println(err)
			}
		}
		defer ServerListFileClose()
	}
	AddrPadding, servers := LoadServerList(ServerListFile)

	// command to run on servers
	OptCommand := flag.Arg(0)

	// make new group
	group := &SshGroup{
		Active:   0,
		Total:    servers.Len(*OptSection),
		Complete: 0,
	}

	// no point to display more processes than
	if *OptProcesses > group.Total {
		*OptProcesses = group.Total
	}

	// print heading text
	TemplateString := `%s
  [*] read (%d) hosts from the list
  [*] executing '%s' as user '%s'
  [*] spawning %d parallel ssh sessions

`
	if _, err = fmt.Fprintf(os.Stderr, TemplateString, GsshVersion, group.Total, OptCommand, *OptUser, *OptProcesses); err != nil {
		log.Println(err)
	}

	// spawn ssh processes
	for section := range servers {
		if len(*OptSection) != 0 && section != *OptSection {
			// skip current section
			continue
		}
		for i, Server := range servers[section] {
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
				// time delay and max processes wait between spawns
				time.Sleep(time.Duration(*OptDelay) * time.Millisecond)
				group.Wait(*OptProcesses)
			}
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

	_, err = fmt.Fprintf(os.Stderr,
		"\n  Done. Processed: %d / Output: %d (%d) / "+
			"\033[01;32m->\033[0m %d (%d) / \033[01;31m=>\033[0m %d (%d)\n",
		group.Total,
		AllServersCount,
		AllLinesCount,
		StdoutServersCount,
		StdoutLinesCount,
		StderrServersCount,
		StderrLinesCount,
	)
	if err != nil {
		log.Println(err)
	}
}
