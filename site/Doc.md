Quick Setup
-------------

Production Deployment
==========

If you want to deply crew in production, I highly recommend that you put it under a reverse proxy, you know what I mean.

<i>Well, I also highly suggest you don't use it for something important, after all I didn't do any optimization at all (I certainly know about caching)</i>


In a hurry!
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
* All .md files in the root directory will be accessible, other files will be rendered as HTML file
* The same goes for subfolders. If you want to define the page of a folder, create an index.md (or index.html) in this folder, the default folder page will only print a simple directory tree
* For .md file, the default is to simply display the filename as the title in the navigation, but the _ will become a space, as in foo_bar.md -> foo bar. Of course, you can create a {filename}.conf.json in the same folder to reset the Title and Description,  e.g. [about.md.conf.json](https://github.com/c4pt0r/crew/blob/master/site/about.md.conf.json)
* You can put static files in $root/_static
* You can create a hidden directory by creating a `.conf.json` in this directory with content `{"hidden": true}`


Remote Render via JSON RPC(2.0)
=======


Create a config file like this:

```
{
    "type": "rpc",
    "rpc_endpoint": "http://localhost:5001/rpc"
}
```

Example: 

`python3 server.py` and then go [remote](./remote)

