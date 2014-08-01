package main

// ssh server information
type SshServer struct {
	Username        string
	Address         string
	StdoutLineCount int
	StderrLineCount int
}
