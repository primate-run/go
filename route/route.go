package route

import (
	"github.com/primate-run/go/core"
)

type Request = core.Request
type Handler = func(Request) any

var (
	Register func(verb string, h Handler) Handler
	Commit   func()
)

// public API
func Post(h Handler) Handler    { return Register("POST", h) }
func Get(h Handler) Handler     { return Register("GET", h) }
func Put(h Handler) Handler     { return Register("PUT", h) }
func Patch(h Handler) Handler   { return Register("PATCH", h) }
func Delete(h Handler) Handler  { return Register("DELETE", h) }
func Head(h Handler) Handler    { return Register("HEAD", h) }
func Connect(h Handler) Handler { return Register("CONNECT", h) }
func Options(h Handler) Handler { return Register("OPTIONS", h) }
func Trace(h Handler) Handler   { return Register("TRACE", h) }
