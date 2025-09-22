# primate-run/go
Primate Go integration.

## Packages
- `github.com/primate-run/go/core`
  Shared types (`Request`, `Body`, `Kind`) + JS/WASM bridge methods.

- `github.com/primate-run/go/route`
  Route verbs: `route.Post`, `route.Get`

## Usage
```go
package main

import "github.com/primate-run/go/route"

var _ = route.Post(func(request route.Request) any {
  s, _ := request.Body.Text()
  return s
})
```
