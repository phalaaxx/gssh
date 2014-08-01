package main

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// load servers list from a file
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
