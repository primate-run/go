package postlude

import (
	"encoding/json"
	"syscall/js"

	"github.com/primate-run/go/core"
	"github.com/primate-run/go/route/internal/registry"
)

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
	// build from your facade (same as you do today)
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

// Install sets global JS entry points for the host.
func Install() {
	js.Global().Set("__primate_handle", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) < 2 {
			return "null"
		}
		verb := args[0].String()
		reqObj := args[1]

		h, ok := registry.Lookup(verb)
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
	verbs := registry.Verbs()
	arr := js.Global().Get("Array").New(len(verbs))
	for i, v := range verbs {
		arr.SetIndex(i, v)
	}
	js.Global().Set("__primate_verbs", arr)
}
