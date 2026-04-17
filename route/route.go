//go:build js && wasm

package route

import (
	"encoding/json"
	"strings"
	"sync"
	"syscall/js"

	"github.com/primate-run/go/core"
)

type Request = core.Request
type Props = map[string]any
type Handler = func(Request) any

type ContentType string

const (
	JSON      ContentType = "application/json"
	Text      ContentType = "text/plain"
	Form      ContentType = "application/x-www-form-urlencoded"
	Multipart ContentType = "multipart/form-data"
	Blob      ContentType = "application/octet-stream"
)

type With struct {
	ContentType ContentType
}

type entry struct {
	handler     Handler
	contentType ContentType
}

var (
	mu      sync.Mutex
	scopes  = map[string]map[string]entry{}
	pending = []struct {
		verb        string
		handler     Handler
		contentType ContentType
	}{}
)

func register(verb string, h Handler, w With) Handler {
	mu.Lock()
	defer mu.Unlock()

	pending = append(pending, struct {
		verb        string
		handler     Handler
		contentType ContentType
	}{verb, h, w.ContentType})

	return h
}

// top-level functions for no-options case
func Get(h Handler) Handler     { return register("GET", h, With{}) }
func Post(h Handler) Handler    { return register("POST", h, With{}) }
func Put(h Handler) Handler     { return register("PUT", h, With{}) }
func Patch(h Handler) Handler   { return register("PATCH", h, With{}) }
func Delete(h Handler) Handler  { return register("DELETE", h, With{}) }
func Head(h Handler) Handler    { return register("HEAD", h, With{}) }
func Connect(h Handler) Handler { return register("CONNECT", h, With{}) }
func Options(h Handler) Handler { return register("OPTIONS", h, With{}) }
func Trace(h Handler) Handler   { return register("TRACE", h, With{}) }

// methods on With for the with-options case
func (w With) Get(h Handler) Handler     { return register("GET", h, w) }
func (w With) Post(h Handler) Handler    { return register("POST", h, w) }
func (w With) Put(h Handler) Handler     { return register("PUT", h, w) }
func (w With) Patch(h Handler) Handler   { return register("PATCH", h, w) }
func (w With) Delete(h Handler) Handler  { return register("DELETE", h, w) }
func (w With) Head(h Handler) Handler    { return register("HEAD", h, w) }
func (w With) Connect(h Handler) Handler { return register("CONNECT", h, w) }
func (w With) Options(h Handler) Handler { return register("OPTIONS", h, w) }
func (w With) Trace(h Handler) Handler   { return register("TRACE", h, w) }

func makeURL(request js.Value) core.URL {
	url := request.Get("url")
	searchParams := make(core.Dict)
	json.Unmarshal([]byte(request.Get("searchParams").String()), &searchParams)

	return core.URL{
		Href:         url.Get("href").String(),
		Origin:       url.Get("origin").String(),
		Protocol:     url.Get("protocol").String(),
		Username:     url.Get("username").String(),
		Password:     url.Get("password").String(),
		Host:         url.Get("host").String(),
		Hostname:     url.Get("hostname").String(),
		Port:         url.Get("port").String(),
		Pathname:     url.Get("pathname").String(),
		Search:       url.Get("search").String(),
		SearchParams: searchParams,
		Hash:         url.Get("hash").String(),
	}
}

func makeRequest(request js.Value, contentType string) core.Request {
	return core.Request{
		Url:     makeURL(request),
		Body:    core.NewBodyFromJS(request.Get("body"), contentType),
		Path:    makeRequestBag(request.Get("path").String(), "path"),
		Query:   makeRequestBag(request.Get("query").String(), "query"),
		Headers: makeRequestBag(request.Get("headers").String(), "headers"),
		Cookies: makeRequestBag(request.Get("cookies").String(), "cookies"),
	}
}

func makeRequestBag(jsonStr, name string) *core.RequestBag {
	data := make(core.Dict)
	if jsonStr != "" {
		json.Unmarshal([]byte(jsonStr), &data)
	}
	return core.NewRequestBag(data, name)
}

func CallJS(scope_id, verb string, request js.Value) any {
	mu.Lock()
	registry := scopes[scope_id]
	if registry == nil {
		mu.Unlock()
		return `{"error":"no scope ` + scope_id + `"}`
	}
	e, ok := registry[verb]
	mu.Unlock()

	if !ok {
		return `{"error":"no handler for ` + verb + ` in scope ` + scope_id + `"}`
	}

	response := e.handler(makeRequest(request, string(e.contentType)))

	if fn, ok := response.(js.Func); ok {
		return fn.Invoke()
	}

	b, _ := json.Marshal(response)
	return string(b)
}

func Commit(scope_id string) {
	mu.Lock()
	defer mu.Unlock()

	if scopes[scope_id] == nil {
		scopes[scope_id] = make(map[string]entry)
	}

	for _, p := range pending {
		scopes[scope_id][p.verb] = entry{
			handler:     p.handler,
			contentType: p.contentType,
		}
	}
	pending = nil

	safe_scope_id := strings.ReplaceAll(scope_id, "/", "_")
	call_go := "__primate_call_go_" + safe_scope_id
	registry_name := "__primate_go_registry_" + safe_scope_id

	js.Global().Set(call_go, js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) < 2 {
			return `{"error":"insufficient arguments"}`
		}
		verb := args[0].String()
		request := args[1]
		return CallJS(scope_id, verb, request)
	}))

	js.Global().Set(registry_name, js.FuncOf(func(_ js.Value, args []js.Value) any {
		registry := scopes[scope_id]
		arr := js.Global().Get("Array").New(len(registry))
		i := 0
		for verb, e := range registry {
			obj := js.Global().Get("Object").New()
			obj.Set("verb", verb)
			obj.Set("contentType", string(e.contentType))
			arr.SetIndex(i, obj)
			i++
		}
		return arr
	}))

	ready_callback := "__primate_go_ready_" + safe_scope_id
	if cb := js.Global().Get(ready_callback); !cb.IsUndefined() {
		cb.Invoke()
	}
}
