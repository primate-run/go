package core

import (
	"fmt"
	"maps"

	"github.com/primate-run/go/pema"
)

type RequestBag struct {
	contents map[string]string
	name     string
}

func NewRequestBag(data Dict, name string) *RequestBag {
	contents := make(map[string]string)
	for k, v := range data {
		if v != nil {
			contents[k] = fmt.Sprintf("%v", v)
		}
	}

	return &RequestBag{
		contents: contents,
		name:     name,
	}
}

func (rb *RequestBag) Size() int {
	return len(rb.contents)
}

func (rb *RequestBag) Get(key string) (string, error) {
	if value, exists := rb.contents[key]; exists {
		return value, nil
	}
	return "", fmt.Errorf("%s has no key %s", rb.name, key)
}

func (rb *RequestBag) Try(key string) string {
	if value, exists := rb.contents[key]; exists {
		return value
	}
	return ""
}

func (rb *RequestBag) Has(key string) bool {
	_, exists := rb.contents[key]
	return exists
}

func (rb *RequestBag) Parse(schema *pema.SchemaBuilder, coerce ...bool) (Dict, error) {
	data := make(Dict)
	for k, v := range rb.contents {
		data[k] = v
	}
	return schema.Parse(data, coerce...)
}

func (rb *RequestBag) ToJSON() map[string]string {
	return maps.Clone(rb.contents)
}
