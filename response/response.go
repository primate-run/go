package response

import (
	"encoding/json"
	"syscall/js"

	"github.com/primate-run/go/types"
)

type Dict = types.Dict

func tryMap(array []Dict, position uint8, fallback Dict) Dict {
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

func serialize(data Dict) string {
	if data == nil {
		return ""
	}
	serialized, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(serialized)
}

func View(component string, props Dict, options ...Dict) any {
	var serde_props = serialize(props)
	var serde_options = serialize(tryMap(options, 0, Dict{}))

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return Dict{
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
		return Dict{
			"handler":  "redirect",
			"location": location,
			"status":   status,
		}
	})
}

func Error(options ...Dict) any {
	var serde_options = serialize(tryMap(options, 0, Dict{}))

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return Dict{
			"handler": "error",
			"options": serde_options,
		}
	})
}
