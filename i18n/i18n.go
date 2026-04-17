//go:build js && wasm

package i18n

import (
	"encoding/json"
	"syscall/js"

	"github.com/primate-run/go/core"
)

type Vars = core.Dict
type LocaleAccessor struct{}

func get() js.Value {
	return js.Global().Get("PRMT_I18N")
}

func T(key string, params ...Vars) string {
	i18n := get()
	if len(params) == 0 || params[0] == nil {
		return i18n.Get("t").Invoke(key).String()
	}
	serialized, _ := json.Marshal(params[0])
	return i18n.Get("t").Invoke(key, string(serialized)).String()
}

func (LocaleAccessor) Get() string {
	return get().Get("locale").String()
}

func (LocaleAccessor) Set(locale string) {
	get().Get("set").Invoke(locale)
}

var Locale LocaleAccessor
