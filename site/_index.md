~~werc~~ <span style="background-color:#FEDD00">crew</span> - A sane web anti-framework<span style="background-color:#FEDD00">, that suckless</span>
================================

~~Werc~~ <span style="background-color:#FEDD00">Crew</span> is a minimalist web anti-framework built following the [Unix](http://doc.cat-v.org/unix/) and [Plan 9](http://plan9.cat-v.org) _tool philosophy_ of software design. <span style="background-color:#FEDD00">Crew is a <a href="http://suckless.org">suckless</a> clone of werc, well, not a 100% clone...But I really like the idea of werc and its weird sense of humor.</span>

~~Werc~~<span style="background-color:#FEDD00">Crew</span> avoids the pain of managing collections of websites and developing web applications.

*   Database free, uses files and directories instead.
*   ~~Written using [the rc shell](http://rc.cat-v.org), leveraging the standard Unix/Plan 9 command toolkit.~~ <span style="background-color:#FEDD00">Written in Go, single file, which suckless, yo!</span>
*   Minimize tedious work: eg., no need to ever write HTML, use markdown (or any other format) instead.
*   Very minimalist yet extensible codebase: ~~highly functional core is 150 lines, with extra functionality in modular [apps](/apps/).~~ <span style="background-color:#FEDD00"> Checkout <a href="https://github.com/c4pt0r/crew/blob/master/main.go">main.go</a></span>

Features
--------

Here are some of the features provided by ~~werc~~<span style="background-color:#FEDD00">crew</span>:

*  <span style="background-color:#FEDD00">When you want to turn a directory into a website in a hurry :)</span>
*   Good integration with pre-existing content, you can add HTML or plain text files and they will be seamlessly integrated with the rest of the site.
*   You can use your favorite tools (text editor, shell, file manager, etc) to edit, manipulate and manage data stored in ~~werc~~<span style="background-color:#FEDD00">crew</span>.
*   ~~Designed to manage any number of ‘virtual’ domains that share a common style and layout from a single werc installation.~~<span style="background-color:#FEDD00">No, too complicated, single directory is easier</span>
*   ~~Configuration and customization can be at at any level: global, per-domain-group, domain-wide, directory sub-tree, and single file.~~<span style="background-color:#FEDD00">Configuration should be with the file and code, if you want to change the layout, just change the code and recompile</span>
*   Can trivially run multiple (customized) versions of ~~werc~~<span style="background-color:#FEDD00">crew</span> side by side. <span style="background-color:#FEDD00">Yeah, just launch another process</span>
*   Very simple and flexible user management and permissions system.<span style="background-color:#FEDD00">You should take good care of your own file system</span>
*   ~~Applications can be easily combined: eg., add comments to your blog or wiki by enabling the ‘bridge’ app; or by enabling the ‘diridir’ wiki convert any document tree into a wiki.~~<span style="background-color:#FEDD00">No, keep it simple</span>
*   ~~Can easily write werc ‘apps’ and extensions in _any_ language! (But of course, rc is recommended).~~<span style="background-color:#FEDD00">No, if you want to modify crew, you should understand what you're going to do, and recompile main.go</span>

Install Requirements
--------------------

~~All you need is some Plan 9 commands (cat, grep, sed, rc, etc.), and an HTTP server with CGI support.~~

~~Werc runs on any Unix-like system where [Plan 9 from User Space](https://9fans.github.io/plan9port/), [9base](https://tools.suckless.org/9base/), or [frontport](https://code.9front.org/hg/frontbase) are available (this includes Linux, \*BSD, OS X and Solaris), and on Plan 9.~~

~~Werc can use any HTTP server that can handle CGI, and has been tested with at least Apache, Lighttpd, Cherokee, nhttpd, Hiawatha, rc-httpd, cgd, and others.~~

<span style="background-color:#FEDD00">Basiclly any systems, but I've never tested crew on Plan9, I will, no, I won't</span>

~~Werc~~<span style="background-color:#FEDD00">Crew</span> uses markdown by default (and the standard ~~Perl~~ markdown <span style="background-color:#FEDD00">(gomarkdown/markdown)</span> is included with the distribution), to format documents, but any other formatting system can be used.

Source
------

To get a copy of the latest development code using mercurial, do:

       git clone https://github.com/c4pt0r/crew
    

You can also [browse the online repository](https://github.com/c4pt0r/crew).

Contact
-------

~~For questions, suggestions, bug reports and contributing patches you can join the werc mailing list. To join, send a message with a body consisting only of the word _subscribe_ to werc-owner@cat-v.org. After you get the confirmation notice, you can post by sending messages to werc@cat-v.org.~~

~~To track commit messages, you can join the werc-commits mailing list. To join, send a message with a body consisting only of the word _subscribe_ to werc-commits-owner@cat-v.org.~~

~~On irc, join [#cat-v](irc://irc.oftc.org/cat-v) on irc.oftc.org~~

<span style="background-color:#FEDD00">I don't use IRC :), just feel free to fire issue on Github, I'll take a look when I have time</span>

License
-------

Public domain, [because so called ‘intellectual property’ is an oxymoron](http://harmful.cat-v.org/economics/intellectual_property/).

Alternatively if your prefer it or your country’s brain dead copyright law doesn’t recognize the public domain werc is made available under the terms of the MIT and ISC licenses.

<span style="background-color:#FEDD00">crew uses MIT, yo! But I totally agree with what was said above ^^^</span>

Credits
-------

Thanks to [Uriel](http://uriel.cat-v.org/) for creating werc.

Thanks to Kris Maglione (aka JG) for implementing rss feeds, for writing the awk rc-templating system, and other help and inspiration (some parts of the code were based on JG’s diri wiki).

Thanks to Mechiel (aka oksel) for the md\_cache script.

Thanks Garbeam (aka arg) for writing the original diri code and showing that writing complex web apps in rc was feasible.

Thanks to Ethan Gardner for writing rc-httpd.

And thanks to everyone else whom we may have forgotten and that has provided fixes and feedback.

* * *

~~To post a comment you need to login first.~~ <span style="background-color:#FEDD00">No, you don't need it</span>
