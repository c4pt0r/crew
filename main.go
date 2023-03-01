package main

import (
	"bytes"
	"context"
	"database/sql"
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
	_ "github.com/mattn/go-sqlite3"
)

var (
	// rootDir is the root directory of the website.
	rootDir = flag.String("rootDir", "./site", "root directory")
	// sqlitePath
	sqlitePath   = flag.String("storage", "./.site.db", "sqlite path")
	siteName     = flag.String("sitename", "crew", "site name")
	siteSubtitle = flag.String("site-subtitle", "Bringing more minimalism and sanity to the web, in a suckless way", "site name")
	// customPageTpl is the path to a custom page template.
	customPageTpl   = flag.String("page-tpl", "", "custom page template file, use -print-page-tpl to print the default template")
	printDefaultTpl = flag.Bool("print-default-page-template", false, "print the default page template")
	// _rootDir is the absolute path to the root directory
	_rootDir string
	// addr is the address to listen on.
	addr = flag.String("addr", ":8080", "address to listen on")
)

var (
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
		log.Fatal(err)
	}

	// create the database
	_globalStorage, err = newSqliteStorage(*sqlitePath)
	if err != nil {
		log.Fatal(err)
	}

	if *customPageTpl != "" {
		// read template file and replace pageTpl
		b, err := ioutil.ReadFile(*customPageTpl)
		if err != nil {
			log.Fatal(err)
		}
		pageTpl = string(b)
	}
}

type NodeType int

const (
	// NodeTypeFile is a file node.
	NodeTypeFile NodeType = iota
	NodeTypeKV
	NodeTypeRPC
)

func (ntp NodeType) String() string {
	switch ntp {
	case NodeTypeFile:
		return "file"
	case NodeTypeKV:
		return "kv"
	case NodeTypeRPC:
		return "rpc"
	default:
		return "unknown"
	}
}

func NodeTypeFromStr(s string) NodeType {
	switch s {
	case "file":
		return NodeTypeFile
	case "kv":
		return NodeTypeKV
	case "rpc":
		return NodeTypeRPC
	default:
		return NodeTypeFile
	}
}

type node struct {
	// filepath is the absolute path to the file
	filepath string
	// key is the key to the node in the database
	key      string
	title    string
	desc     string
	isDir    bool
	isHidden bool
	tp       NodeType
}

type nodeConf struct {
	Title    string `json:"title'"`
	Desc     string `json:"desc"`
	IsHidden bool   `json:"hidden"`
	// Type is the type of the node, it can be "file" or "kv"
	Tp string `json:"type"`
	// Key is the key to the node in the database if the node type is "kv", default value is the node URL
	Key string `json:"key"`
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
		// skip hidden/meta files
		if strings.HasPrefix(f.Name(), "_") || strings.HasPrefix(f.Name(), ".") {
			continue
		}
		node, err := newNodeFromPath(path.Join(n.filepath, f.Name()))
		if err != nil {
			return nil, err
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

func (n *node) render(ctx context.Context) ([]byte, error) {
	if n.isDir {
		return n.renderDir(ctx)
	}
	if n.ext() == ".md" {
		return n.renderMarkdown(ctx)
	} else if n.ext() == ".html" {
		return n.renderHTML(ctx)
	} else {
		return nil, fmt.Errorf("unknown file extension %q", n.ext())
	}
}

func (n *node) renderMarkdown(ctx context.Context) ([]byte, error) {
	filePath := n.filepath
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	// convert markdown to html
	output := markdown.ToHTML(content, nil, nil)
	return output, nil
}

func (n *node) renderHTML(ctx context.Context) ([]byte, error) {
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
		if n.isHidden {
			continue
		}
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
		if n.isHidden {
			continue
		}
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

func (n *node) renderDir(ctx context.Context) ([]byte, error) {
	// if there's _index.md or _index.html, render that
	indexFile := path.Join(n.filepath, "_index.md")
	if fileExists(indexFile) {
		indexNode, err := newNodeFromPath(indexFile)
		if err != nil {
			return nil, err
		}
		return indexNode.render(ctx)
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
	cfgPath := ""
	hidden := false
	tp := "file"
	key := ""
	// if there's a config file, load config
	if !info.IsDir() {
		dir, fn := path.Split(fpath)
		cfgPath = path.Join(dir, "_"+fn+".conf.json")

	} else {
		// read the config file
		cfgPath = path.Join(fpath, "_.conf.json")
	}
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
		if cfg.IsHidden {
			hidden = true
		}
		if len(cfg.Tp) > 0 {
			tp = cfg.Tp
			if cfg.Tp == "kv" && cfg.Key != "" {
				key = cfg.Key
			}
		}
	}

	return &node{
		filepath: fpath,
		title:    title,
		desc:     desc,
		isHidden: hidden,
		isDir:    info.IsDir(),
		tp:       NodeTypeFromStr(tp),
		key:      key,
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
	bodyRender func(p *page, ctx context.Context) ([]byte, error)
}

type Storage interface {
	Get(key string) ([]byte, error)
	Put(key string, val []byte) error
	Del(key string) error
}

// SqliteStorage is a storage that uses sqlite as backend
type SqliteStorage struct {
	db *sql.DB
}

func newSqliteStorage(dbPath string) (Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS storage (key TEXT PRIMARY KEY, val TEXT)")
	if err != nil {
		return nil, err
	}
	return &SqliteStorage{
		db: db,
	}, nil
}

func (s *SqliteStorage) Get(key string) ([]byte, error) {
	var val string
	err := s.db.QueryRow("SELECT val FROM storage WHERE key = ?", key).Scan(&val)
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

func (s *SqliteStorage) Put(key string, val []byte) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO storage (key, val) VALUES (?, ?)", key, string(val))
	return err
}

func (s *SqliteStorage) Del(key string) error {
	_, err := s.db.Exec("DELETE FROM storage WHERE key = ?", key)
	return err
}

var _globalStorage Storage

func getStorage() Storage {
	return _globalStorage

}

func pageFromNode(n *node) *page {
	p := &page{
		node:        n,
		Headline:    *siteName,
		SubHeadline: *siteSubtitle,
	}
	p.Title = n.title
	if p.node.tp == NodeTypeKV {
		p.bodyRender = func(p *page, ctx context.Context) ([]byte, error) {
			key := n.URL()
			if len(n.key) > 0 {
				key = n.key
			}
			log.I("get kv store for key:", key)
			v, err := getStorage().Get(key)
			if err != nil {
				return []byte("error: " + err.Error()), nil
			}
			return v, nil
		}
	}
	return p
}

func sitemapPage() *page {
	p := pageFromNode(_rootNode)
	p.bodyRender = func(p *page, ctx context.Context) ([]byte, error) {
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
		if n.isHidden {
			continue
		}
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

func (p *page) render(ctx context.Context) ([]byte, error) {
	tpl, err := template.New("page").Parse(pageTpl)
	if err != nil {
		return nil, err
	}

	// get the body
	var body []byte
	if p.bodyRender == nil {
		body, err = p.node.render(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		body, err = p.bodyRender(p, ctx)
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

// get query params from http.Request
func getQueryParams(r *http.Request) map[string]string {
	params := make(map[string]string)
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}
	return params
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
		ctx := context.WithValue(context.Background(), "params", getQueryParams(r))
		content, err := page.render(ctx)
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
	if *printDefaultTpl {
		fmt.Print(pageTpl)
		return
	}
	log.Fatal(httpServer(*addr))
}
