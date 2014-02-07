#!/bin/sh

# Might help: killall -r monsti-
rm run/*.socket
PATH=$PATH:../go/bin monsti-daemon config
