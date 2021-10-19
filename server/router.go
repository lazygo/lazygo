package server

import (
	"reflect"
	"strings"
)

type (
	// Router is the registry of all registered routes for an `Server` instance for
	// request matching and URL path parameter parsing.
	Router struct {
		tree   *node
		server *Server
	}
	kind          uint8
	methodHandler struct {
		connect  HandlerFunc
		delete   HandlerFunc
		get      HandlerFunc
		head     HandlerFunc
		options  HandlerFunc
		patch    HandlerFunc
		post     HandlerFunc
		propfind HandlerFunc
		put      HandlerFunc
		trace    HandlerFunc
		report   HandlerFunc
	}

	pair struct {
		Method string
		Path   string
	}
)

const (
	skind kind = iota
	pkind
	akind
)

// NewRouter returns a new Router instance.
func NewRouter(s *Server) *Router {
	return &Router{
		tree: &node{
			methodHandler: new(methodHandler),
		},
		server: s,
	}
}

// Add registers a new route for method and path with matching handler.
func (r *Router) Add(method, path string, h HandlerFunc) {
	// Validate path
	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	pnames := []string{} // Param names
	ppath := path        // Pristine path

	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			j := i + 1

			r.insert(method, path[:i], nil, skind, "", nil)
			for ; i < l && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, l = j, len(path)

			if i == l {
				r.insert(method, path[:i], h, pkind, ppath, pnames)
			} else {
				r.insert(method, path[:i], nil, pkind, "", nil)
			}
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, skind, "", nil)
			pnames = append(pnames, "*")
			r.insert(method, path[:i+1], h, akind, ppath, pnames)
		}
	}

	r.insert(method, path, h, skind, ppath, pnames)
}

func (r *Router) insert(method, path string, h HandlerFunc, t kind, ppath string, pnames []string) {
	// Adjust max param
	l := len(pnames)
	if *r.server.maxParam < l {
		*r.server.maxParam = l
	}

	cn := r.tree // Current node as root
	if cn == nil {
		panic("engine: invalid method")
	}
	search := path

	for {
		sl := len(search)
		pl := len(cn.prefix)
		l := 0

		// LCP
		max := pl
		if sl < max {
			max = sl
		}
		for ; l < max && search[l] == cn.prefix[l]; l++ {
		}

		if l == 0 {
			// At root node
			cn.label = search[0]
			cn.prefix = search
			if h != nil {
				cn.kind = t
				cn.addHandler(method, h)
				cn.ppath = ppath
				cn.pnames = pnames
			}
		} else if l < pl {
			// Split node
			n := newNode(cn.kind, cn.prefix[l:], cn, cn.children, cn.methodHandler, cn.ppath, cn.pnames)

			// Update parent path for all children to new node
			for _, child := range cn.children {
				child.parent = n
			}

			// Reset parent node
			cn.kind = skind
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.children = nil
			cn.methodHandler = new(methodHandler)
			cn.ppath = ""
			cn.pnames = nil

			cn.addChild(n)

			if l == sl {
				// At parent node
				cn.kind = t
				cn.addHandler(method, h)
				cn.ppath = ppath
				cn.pnames = pnames
			} else {
				// Create child node
				n = newNode(t, search[l:], cn, nil, new(methodHandler), ppath, pnames)
				n.addHandler(method, h)
				cn.addChild(n)
			}
		} else if l < sl {
			search = search[l:]
			c := cn.findChildWithLabel(search[0])
			if c != nil {
				// Go deeper
				cn = c
				continue
			}
			// Create child node
			n := newNode(t, search, cn, nil, new(methodHandler), ppath, pnames)
			n.addHandler(method, h)
			cn.addChild(n)
		} else {
			// Node already exists
			if h != nil {
				cn.addHandler(method, h)
				cn.ppath = ppath
				if len(cn.pnames) == 0 { // Issue #729
					cn.pnames = pnames
				}
			}
		}
		return
	}
}

// Find lookup a handler registered for method and path. It also parses URL for path
// parameters and load them into context.
//
// For performance:
//
// - Get context from `Server#AcquireContext()`
// - Reset it `Context#Reset()`
// - Return it `Server#ReleaseContext()`.
func (r *Router) Find(method, path string, c Context) {
	ctx := c.(*context)
	ctx.path = path
	cn := r.tree // Current node as root

	var (
		search  = path
		child   *node         // Child node
		n       int           // Param counter
		nk      kind          // Next kind
		nn      *node         // Next node
		ns      string        // Next search
		pvalues = ctx.pvalues // Use the internal slice so the interface can keep the illusion of a dynamic slice
	)

	// Search order static > param > any
	for {
		if search == "" {
			break
		}

		pl := 0 // Prefix length
		l := 0  // LCP length

		if cn.label != ':' {
			sl := len(search)
			pl = len(cn.prefix)

			// LCP
			max := pl
			if sl < max {
				max = sl
			}
			for ; l < max && search[l] == cn.prefix[l]; l++ {
			}
		}

		if l == pl {
			// Continue search
			search = search[l:]
			// Finish routing if no remaining search and we are on an leaf node
			if search == "" && (nn == nil || cn.parent == nil || cn.ppath != "") {
				break
			}
		}

		// Attempt to go back up the tree on no matching prefix or no remaining search
		if l != pl || search == "" {
			// Handle special case of trailing slash route with existing any route (see #1526)
			if path[len(path)-1] == '/' && cn.findChildByKind(akind) != nil {
				goto Any
			}
			if nn == nil { // Issue #1348
				return // Not found
			}
			cn = nn
			search = ns
			if nk == pkind {
				goto Param
			} else if nk == akind {
				goto Any
			}
		}

		// Static node
		if child = cn.findChild(search[0], skind); child != nil {
			// Save next
			if cn.prefix[len(cn.prefix)-1] == '/' { // Issue #623
				nk = pkind
				nn = cn
				ns = search
			}
			cn = child
			continue
		}

	Param:
		// Param node
		if child = cn.findChildByKind(pkind); child != nil {
			// Issue #378
			if len(pvalues) == n {
				continue
			}

			// Save next
			if cn.prefix[len(cn.prefix)-1] == '/' { // Issue #623
				nk = akind
				nn = cn
				ns = search
			}

			cn = child
			i, l := 0, len(search)
			for ; i < l && search[i] != '/'; i++ {
			}
			pvalues[n] = search[:i]
			n++
			search = search[i:]
			continue
		}

	Any:
		// Any node
		if cn = cn.findChildByKind(akind); cn != nil {
			// If any node is found, use remaining path for pvalues
			pvalues[len(cn.pnames)-1] = search
			break
		}

		// No node found, continue at stored next node
		// or find nearest "any" route
		if nn != nil {
			// No next node to go down in routing (issue #954)
			// Find nearest "any" route going up the routing tree
			search = ns
			np := nn.parent
			// Consider param route one level up only
			if cn = nn.findChildByKind(pkind); cn != nil {
				pos := strings.IndexByte(ns, '/')
				if pos == -1 {
					// If no slash is remaining in search string set param value
					pvalues[len(cn.pnames)-1] = search
					break
				} else if pos > 0 {
					// Otherwise continue route processing with restored next node
					cn = nn
					nn = nil
					ns = ""
					goto Param
				}
			}
			// No param route found, try to resolve nearest any route
			for {
				np = nn.parent
				if cn = nn.findChildByKind(akind); cn != nil {
					break
				}
				if np == nil {
					break // no further parent nodes in tree, abort
				}
				var str strings.Builder
				str.WriteString(nn.prefix)
				str.WriteString(search)
				search = str.String()
				nn = np
			}
			if cn != nil { // use the found "any" route and update path
				pvalues[len(cn.pnames)-1] = search
				break
			}
		}
		return // Not found

	}

	ctx.handler = cn.findHandler(method)
	ctx.path = cn.ppath
	ctx.pnames = cn.pnames

	// NOTE: Slow zone...
	if ctx.handler == nil {
		ctx.handler = cn.checkMethodNotAllowed()

		// Dig further for any, might have an empty value for *, e.g.
		if cn = cn.findChildByKind(akind); cn == nil {
			return
		}
		if h := cn.findHandler(method); h != nil {
			ctx.handler = h
		} else {
			ctx.handler = cn.checkMethodNotAllowed()
		}
		ctx.path = cn.ppath
		ctx.pnames = cn.pnames
		pvalues[len(cn.pnames)-1] = ""
	}

	return
}

type sNode struct {
	childs children
	curr   int
}

func (n *sNode) current() *node {
	if n.curr < len(n.childs) {
		return n.childs[n.curr]
	}
	return nil
}

// GetList 获取所有路由
func (r *Router) GetList() []*pair {

	var list []*pair

	stack := []*sNode{
		{children{r.tree}, 0},
	}

	for {
		l := len(stack)
		if l <= 0 {
			break
		}
		sn := stack[l-1] // 取栈最后一个node
		csn := sn.current()

		if sn.childs == nil {
			stack = stack[0 : l-1]
		} else {
			if csn.children != nil {
				childs := csn.children
				stack = append(stack, &sNode{
					childs,
					0,
				})
			}

			sn.curr++
			if sn.curr >= len(sn.childs) {
				// last
				sn.childs = nil
			}

			// 获取路由
			if pr, ok := r.fetchPath(csn); ok {
				list = append(list, pr...)
			}
			//
		}
	}

	return list
}

// fetchPath 获取有效node的path
func (r *Router) fetchPath(csn *node) (list []*pair, ok bool) {
	ok = false
	if csn.kind != 0 {
		list = nil
		return
	}
	if csn.methodHandler == nil {
		list = nil
		return
	}
	t := reflect.TypeOf(*csn.methodHandler)
	v := reflect.ValueOf(*csn.methodHandler)
	for k := 0; k < t.NumField(); k++ {
		if v.Field(k).IsNil() {
			continue
		}
		method := strings.ToUpper(t.Field(k).Name)
		list = append(list, &pair{method, csn.ppath})
		ok = true
	}
	return
}
