gssh
----

Run ssh command on a group of servers simultaneously.
This project was inspired from *mpssh* and is written in Go. The reason for another mpssh fork is that it is a fun thing to do with Go.


Requirements
------------

In order to use gssh, the ssh binary from openssh package must be installed in user's path.
Also the machine running gssh should be able to connect to every server listed in the file with hosts without a password - either with a passwordless key or with ssh agent. 

Build
-----

To build gssh with official golang compiler, use the following command:

	go build
	./gssh -h

Another option is to use make to build with the sample makefile:

	make
	make install PREFIX=/usr

Usage
-----

A list of servers is mandatory to use gssh. The list is a plain text file with one server at a line (no username):

	cat << EOF > servers.txt
	server1.domain.tld
	server2.domain2.tld
	1.2.3.4
	EOF


To actually run a command on all files from the list:

	./gssh -file servers.txt 'uptime'


Alternative method to run gssh is to supply list of servers to standard input:

	cat << EOF | gssh 'uptime'
	server1.domain.tld
	server2.domain2.tld
	1.2.3.4
	EOF
	

Or to cat list files:

	cat servers.txt servers2.txt | gssh 'uname -r'


A full list of currently supported arguments can be obtained with the -h option:

	$ gssh -h
	Usage of gssh:
	  -d int
	        delay between each ssh fork (default 100 msec) (default 100)
	  -f string
	        file with the list of hosts
	  -h    show this help screen
	  -n    don't use strict ssh fingerprint checking
	  -p int
	        number of parallel ssh processes (default: 500) (default 500)
	  -s string
	        name of ini section containing servers list
	  -u string
	        ssh login as this username (default "root")

Options:

  * **d** - this is the time in miliseconds to wait between spawning next process
  * **f** - name of a text file containing list of servers; lines starting with # and empty lines are ignored
  * **p** - maximum number of processes to spawn
  * **n** - ask not to use strict fingerprint checking; default is to use strict checking
  * **u** - username to use for ssh login
  * **s** - name of ini-like section in input file under which is the list of servers to process
