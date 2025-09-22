package core

import (
	"encoding/json"
	"errors"
	"io"
	"sync"
	"syscall/js"
)

type Kind int

const (
	KindNone Kind = iota
	KindText
	KindJSON
	KindFields
	KindBin
)

func parseKind(s string) Kind {
	switch s {
	case "text":
		return KindText
	case "json":
		return KindJSON
	case "fields":
		return KindFields
	case "bin":
		return KindBin
	default:
		return KindNone
	}
}

type Body struct {
	jsObj js.Value
	kind  Kind

	onceText sync.Once
	text     string
	textErr  error

	onceJSON sync.Once
	jsonRaw  []byte
	jsonErr  error

	onceFields sync.Once
	fieldsRaw  []byte
	fieldsErr  error

	onceBin sync.Once
	bin     []byte
	binType string
	binErr  error
}

func NewBodyFromJS(v js.Value) *Body {
	return &Body{
		jsObj: v,
		kind:  parseKind(v.Get("type").String()),
	}
}

func (body *Body) Kind() Kind { return body.kind }

// Text -> string (from textSync)
func (body *Body) Text() (string, error) {
	if body.kind != KindText {
		return "", errors.New("expected text body")
	}
	body.onceText.Do(func() {
		body.text = body.jsObj.Call("textSync").String()
	})
	return body.text, body.textErr
}

// JSON -> unmarshal into dst (from jsonSync)
func (body *Body) JSON(dst any) error {
	if body.kind != KindJSON {
		return errors.New("expected json body")
	}
	body.onceJSON.Do(func() {
		s := body.jsObj.Call("jsonSync").String()
		body.jsonRaw = []byte(s)
	})
	if body.jsonErr != nil {
		return body.jsonErr
	}
	dec := json.NewDecoder(bytesReader(body.jsonRaw))
	return dec.Decode(dst)
}

// Fields -> unmarshal fields JSON into dst (from fieldsSync)
func (body *Body) Fields(dst any) error {
	if body.kind != KindFields {
		return errors.New("expected fields body")
	}
	body.onceFields.Do(func() {
		s := body.jsObj.Call("fieldsSync").String()
		body.fieldsRaw = []byte(s)
	})
	if body.fieldsErr != nil {
		return body.fieldsErr
	}
	return json.Unmarshal(body.fieldsRaw, dst)
}

// UploadFile describes a file from multipart fields (bytes come from filesSync)
type UploadFile struct {
	Field string
	Name  string
	Type  string
	Size  int64
	Bytes []byte
}

// Files -> read filesSync() array [{field,name,type,size,bytes:Uint8Array}]
func (body *Body) Files() ([]UploadFile, error) {
	if body.kind != KindFields {
		return nil, errors.New("expected fields body")
	}
	arr := body.jsObj.Call("filesSync")
	if arr.IsUndefined() || arr.IsNull() {
		return nil, nil
	}
	n := arr.Length()
	out := make([]UploadFile, 0, n)
	for i := 0; i < n; i++ {
		it := arr.Index(i)
		field := it.Get("field").String()
		name := it.Get("name").String()
		typ := it.Get("type").String()
		size := int64(it.Get("size").Int())
		u8 := it.Get("bytes")
		buf := make([]byte, u8.Get("length").Int())
		js.CopyBytesToGo(buf, u8)
		out = append(out, UploadFile{
			Field: field, Name: name, Type: typ, Size: size, Bytes: buf,
		})
	}
	return out, nil
}

// Binary -> data + mime (from binarySync/binaryTypeSync)
func (body *Body) Binary() ([]byte, string, error) {
	if body.kind != KindBin {
		return nil, "", errors.New("expected binary body")
	}
	body.onceBin.Do(func() {
		u8 := body.jsObj.Call("binarySync")
		n := u8.Get("length").Int()
		buf := make([]byte, n)
		js.CopyBytesToGo(buf, u8)
		body.bin = buf
		body.binType = body.jsObj.Call("binaryTypeSync").String()
	})
	return body.bin, body.binType, body.binErr
}

// tiny reader to avoid bringing in bytes pkg here
type byteReader []byte

func (r byteReader) Read(p []byte) (int, error) {
	n := copy(p, r)
	if n < len(r) {
		return n, nil
	}
	return n, io.EOF
}
func bytesReader(b []byte) io.Reader { return byteReader(b) }
