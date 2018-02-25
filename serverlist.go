package main

import (
	"bufio"
	"io"
	"os"
	"strings"
)

/* LoadServerList loads a list of server addresses from a file */
func LoadServerList(file *os.File) (AddrPadding int, ServerList map[string][]string) {
	ServerList = make(map[string][]string)
	AppendUnique := func(ServerList []string, Server string) []string {
		for _, S := range ServerList {
			if S == Server {
				return ServerList
			}
		}
		return append(ServerList, Server)
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
		ServerList[section] = AppendUnique(ServerList[section], SLine)
	}
	return
}
