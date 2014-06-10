Monsti CMS
==========

Monsti is a simple CMS designed to efficiently manage multiple small
websites.

[![Build Status](https://travis-ci.org/monsti/monsti.svg?branch=master)](https://travis-ci.org/monsti/monsti)

Features
--------

 - Fast; thanks to Go, a statically typed compiled language
 - Low armortized (i.e. for many hosted sites) resource usage
 - No database required; configuration and data is stored in human
   readable files
 - Internationalization ready
 - Simple web frontend
 - Separation of code, configuration and presentation
 - Developer friendly: Includes a HTTPd; Go templates; use
   configuration files to build new node types
 - Administrator friendly: Syslog; init script; Makefile target for
   basic Debian packaging (via fpm, other distributions should be
   easy); respecting the filesystem hierarchy


Monsti is still under heavy development and unstable. You should not use it for
critical tasks.

Project
-------

Website: http://www.monsti.org/
Code: http://www.gitorious.org/monsti | http://github.com/monsti