Monsti CMS
==========

Monsti is a simple CMS designed to efficiently manage multiple small
websites.

Monsti is still under development and unstable. Don't expect any
backward compatibility at this point.

[![Build Status](https://travis-ci.org/monsti/monsti.svg?branch=master)](https://travis-ci.org/monsti/monsti)

Features
--------

 - Fast; thanks to Go, a statically typed compiled language
 - Low armortized (i.e. for many hosted sites) resource usage
 - No database required; configuration and data is stored in human
   readable files
 - Internationalization ready (Included languages: de, en)
 - Simple web frontend
 - Separation of code, configuration and presentation
 - Developer friendly: Includes a HTTPd; Go templates; high level API
   for node type and field creation and other common tasks
 - Administrator friendly: Syslog; init script; Makefile target for
   basic Debian packaging (via fpm, other distributions should be
   easy); respecting the filesystem hierarchy

Project
-------

Website: http://www.monsti.org/
Code: http://www.gitorious.org/monsti | http://github.com/monsti

Acknowledgments
---------------

 - [Silk Icons by Mark James](http://www.famfamfam.com/lab/icons/silk/)
