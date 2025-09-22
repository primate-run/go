package postlude

import (
	"encoding/json"
	"fmt"
	"sync"
	"syscall/js"

	"github.com/primate-run/go/core"
	"github.com/primate-run/go/route"
)

var (
	mu  sync.Mutex
	reg = map[string]route.Handler{}
)

func init() {
	// wire up
	route.Register = register
	route.Commit = Install
}

func register(verb string, h route.Handler) route.Handler {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := reg[verb]; exists {
		panic(fmt.Sprintf("route: duplicate handler for %s", verb))
	}
	reg[verb] = h
	return h
}

func lookup(verb string) (route.Handler, bool) {
	mu.Lock()
	defer mu.Unlock()
	h, ok := reg[verb]
	return h, ok
}

func verbs() []string {
	mu.Lock()
	defer mu.Unlock()
	out := make([]string, 0, len(reg))
	for k := range reg {
		out = append(out, k)
	}
	return out
}

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

// install sets global JS entry points for the host
func Install() {
	js.Global().Set("__primate_handle", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) < 2 {
			return "null"
		}
		verb := args[0].String()
		reqObj := args[1]

		h, ok := lookup(verb)
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

	// publish registered verbs for debugging
	verbList := verbs()
	arr := js.Global().Get("Array").New(len(verbList))
	for i, v := range verbList {
		arr.SetIndex(i, v)
	}
	js.Global().Set("__primate_verbs", arr)
}
