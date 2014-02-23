gssh
----

Run ssh command on a group of servers simultaneously.


Requirements
------------

In order to use this gssh, the ssh binary from openssh package must be installed in user's path.


Usage
-----

To build gssh use the following command:

	go build gssh.go
	./gssh -h


A list of servers is mandatory to use gssh. The list is a plain text file with one server at a line (no username):

	cat << EOF > servers.txt
	server1.domain.tld
	server2.domain2.tld
	1.2.3.4
	EOF


To actually run a command on all files from the list:

	./gssh -file servers.txt "uptime"
