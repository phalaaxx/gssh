PREFIX = /usr/local
BINDIR = $(PREFIX)/bin
BINPREFIX =
INSTALL = /usr/bin/install -c
SRCS = gssh.go output.go serverlist.go sshgroup.go terminal.go
GO = go

gssh: $(SRCS)
	$(GO) build -o $@ $(SRCS)

all: gssh

install: all
	$(INSTALL) gssh $(BINDIR)/$(BINPREFIX)gssh

uninstall:
	rm $(BINDIR)/$(BINPREFIX)gssh
