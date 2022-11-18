package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/c4pt0r/log"
	"github.com/gomarkdown/markdown"
)

var (
	// rootDir is the root directory of the website.
	rootDir = flag.String("rootDir", "./site", "root directory")
	// _rootDir is the absolute path to the root directory
	_rootDir string

	// addr is the address to listen on.
	addr = flag.String("addr", ":8080", "address to listen on")
)

func init() {
	flag.Parse()
	var err error
	_rootDir, err = filepath.Abs(*rootDir)
	if err != nil {
		log.Fatal(err)
	}
}

type node struct {
	// filepath is the absolute path to the file
	filepath string
	title    string
	isDir    bool
}

func (n *node) walk() (dir []*node, files []*node, err error) {
	if !n.isDir {
		return nil, nil, fmt.Errorf("not a directory")
	}
	filepath.Walk(n.filepath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		node, err := newNodeFromPath(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			node.isDir = true
			dir = append(dir, node)
		} else {
			files = append(files, node)
		}
		return nil
	})
	return
}

func (n *node) URL() string {
	// get the relative path to the root directory
	if n.filepath == _rootDir {
		return "/"
	}
	// get the relative path
	relPath, err := filepath.Rel(_rootDir, n.filepath)
	if err != nil {
		log.Fatal(err)
	}
	// replace spaces with underscores
	relPath = strings.Replace(relPath, " ", "_", -1)
	// add the leading slash
	relPath = "/" + relPath
	return relPath
}

func (n *node) ext() string {
	return filepath.Ext(n.filepath)
}

func (n *node) render() ([]byte, error) {
	if n.isDir {
		return n.renderMarkdown()
	}
	if n.ext() == ".md" {
		return n.renderMarkdown()
	} else if n.ext() == ".html" {
		return n.renderHTML()
	} else {
		return nil, fmt.Errorf("unknown file extension %q", n.ext())
	}
}

func (n *node) renderMarkdown() ([]byte, error) {
	filePath := n.filepath
	if n.isDir {
		filePath = filepath.Join(filePath, "index.md")
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	// convert markdown to html
	output := markdown.ToHTML(content, nil, nil)
	return output, nil
}

func (n *node) renderHTML() ([]byte, error) {
	filePath := n.filepath
	if n.isDir {
		filePath = filepath.Join(filePath, "index.html")
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (n *node) String() string {
	if n.isDir {
		return fmt.Sprintf("%s [D]: %s", n.filepath, n.title)
	}
	return fmt.Sprintf("%s [F]: %s", n.filepath, n.title)
}

func newNodeFromPath(fullname string) (*node, error) {
	fpath, err := filepath.Abs(fullname)
	if err != nil {
		return nil, err
	}
	// check if is a directory
	info, err := os.Stat(fpath)
	if err != nil {
		return nil, err
	}
	fname := filepath.Base(fpath)
	title := strings.TrimSuffix(fname, filepath.Ext(fname))
	// replace underscores with spaces
	title = strings.Replace(title, "_", " ", -1)
	return &node{
		filepath: fpath,
		title:    title,
		isDir:    info.IsDir(),
	}, nil
}

func nodeFilter(nodes []*node, f func(*node) bool) []*node {
	var filtered []*node
	for _, node := range nodes {
		if f(node) {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

func httpServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// get the path from the request
		path := r.URL.Path
		// remove the leading slash
		path = path[1:]
		// get the node for the path
		fpath := filepath.Join(_rootDir, path)
		node, err := newNodeFromPath(fpath)
		log.I(node)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			} else {
				log.E(err)
				http.Error(w, "", http.StatusInternalServerError)
			}
			return
		}
		// render the node
		content, err := node.render()
		if err != nil {
			log.E(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// write the content to the response
		log.I(node.URL())
		w.Write(content)
	})
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func main() {
	httpServer()
}
