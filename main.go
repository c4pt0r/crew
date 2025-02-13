package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/template"

	"encoding/base64"

	"github.com/c4pt0r/log"
	"github.com/gomarkdown/markdown"
	lua "github.com/yuin/gopher-lua"
)

var (
	// rootDir is the root directory of the website.
	rootDir      = flag.String("rootDir", "./site", "root directory")
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
	<link rel="shortcut icon" href="/_static/favicon.ico" type="image/vnd.microsoft.icon">

	<link rel="stylesheet" href="/_static/highlight.js/default.min.css">
    <script src="/_static/highlight.js/highlight.min.js"></script>
    <script>hljs.initHighlightingOnLoad();</script>

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

// state is the global state for lua scripts
type state struct {
	m map[string]interface{}
	sync.RWMutex
}

func (s *state) Get(key string) interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.m[key]
}

func (s *state) Set(key string, value interface{}) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = value
}

func (s *state) Delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, key)
}

var (
	_state = &state{
		m: make(map[string]interface{}),
	}
)

func getRootNode() *node {
	return _rootNode
}

func init() {
	flag.Parse()
	var err error

	_rootDir = *rootDir
	_rootNode, err = newNodeFromPath(_rootDir)
	if err != nil {
		log.Fatal(err)
	}

	if *customPageTpl != "" {
		// read template file and replace pageTpl
		b, err := os.ReadFile(*customPageTpl)
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
)

func (ntp NodeType) String() string {
	switch ntp {
	case NodeTypeFile:
		return "file"
	default:
		return "unknown"
	}
}

func NodeTypeFromStr(s string) NodeType {
	switch s {
	case "file":
		return NodeTypeFile
	default:
		return NodeTypeFile
	}
}

type node struct {
	// filepath is the absolute path to the file
	filepath string
	// key is the key to the node in the database
	key string
	// rpcEndpoint is the endpoint to the rpc server
	rpcEndpoint string
	title       string
	desc        string
	isDir       bool
	isHidden    bool
	tp          NodeType
	authToken   string
	basicAuth   struct {
		username string
		password string
	}
}

type nodeConf struct {
	Title    string `json:"title'"`
	Desc     string `json:"desc"`
	IsHidden bool   `json:"hidden"`
	// Type is the type of the node, it can be "file"
	Tp string `json:"type"`
	// Key is the key to the node in the database if the node type is "kv", default value is the node URL
	Key string `json:"key"`
	// RpcEndpoint is the endpoint of the JsonRPC server if the node type is "rpc", default value is the node URL
	RpcEndpoint string `json:"rpc_endpoint"`
	// AuthToken is the token to access the node in header
	AuthToken string `json:"auth_token"`
	BasicAuth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"basic_auth"`
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

func isReservedName(name string) bool {
	if strings.HasPrefix(name, ".") ||
		strings.HasPrefix(name, "_") ||
		strings.HasSuffix(name, ".conf.json") ||
		name == "index.html" ||
		name == "index.md" {
		return true
	}
	return false
}

func (n *node) getSubNodes() ([]*node, error) {
	// get the files in the directory
	if !n.isDir {
		return nil, nil
	}
	files, err := os.ReadDir(n.filepath)
	if err != nil {
		return nil, err
	}
	// create the nodes
	var ns []*node
	for _, f := range files {
		// skip hidden/meta files
		if isReservedName(f.Name()) {
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
	if path.Clean(n.filepath) == path.Clean(_rootDir) {
		return nil, nil
	}
	parentDir := path.Dir(n.filepath)
	// create the node
	return newNodeFromPath(parentDir)
}

func (n *node) ext() string {
	return filepath.Ext(n.filepath)
}

func (n *node) Render(ctx context.Context) ([]byte, error) {
	if n.isDir {
		return n.renderDir(ctx)
	}
	switch n.ext() {
	case ".md":
		return n.renderMarkdown(ctx)
	case ".html":
		return n.renderHTML(ctx)
	case ".lua":
		return n.renderLua(ctx)
	default:
		return n.renderHTML(ctx)
	}
}

func (n *node) renderMarkdown(ctx context.Context) ([]byte, error) {
	filePath := n.filepath
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	// convert markdown to html
	output := markdown.ToHTML(content, nil, nil)
	return output, nil
}

func (n *node) renderHTML(ctx context.Context) ([]byte, error) {
	return n.rawContent()
}

func (n *node) rawContent() ([]byte, error) {
	// if it's directory, just return index file
	if n.isDir {
		n, err := getIndexNodeForDir(n.filepath)
		if err != nil {
			return nil, err
		}
		if n != nil {
			return n.rawContent()
		} else {
			return nil, fmt.Errorf("no index file (index.html or index.md) found for directory")
		}
	}
	filePath := n.filepath
	content, err := os.ReadFile(filePath)
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

func getIndexNodeForDir(dir string) (*node, error) {
	htmlIndex := path.Join(dir, "index.html")
	if fileExists(htmlIndex) {
		return newNodeFromPath(htmlIndex)
	}
	mdIndex := path.Join(dir, "index.md")
	if fileExists(mdIndex) {
		return newNodeFromPath(mdIndex)
	}
	return nil, nil
}

func (n *node) renderDir(ctx context.Context) ([]byte, error) {
	// if there's _index.md or _index.html, render that
	if indexNode, err := getIndexNodeForDir(n.filepath); err == nil && indexNode != nil {
		return indexNode.Render(ctx)
	}
	// get the sub nodes
	subNodes, err := n.getSubNodes()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteString("<h1>" + n.URL() + "</h1>")
	buf.WriteString("<ul>")
	for _, n := range subNodes {
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

func getConfigFileForFile(fpath string) (bool, string, error) {
	cfgPath := ""
	info, err := os.Stat(fpath)
	if err != nil {
		return false, "", err
	}
	if !info.IsDir() {
		dir, fn := path.Split(fpath)
		cfgPath = path.Join(dir, fn+".conf.json")
		return false, cfgPath, nil
	} else {
		cfgPath = path.Join(fpath, ".conf.json")
		return true, cfgPath, nil
	}
}

func newNodeFromPath(fullname string) (*node, error) {
	fpath := fullname
	// check if is a directory
	fname := filepath.Base(fpath)
	title := strings.TrimSuffix(fname, filepath.Ext(fname))
	// replace underscores with spaces
	title = strings.Replace(title, "_", " ", -1)
	// node desc
	desc := ""
	hidden := false
	tp := "file"
	key := ""
	rpcEndpoint := ""
	authToken := ""
	basicAuth := struct {
		username string
		password string
	}{}

	isDir, cfgPath, err := getConfigFileForFile(fpath)
	if err != nil {
		return nil, err
	}
	if fileExists(cfgPath) {
		data, err := os.ReadFile(cfgPath)
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
			if cfg.Tp == NodeTypeFile.String() && cfg.RpcEndpoint != "" {
				rpcEndpoint = cfg.RpcEndpoint
			}
		}
		if len(cfg.AuthToken) > 0 {
			authToken = cfg.AuthToken
		}
		if len(cfg.BasicAuth.Username) > 0 && len(cfg.BasicAuth.Password) > 0 {
			basicAuth.username = cfg.BasicAuth.Username
			basicAuth.password = cfg.BasicAuth.Password
		}
	}
	return &node{
		filepath:    fpath,
		title:       title,
		desc:        desc,
		isHidden:    hidden,
		isDir:       isDir,
		tp:          NodeTypeFromStr(tp),
		key:         key,
		rpcEndpoint: rpcEndpoint,
		authToken:   authToken,
		basicAuth:   basicAuth,
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
	bodyRender  func(p *page, ctx context.Context) ([]byte, error)
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
	p := pageFromNode(getRootNode())
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
	out, err := printList(getRootNode(), p.node)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

func (p *page) Render(ctx context.Context) ([]byte, error) {
	// if raw flag is set, just return the raw data
	if params, ok := ctx.Value("params").(map[string]string); ok {
		if v, ok := params["raw"]; ok && (v == "true" || v == "1") {
			return p.node.rawContent()
		}
	}
	tpl, err := template.New("page").Parse(pageTpl)
	if err != nil {
		return nil, err
	}

	// get the body
	var body []byte
	if p.bodyRender == nil {
		body, err = p.node.Render(ctx)
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
			if node.authToken != "" {
				// check http header got the auth token
				if v := r.Header.Get("Authorization"); len(v) > 0 {
					// split the Bearer token
					parts := strings.Split(v, " ")
					if !(len(parts) == 2 && parts[0] == "Bearer" && parts[1] == node.authToken) {
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
				} else {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}
			// Check for basic auth
			authNode := node
			for authNode != nil {
				if authNode.basicAuth.username != "" && authNode.basicAuth.password != "" {
					auth := r.Header.Get("Authorization")
					if !checkBasicAuth(auth, authNode.basicAuth.username, authNode.basicAuth.password) {
						w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
					break
				}
				authNode, _ = authNode.getParentNode()
			}
			// For .lua files with POST/PUT methods, handle directly
			if node.ext() == ".lua" && (r.Method == "POST" || r.Method == "PUT") {
				ctx := context.WithValue(
					context.WithValue(context.Background(), "params", getQueryParams(r)),
					"request",
					r,
				)
				content, err := node.Render(ctx)
				if err != nil {
					log.E(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Write(content)
				return
			}
			page = pageFromNode(node)
		}
		// Add request to context
		ctx := context.WithValue(
			context.WithValue(context.Background(), "params", getQueryParams(r)),
			"request",
			r,
		)
		content, err := page.Render(ctx)
		if err != nil {
			log.E(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(content)
	})
	log.I("Starting server on", addr)
	return http.ListenAndServe(addr, nil)
}

func checkBasicAuth(auth, username, password string) bool {
	if !strings.HasPrefix(auth, "Basic ") {
		return false
	}
	payload, _ := base64.StdEncoding.DecodeString(auth[6:])
	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		return false
	}
	return pair[0] == username && pair[1] == password
}

func (n *node) renderLua(ctx context.Context) ([]byte, error) {
	L := lua.NewState()
	defer L.Close()

	// Add globalState table
	stateTable := L.NewTable()

	// Add get method
	L.SetField(stateTable, "get", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := _state.Get(key)
		if value == nil {
			L.Push(lua.LNil)
			return 1
		}

		// Convert Go value to Lua value
		switch v := value.(type) {
		case string:
			L.Push(lua.LString(v))
		case int:
			L.Push(lua.LNumber(v))
		case float64:
			L.Push(lua.LNumber(v))
		case bool:
			L.Push(lua.LBool(v))
		default:
			L.Push(lua.LNil)
		}
		return 1
	}))

	// Add set method
	L.SetField(stateTable, "set", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.Get(2)

		// Convert Lua value to Go value
		var goValue interface{}
		switch value.Type() {
		case lua.LTString:
			goValue = string(value.(lua.LString))
		case lua.LTNumber:
			goValue = float64(value.(lua.LNumber))
		case lua.LTBool:
			goValue = bool(value.(lua.LBool))
		default:
			L.Push(lua.LBool(false))
			return 1
		}

		_state.Set(key, goValue)
		L.Push(lua.LBool(true))
		return 1
	}))

	// Add delete method
	L.SetField(stateTable, "delete", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		_state.Delete(key)
		return 0
	}))

	// Set globalState table as a global variable
	L.SetGlobal("globalState", stateTable)

	// Add createNode function
	L.SetGlobal("createNode", L.NewFunction(func(L *lua.LState) int {
		nodePath := L.CheckString(1)
		content := L.CheckString(2)

		// Get absolute path
		absPath := filepath.Join(_rootDir, nodePath)

		// Create parent directories if they don't exist
		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Write content to file
		if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	// Add readNode function
	L.SetGlobal("readNode", L.NewFunction(func(L *lua.LState) int {
		nodePath := L.CheckString(1)

		// Get absolute path
		absPath := filepath.Join(_rootDir, nodePath)

		// Read file content
		content, err := os.ReadFile(absPath)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(string(content)))
		return 1
	}))

	// Add removeNode function
	L.SetGlobal("removeNode", L.NewFunction(func(L *lua.LState) int {
		nodePath := L.CheckString(1)

		// Get absolute path
		absPath := filepath.Join(_rootDir, nodePath)

		// Check if it's a directory
		fileInfo, err := os.Stat(absPath)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Don't allow directory removal
		if fileInfo.IsDir() {
			L.Push(lua.LBool(false))
			L.Push(lua.LString("cannot remove directory"))
			return 2
		}

		// Remove the file
		err = os.Remove(absPath)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	// Create request table
	reqTable := L.NewTable()

	// Get request from context
	if r, ok := ctx.Value("request").(*http.Request); ok {
		// Add method
		L.SetField(reqTable, "method", lua.LString(r.Method))

		// Add path
		L.SetField(reqTable, "path", lua.LString(r.URL.Path))

		// Add params
		params := ctx.Value("params").(map[string]string)

		if r.Method == "POST" || r.Method == "PUT" {
			// If it's POST/PUT request, try to parse JSON body
			if r.Header.Get("Content-Type") == "application/json" {
				var jsonData map[string]interface{}
				decoder := json.NewDecoder(r.Body)
				if err := decoder.Decode(&jsonData); err == nil {
					// Add JSON data to params
					for k, v := range jsonData {
						if str, ok := v.(string); ok {
							params[k] = str
						} else {
							// Convert other types to string
							params[k] = fmt.Sprintf("%v", v)
						}
					}
				}
			}
			// If it's POST/PUT request, try to parse form data
			if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
				if err := r.ParseForm(); err == nil {
					for k, v := range r.PostForm {
						params[k] = v[0]
					}
				}
			}
		}

		// Create params table
		paramsTable := L.NewTable()
		for k, v := range params {
			L.SetField(paramsTable, k, lua.LString(v))
		}
		L.SetField(reqTable, "params", paramsTable)

		// Add query parameters
		queryTable := L.NewTable()
		for k, v := range r.URL.Query() {
			if len(v) > 0 {
				L.SetField(queryTable, k, lua.LString(v[0]))
			}
		}
		L.SetField(reqTable, "query", queryTable)

		// Add headers
		headerTable := L.NewTable()
		for k, v := range r.Header {
			if len(v) > 0 {
				L.SetField(headerTable, k, lua.LString(v[0]))
			}
		}
		L.SetField(reqTable, "headers", headerTable)

		// Set the request table as a global variable
		L.SetGlobal("request", reqTable)
	}

	content, err := n.rawContent()
	if err != nil {
		return nil, err
	}

	if err := L.DoString(string(content)); err != nil {
		return nil, fmt.Errorf("error executing lua file: %v", err)
	}

	// Get the appropriate function based on HTTP method
	var fnName string
	if r, ok := ctx.Value("request").(*http.Request); ok {
		switch r.Method {
		case "GET":
			fnName = "render"
		case "POST":
			fnName = "post"
		case "PUT":
			fnName = "put"
		default:
			fnName = "render"
		}
	} else {
		fnName = "render"
	}

	fn := L.GetGlobal(fnName)
	if fn.Type() != lua.LTFunction {
		// Fallback to render if method-specific function not found
		if fnName != "render" {
			fn = L.GetGlobal("render")
			if fn.Type() != lua.LTFunction {
				return nil, fmt.Errorf("render function not found in lua file")
			}
		} else {
			return nil, fmt.Errorf("%s function not found in lua file", fnName)
		}
	}

	// Call function with request table as parameter
	L.Push(fn)
	L.Push(reqTable)
	if err := L.PCall(1, 2, nil); err != nil { // Changed to expect 2 return values
		return nil, fmt.Errorf("error calling %s function: %v", fnName, err)
	}

	// Get status code and content
	statusCode := L.Get(-2) // First return value
	ret := L.Get(-1)        // Second return value
	L.Pop(2)

	if statusCode.Type() != lua.LTNumber {
		return nil, fmt.Errorf("%s function must return a number as first return value", fnName)
	}
	if ret.Type() != lua.LTString {
		return nil, fmt.Errorf("%s function must return a string as second return value", fnName)
	}

	// Set response status code in context
	if r, ok := ctx.Value("request").(*http.Request); ok {
		if w, ok := r.Context().Value("responseWriter").(http.ResponseWriter); ok {
			w.WriteHeader(int(lua.LVAsNumber(statusCode)))
		}
	}

	// check status code is 200
	if lua.LVAsNumber(statusCode) != 200 {
		return nil, fmt.Errorf("status code: %d msg: %s", lua.LVAsNumber(statusCode), ret.String())
	}

	return []byte(ret.String()), nil
}

func main() {
	if *printDefaultTpl {
		fmt.Print(pageTpl)
		return
	}
	log.Fatal(httpServer(*addr))
}
