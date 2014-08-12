PREFIX = /usr/local
BINDIR = $(PREFIX)/bin
BINPREFIX =
INSTALL = /usr/bin/install -c
SRCS = gssh.go gssh_test.go serverlist.go sshgroup.go sshserver.go terminal.go
GCCGO = gccgo

gssh: $(SRCS)
	$(GCCGO) -o $@ $(SRCS)

all: gssh

install: all
	$(INSTALL) gssh $(BINDIR)/$(BINPREFIX)gssh

uninstall:
	rm $(BINDIR)/$(BINPREFIX)gssh
