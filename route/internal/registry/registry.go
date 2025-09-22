package registry

import (
  "fmt"
  "sync"

  "github.com/primate-run/go/core"
)

type Handler = func(core.Request) any

var (
  mu  sync.Mutex
  reg = map[string]Handler{}
)

func Register(verb string, h Handler) Handler {
  mu.Lock()
  defer mu.Unlock()
  if _, exists := reg[verb]; exists {
    panic(fmt.Sprintf("route: duplicate handler for %s", verb))
  }
  reg[verb] = h
  return h
}

func Lookup(verb string) (Handler, bool) {
  mu.Lock(); defer mu.Unlock()
  h, ok := reg[verb]
  return h, ok
}

func Verbs() []string {
  mu.Lock(); defer mu.Unlock()
  out := make([]string, 0, len(reg))
  for k := range reg { out = append(out, k) }
  return out
}
