package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/c4pt0r/log"
	"github.com/gomarkdown/markdown"
)

var (
	// rootDir is the root directory of the website.
	rootDir = flag.String("rootDir", "./site", "root directory")

	siteName     = flag.String("sitename", "crew", "site name")
	siteSubtitle = flag.String("site-subtitle", "Bringing more minimalism and sanity to the web, in a suckless way", "site name")

	// _rootDir is the absolute path to the root directory
	_rootDir string

	// addr is the address to listen on.
	addr = flag.String("addr", ":8080", "address to listen on")
)

const (
	pageTpl = `<!DOCTYPE html>
<html>
<head>
    <title>{{ .Title }}</title>
    <link rel="stylesheet" href="/_static/style.css" type="text/css" media="screen, handheld" title="default">
    <link rel="shortcut icon" href="/_static/favicon.ico" type="image/vnd.microsoft.icon">
    <meta charset="UTF-8">
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"> 
</head>
<body>

<header>
    <nav class="head-nav">
		<div class="left">
			<a href="http://quotes.cat-v.org">quotes</a> |
			<a href="http://doc.cat-v.org">docs</a> |
			<a href="http://repo.cat-v.org">repo</a> |
			<a href="http://go-lang.cat-v.org">golang</a> |
			<a href="http://sam.cat-v.org">sam</a> |
			<a href="http://man.cat-v.org">man</a> |
			<a href="http://acme.cat-v.org">acme</a> |
			<a href="http://glenda.cat-v.org">Glenda</a> |
			<a href="http://ninetimes.cat-v.org">9times</a> |
			<a href="http://harmful.cat-v.org">harmful</a> |
			<a href="http://9p.cat-v.org/">9P</a> |
			<a href="http://cat-v.org">cat-v.org</a>
		</div>

		<div class="right">
		  <span class="doNotDisplay">Related sites:</span>
		  | <a href="/sitemap">site map</a>
		</div>
    </nav>
    <h1><a href="/">{{ .Headline }} <span id="headerSubTitle">{{ .SubHeadline }}</span></a></h1>
</header>

<nav id="side-bar">
    <div>
		{{ .Nav }}
	</div>
</nav>

<article>
	{{ .Body }}
</article>

<footer>
<br class="doNotDisplay doNotPrint" />
<div style="margin-right: auto;"><a href="http://crew.0xffff.me">Powered by crew</a></div>
</footer>
</body></html>
`
)

var (
	// for render navbar & sitemap
	_rootNode *node
)

func init() {
	flag.Parse()
	var err error

	_rootDir = *rootDir
	_rootNode, err = newNodeFromPath(_rootDir)
	if err != nil {
		// this should never happen
		log.Fatal(err)
	}
}

type node struct {
	// filepath is the absolute path to the file
	filepath string
	title    string
	desc     string
	isDir    bool
}

type nodeConf struct {
	Title string `json:"title'"`
	Desc  string `json:"desc"`
}

func (n *node) URL() string {
	// get the relative path to the root directory
	if n.filepath == _rootDir {
		return "/"
	}
	// get the relative path
	relPath, err := filepath.Rel(_rootDir, n.filepath)
	if err != nil {
		// this should never happen
		log.Fatal(err)
	}
	// replace spaces with underscores
	relPath = strings.Replace(relPath, " ", "_", -1)
	// add the leading slash
	relPath = "/" + relPath
	return relPath
}

func (n *node) getSubNodes() ([]*node, error) {
	// get the files in the directory
	if !n.isDir {
		return nil, nil
	}
	files, err := ioutil.ReadDir(n.filepath)
	if err != nil {
		return nil, err
	}
	// create the nodes
	var ns []*node
	for _, f := range files {
		node, err := newNodeFromPath(path.Join(n.filepath, f.Name()))
		if err != nil {
			return nil, err
		}
		// skip hidden/meta files
		if strings.HasPrefix(f.Name(), "_") || strings.HasPrefix(f.Name(), ".") {
			continue
		}
		ns = append(ns, node)
	}

	// sort the nodes
	sortNodes(ns)
	return ns, nil
}

func sortNodes(ns []*node) {
	// sort the nodes, directories first, then files
	sort.Slice(ns, func(i, j int) bool {
		if ns[i].isDir && !ns[j].isDir {
			return true
		}
		return ns[i].title < ns[j].title
	})
}

func (n *node) getParentNode() (*node, error) {
	// get the parent directory
	parentDir := path.Dir(n.filepath)
	// create the node
	return newNodeFromPath(parentDir)
}

func (n *node) ext() string {
	return filepath.Ext(n.filepath)
}

func (n *node) render() ([]byte, error) {
	if n.isDir {
		return n.renderDir()
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
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func nodeTree(wr io.Writer, root *node, prefix string) {
	subnodes, _ := root.getSubNodes()

	if len(subnodes) > 0 {
		wr.Write([]byte(prefix + "<ul>"))
	}
	for _, n := range subnodes {
		if n.isDir {
			wr.Write([]byte("<li><a href=\"" + n.URL() + "/\">" + n.title + "/</a> " + n.desc + "</li>"))
		} else {
			wr.Write([]byte("<li><a href=\"" + n.URL() + "\">" + n.title + "</a> " + n.desc + "</li>"))
		}
		nodeTree(wr, n, prefix)
	}
	if len(subnodes) > 0 {
		wr.Write([]byte(prefix + "</ul>"))
	}
}

func nodesToHTML(ns []*node) []byte {
	var buf bytes.Buffer
	buf.WriteString("<ul>")
	for _, n := range ns {
		buf.WriteString("<li>")
		if n.isDir {
			buf.WriteString("<a href=\"" + n.URL() + "/\">" + n.title + "/</a> " + n.desc)
		} else {
			buf.WriteString("<a href=\"" + n.URL() + "\">" + n.title + "</a> " + n.desc)
		}
		buf.WriteString("</li>")
	}
	buf.WriteString("</ul>")
	return buf.Bytes()
}

func (n *node) renderDir() ([]byte, error) {
	// if there's _index.md or _index.html, render that
	indexFile := path.Join(n.filepath, "_index.md")
	if fileExists(indexFile) {
		indexNode, err := newNodeFromPath(indexFile)
		if err != nil {
			return nil, err
		}
		return indexNode.render()
	}
	// get the sub nodes
	subNodes, err := n.getSubNodes()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteString("<h1>" + n.URL() + "</h1>")
	buf.Write(nodesToHTML(subNodes))
	return buf.Bytes(), nil
}

func (n *node) String() string {
	if n.isDir {
		return fmt.Sprintf("%s [D]: %s", n.filepath, n.title)
	}
	return fmt.Sprintf("%s [F]: %s", n.filepath, n.title)
}

func fileExists(fpath string) bool {
	_, err := os.Stat(fpath)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		log.E(err)
		return false
	}
	return true
}

func newNodeFromPath(fullname string) (*node, error) {
	fpath := fullname
	// check if is a directory
	info, err := os.Stat(fpath)
	if err != nil {
		return nil, err
	}
	fname := filepath.Base(fpath)
	title := strings.TrimSuffix(fname, filepath.Ext(fname))
	// replace underscores with spaces
	title = strings.Replace(title, "_", " ", -1)
	// node desc
	desc := ""
	// if there's a config file, load config
	if !info.IsDir() {
		dir, fn := path.Split(fpath)
		cfgPath := path.Join(dir, "_"+fn+".conf.json")
		if fileExists(cfgPath) {
			data, err := ioutil.ReadFile(cfgPath)
			if err != nil {
				return nil, err
			}
			var cfg nodeConf
			err = json.Unmarshal(data, &cfg)
			if err != nil {
				return nil, err
			}
			if len(cfg.Title) > 0 {
				title = cfg.Title
			}
			if len(cfg.Desc) > 0 {
				desc = cfg.Desc
			}
		}
	} else {
		// TODO dir config
	}

	return &node{
		filepath: fpath,
		title:    title,
		desc:     desc,
		isDir:    info.IsDir(),
	}, nil
}

// +-Title------------------+
// | Headerline SubHeadline |
// +------------------------+
// |   |                    |
// | N |                    |
// | a |      Body          |
// | v |                    |
// |   |                    |
// +------------------------|
// | Footer                 |
// +------------------------+
type page struct {
	node *node
	tp   string
	// for the template
	Header      string
	Headline    string
	SubHeadline string
	Footer      string
	Nav         string
	Body        string
	Title       string
	Vals        map[string]string

	// for different type
	bodyRender func(p *page, params map[string]string) ([]byte, error)
}

func pageFromNode(n *node) *page {
	p := &page{
		node:        n,
		Headline:    *siteName,
		SubHeadline: *siteSubtitle,
	}
	p.Title = n.title
	return p
}

func sitemapPage() *page {
	p := pageFromNode(_rootNode)
	p.bodyRender = func(p *page, params map[string]string) ([]byte, error) {
		var buf bytes.Buffer
		buf.WriteString("<h1> Site map </h1>")
		nodeTree(&buf, p.node, "")
		return buf.Bytes(), nil
	}
	return p
}

func filterNode(ns []*node, f func(*node) bool) []*node {
	var filtered []*node
	for _, n := range ns {
		if f(n) {
			filtered = append(filtered, n)
		}
	}
	return filtered
}

func i(text string) string {
	return "<i>" + text + "</i>"
}

func b(text string) string {
	return "<b>" + text + "</b>"
}

func printList(from *node, to *node) (string, error) {
	var buf bytes.Buffer
	buf.WriteString("<ul>")
	subnodes, err := from.getSubNodes()
	if err != nil {
		return "", err
	}
	for _, n := range subnodes {
		buf.WriteString("<li>")
		title := n.title
		if n.isDir {
			if strings.HasPrefix(to.filepath, n.filepath) {
				title = i(title)
				if n.filepath == to.filepath {
					title = b(title)
				}
				title = "» " + title
			} else {
				title = "› " + title
			}
			buf.WriteString("<a href=\"" + n.URL() + "\">" + title + "/</a>")
		} else {
			title = "› " + title
			if n.filepath == to.filepath {
				title = "» " + n.title
				title = b(title)
			}
			buf.WriteString("<a href=\"" + n.URL() + "\">" + title + "</a>")
		}

		if n.isDir && strings.HasPrefix(to.filepath, n.filepath) {
			buf.WriteString("<ul>")
			out, err := printList(n, to)
			if err != nil {
				return "", err
			}
			buf.WriteString(out)
			buf.WriteString("</ul>")
		}

		buf.WriteString("</li>")
	}
	buf.WriteString("</ul>")
	return buf.String(), nil
}

func (p *page) renderNav() ([]byte, error) {
	out, err := printList(_rootNode, p.node)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

func (p *page) render() ([]byte, error) {
	tpl, err := template.New("page").Parse(pageTpl)
	if err != nil {
		return nil, err
	}

	// get the body
	var body []byte
	if p.bodyRender == nil {
		body, err = p.node.render()
		if err != nil {
			return nil, err
		}
	} else {
		body, err = p.bodyRender(p, nil)
		if err != nil {
			return nil, err
		}
	}
	p.Body = string(body)
	// get nav
	nav, err := p.renderNav()
	if err != nil {
		return nil, err
	}
	p.Nav = string(nav)

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func serverStatic(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to the file
	filepath := filepath.Join(_rootDir, r.URL.Path)
	// check if the file exists
	if fi, err := os.Stat(filepath); (err == nil && fi.IsDir()) || os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}
	// serve the file
	http.ServeFile(w, r, filepath)
}

func httpServer(addr string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// get the path from the request, and remove the leading slash
		var page *page
		log.Infof("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		path := r.URL.Path[1:]
		if strings.HasPrefix(path, "_static") {
			serverStatic(w, r)
			return
		} else if strings.HasPrefix(path, "sitemap") {
			// site map
			page = sitemapPage()
		} else {
			// get the node for the path
			fpath := filepath.Join(_rootDir, path)
			node, err := newNodeFromPath(fpath)
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
			if err != nil {
				log.E(err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			// render the node
			page = pageFromNode(node)
		}
		content, err := page.render()
		if err != nil {
			log.E(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// write the content to the response
		w.Write([]byte(content))
	})
	return http.ListenAndServe(addr, nil)
}

func main() {
	log.Fatal(httpServer(*addr))
}
