gssh
----

Run ssh command on a group of servers simultaneously.
This project was inspired from *mpssh* and is written in Go. The reason for another mpssh fork is that it is a fun thing to do with Go.


Requirements
------------

In order to use gssh, the ssh binary from openssh package must be installed in user's path.
Also the machine running gssh should be able to connect to every server listed in the file with hosts without a password - either with a passwordless key or with ssh agent. 


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
	


A full list of currently supported arguments can be obtained with the -h option:

	gssh -h
	Usage of ./gssh:
	  -delay=10: delay between each ssh fork (default 10 msec)
	  -file="": file with the list of hosts
	  -procs=500: number of parallel ssh processes (default: 500)
	  -user="root": ssh login as this username

