package route

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"syscall/js"

	"github.com/primate-run/go/core"
	"github.com/primate-run/go/types"
)

type Request = core.Request
type Handler = func(Request) any
type Dict = types.Dict

var (
	mu      sync.Mutex
	scopes  = map[string]map[string]Handler{}
	pending = []struct {
		verb    string
		handler Handler
	}{}
)

func register(verb string, h Handler) Handler {
	mu.Lock()
	defer mu.Unlock()

	pending = append(pending, struct {
		verb    string
		handler Handler
	}{verb, h})

	return h
}

func Post(h Handler) Handler    { return register("POST", h) }
func Get(h Handler) Handler     { return register("GET", h) }
func Put(h Handler) Handler     { return register("PUT", h) }
func Patch(h Handler) Handler   { return register("PATCH", h) }
func Delete(h Handler) Handler  { return register("DELETE", h) }
func Head(h Handler) Handler    { return register("HEAD", h) }
func Connect(h Handler) Handler { return register("CONNECT", h) }
func Options(h Handler) Handler { return register("OPTIONS", h) }
func Trace(h Handler) Handler   { return register("TRACE", h) }

func Registry(id string) []string {
	mu.Lock()
	defer mu.Unlock()

	registry := scopes[id]
	if registry == nil {
		return []string{}
	}

	verbs := make([]string, 0, len(registry))
	for verb := range registry {
		verbs = append(verbs, verb)
	}
	return verbs
}

func makeURL(request js.Value) core.URL {
	url := request.Get("url")
	searchParams := make(Dict)
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

func makeRequest(request js.Value) core.Request {
	return core.Request{
		Url:     makeURL(request),
		Body:    core.NewBodyFromJS(request.Get("body")),
		Path:    makeRequestBag(request.Get("path").String(), "path"),
		Query:   makeRequestBag(request.Get("query").String(), "query"),
		Headers: makeRequestBag(request.Get("headers").String(), "headers"),
		Cookies: makeRequestBag(request.Get("cookies").String(), "cookies"),
	}
}

func makeRequestBag(jsonStr, name string) *core.RequestBag {
	data := make(Dict)
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
	h, ok := registry[verb]
	mu.Unlock()

	if !ok {
		return `{"error":"no handler for ` + verb + ` in scope ` + scope_id + `"}`
	}

	response := h(makeRequest(request))

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
		scopes[scope_id] = make(map[string]Handler)
	}

	for _, p := range pending {
		if _, exists := scopes[scope_id][p.verb]; exists {
			panic(fmt.Sprintf("route: duplicate handler for %s in scope %s", p.verb, scope_id))
		}
		scopes[scope_id][p.verb] = p.handler
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
		verbs := Registry(scope_id)
		arr := js.Global().Get("Array").New(len(verbs))
		for i, v := range verbs {
			arr.SetIndex(i, v)
		}
		return arr
	}))

	ready_callback := "__primate_go_ready_" + safe_scope_id
	if cb := js.Global().Get(ready_callback); !cb.IsUndefined() {
		cb.Invoke()
	}
}
