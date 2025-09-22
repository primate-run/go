package route

import (
	"encoding/json"
	"fmt"
	"github.com/primate-run/go/core"
	"sync"
	"syscall/js"
)

type Request = core.Request
type Handler = func(Request) any

var (
	mu          sync.Mutex
	reg         = map[string]Handler{}
	initialized bool
)

func register(verb string, h Handler) Handler {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := reg[verb]; exists {
		panic(fmt.Sprintf("route: duplicate handler for %s", verb))
	}
	reg[verb] = h
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

func makeURL(request js.Value) core.URL {
	url := request.Get("url")
	searchParams := make(map[string]any)
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
		Path:    decodeMap(request.Get("path").String()),
		Query:   decodeMap(request.Get("query").String()),
		Headers: decodeMap(request.Get("headers").String()),
		Cookies: decodeMap(request.Get("cookies").String()),
	}
}

func decodeMap(s string) map[string]any {
	m := map[string]any{}
	if s != "" {
		_ = json.Unmarshal([]byte(s), &m)
	}
	return m
}

func Commit() {
	mu.Lock()
	defer mu.Unlock()

	if initialized {
		return
	}
	initialized = true

	js.Global().Set("__primate_handle", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) < 2 {
			return "null"
		}
		verb := args[0].String()
		reqObj := args[1]

		mu.Lock()
		h, ok := reg[verb]
		mu.Unlock()

		if !ok {
			return `{"error":"no handler for ` + verb + `"}`
		}
		goReq := makeRequest(reqObj)
		resp := h(goReq)

		if fn, ok := resp.(js.Func); ok {
			return fn
		}
		b, _ := json.Marshal(resp)
		return string(b)
	}))

	verbs := make([]string, 0, len(reg))
	for k := range reg {
		verbs = append(verbs, k)
	}
	arr := js.Global().Get("Array").New(len(verbs))
	for i, v := range verbs {
		arr.SetIndex(i, v)
	}
	js.Global().Set("__primate_verbs", arr)
}
