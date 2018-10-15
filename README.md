Monsti CMS
==========

Monsti is a CMS designed to host multiple websites or blogs. It is
mainly designed for web projects like personal, small business, or
small NGO websites.

It provides a simple web frontend for basic site building and editing
tasks. More advanced tasks like adding new content types have to be
done by writing modules in Go that communicate to Monsti via RPC
using a high level API.

Monsti should not be considered rock stable. The API and architecture
still might change until the first stable release. But as it's already
in use to host some websites, migrations to new releases will be as
pleasant as possible and fully documented.

If you have a problem, please first have a look at Monsti's manual in
the `doc` directory or
[online](https://github.com/monsti/monsti/blob/master/doc/manual.adoc). Also
have a look at and search the [open
issues](https://github.com/monsti/monsti/issues). If you can't find
any help in the manual or open issues, you might [open a new
issue](https://github.com/monsti/monsti/issues/new) (preferred) or
contact the author: [c@2foo.net](mailto:c@2foo.net).

[![Build Status](https://travis-ci.org/monsti/monsti.svg?branch=master)](https://travis-ci.org/monsti/monsti)
[![Gratipay](https://img.shields.io/gratipay/chrneumann.svg)](https://gratipay.com/chrneumann/)

Features
--------

 - Fast; thanks to Go, a statically typed compiled language, and
   dependency based caching of pages, queries and calculations. Make
   your websites almost as fast as statically generated ones!
 - Low armortized (i.e. for many hosted sites) resource usage
 - No database system required; configuration and content is stored in
   human readable files. Xapian integration is planned for searching
   and indexing.
 - Internationalization ready (Included languages: de, en, nl).
 - Easy to use (albeit basic at the current stage of development) web
   frontend.
 - Separation of code, configuration and presentation.
 - Developer friendly: Includes a HTTPd; Go templates; high level API
   for node type and field creation and other common tasks.
 - Administrator friendly: Syslog; init script; Makefile target for
   basic Debian packaging (via fpm, other distributions should be
   easy); respecting the filesystem hierarchy

Project
-------

Website: http://www.monsti.org/
Code: http://www.gitorious.org/monsti | http://github.com/monsti

Contributions
-------------

Please note that this project is released with a
[Contributor Code of Conduct](CODE_OF_CONDUCT.md).
By participating in this project you agree to abide by its terms.

Acknowledgements
----------------

 - [Silk Icons by Mark James](http://www.famfamfam.com/lab/icons/silk/)
 - Influenced by Wordpress, Kotti, Plone, Drupal, and other great open
   source CMS.
