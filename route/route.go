package route

import (
	"github.com/primate-run/go/core"
	"github.com/primate-run/go/route/internal/registry"
)

type Request = core.Request
type Handler = func(Request) any

func Post(h Handler) Handler    { return registry.Register("POST", h) }
func Get(h Handler) Handler     { return registry.Register("GET", h) }
func Put(h Handler) Handler     { return registry.Register("PUT", h) }
func Patch(h Handler) Handler   { return registry.Register("PATCH", h) }
func Delete(h Handler) Handler  { return registry.Register("DELETE", h) }
func Head(h Handler) Handler    { return registry.Register("HEAD", h) }
func Connect(h Handler) Handler { return registry.Register("CONNECT", h) }
func Options(h Handler) Handler { return registry.Register("OPTIONS", h) }
func Trace(h Handler) Handler   { return registry.Register("TRACE", h) }
