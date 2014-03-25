Installation instructions
=========================

1) Get the latest sources:
https://gitorious.org/monsti/monsti/

2) Install dependencies:

- make
- C compiler
- Git, Bazaar, Mercurial (to fetch modules and Go packages)
- MagickCore library and development files
- Latest Go compiler and tools

3) Build:

$ make

(Optional) To run tests:

$ make tests

4) Run:

$ go/bin/monsti-daemon <configuration directory>

To run the example site, go to the example site directory (example/)
and run the start script:
$ ./start.sh
Monsti will be listening on http://localhost:8080

5) Create packages to deploy (if you like):

tar:
$ make dist

Debian archive (using fpm, https://github.com/jordansissel/fpm):
$ make dist-deb
