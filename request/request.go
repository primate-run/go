package request

import (
	"encoding/json"
	"syscall/js"

	"github.com/primate-run/go/core"
)

func tryMap(array []core.Dict, position uint8, fallback core.Dict) core.Dict {
	if len(array) <= int(position) {
		return fallback
	}
	return array[position]
}

func tryInt(array []int, position uint8, fallback int) int {
	if len(array) <= int(position) {
		return fallback
	}
	return array[position]
}

func serialize(data map[string]any) string {
	if data == nil {
		return ""
	}

	serialized, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(serialized)
}

func View(component string, props core.Dict, options ...core.Dict) any {
	var serde_props = serialize(props)

	var serde_options = serialize(tryMap(options, 0, core.Dict{}))

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return map[string]any{
			"handler":   "view",
			"component": component,
			"props":     serde_props,
			"options":   serde_options,
		}
	})
}

func Redirect(location string, ints ...int) any {
	var status = tryInt(ints, 0, 302)

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return map[string]any{
			"handler":  "redirect",
			"location": location,
			"status":   status,
		}
	})
}

func Error(options ...core.Dict) any {
	var serde_options = serialize(tryMap(options, 0, core.Dict{}))

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return map[string]any{
			"handler": "error",
			"options": serde_options,
		}
	})
}
