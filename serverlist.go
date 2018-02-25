package main

import (
	"bufio"
	"io"
	"os"
	"strings"
)

/* ServerList defines a type for list of servers with sections */
type ServerList map[string][]string

/* Len returns the number of servers in the specified section */
func (s ServerList) Len(sectionName string) (count int) {
	for section := range s {
		if len(sectionName) == 0 || sectionName == section {
			count = count + len(s[section])
		}
	}
	return count
}

/* LoadServerList loads a list of server addresses from a file */
func LoadServerList(file *os.File) (AddrPadding int, servers ServerList) {
	servers = make(map[string][]string)
	AppendUnique := func(sectionList []string, Server string) []string {
		for _, S := range sectionList {
			if S == Server {
				return sectionList
			}
		}
		return append(sectionList, Server)
	}
	Reader := bufio.NewReader(file)
	section := "main"
	for Line, err := Reader.ReadString('\n'); err != io.EOF; Line, err = Reader.ReadString('\n') {
		SLine := strings.TrimSpace(Line)
		if strings.HasPrefix(SLine, "[") && strings.HasSuffix(SLine, "]") {
			section = SLine[1 : len(SLine)-1]
			continue
		}
		if SLine == "" || strings.HasPrefix(SLine, "#") {
			continue
		}
		if AddrPadding < len(SLine) {
			AddrPadding = len(SLine)
		}
		servers[section] = AppendUnique(servers[section], SLine)
	}
	return
}
