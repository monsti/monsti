# Philosophy

## Private, small business, and small NGOs

Monsti contains features that are commonly needed by private, small
business, and small NGO web projects. It assumes that the content of
one of the hosted websites does not change very frequently and that
changes are seldomly performed by many users in parallel. Monsti also
assumes modest use of complex queries.

The web frontend provides all features needed for basic content
editing and site building tasks performed by end-users.

## Efficient (multiple) website hosting

It is possible to host any number of websites on a single monsti
instance without using many resources per website. Monsti's use of Go
and advanced caching system allow for small request times.

## Module system and high level API 

Monsti is designed to be installed, configured, and extended by a
technically adept person.

A developer can change and extend Monsti's core functionality using
modules. Using the high level API, it's e.g. possible to easily create
new node types.

## Opinionated

Features that are commonly needed for realizing common web projects
are contained in Monsti's core. The module system allows to change and
extend the core functionality, but there is be no need to decide for
and configure e.g. a WYSIWYG-editor.

## Software quality

The project aims to create a high quality CMS with a fair amount of
documentation, few bugs and security issues, stable releases, end-user
and developer friendlyness, and an ecosystem of free high quality
addons.

## Free software

Monsti is and will always be free software and encourage use of free
software. There is no need to use non-free software to use or extend
Monsti. The project will never promote non-free software, explicitly
including non-free modules.