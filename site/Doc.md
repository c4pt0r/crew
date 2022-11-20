Quick Setup
-------------

Production Deployment
==========

If you want to deply crew in production, I highly recommend that you put it under a reverse proxy, you know what I mean.

<i>Well, I also highly suggest you don't use it for something important, after all I didn't do any optimization at all (I certainly know about caching)</i>


In a hurry
===============

```
$ git clone https://github.com/c4pt0r/crew
$ go run main.go --addr $ADDR --rootDir $SITEDIR

or 

$ go run main.go (default dir is ./site, and default addr is 0.0.0.0:8080)
```

You're all set


How to use
============

Only few things you need to know (because it's suckless):

* You need to specify a directory as the root directory, by default it's `./site`
* All .md files in the root directory will be accessible
* The same goes for subfolders. If you want to define the page of a folder, create an _index.md in this folder, the default folder page will only print a simple directory tree
* For .md file, the default is to simply display the filename as the title in the navigation, but the _ will become a space, as in foo_bar.md -> foo bar. Of course, you can create a _filename.md.conf.json in the same folder to reset the Title and Description,  e.g. [_about.md.conf.json](https://github.com/c4pt0r/crew/blob/master/site/_about.md.conf.json)
* You can put static files in $root/_static

