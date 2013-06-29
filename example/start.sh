#!/bin/sh

# killall monsti-httpd
rm run/*.socket
PATH=$PATH:../go/bin monsti-daemon config
