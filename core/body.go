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
	KindForm
	KindMultipart
	KindBlob
)

func parseKind(s string) Kind {
	switch s {
	case "text/plain":
		return KindText
	case "application/json":
		return KindJSON
	case "application/x-www-form-urlencoded":
		return KindForm
	case "multipart/form-data":
		return KindMultipart
	case "application/octet-stream":
		return KindBlob
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

	onceForm sync.Once
	formRaw  []byte
	formErr  error

	onceMultipart sync.Once
	multipartRaw  []byte
	multipartErr  error

	onceBlob sync.Once
	blobData []byte
	blobType string
	blobErr  error
}

func NewBodyFromJS(v js.Value, contentType string) *Body {
	return &Body{
		jsObj: v,
		kind:  parseKind(contentType),
	}
}

func (body *Body) Kind() Kind { return body.kind }

func (body *Body) Text() (string, error) {
	if body.kind == KindNone {
		return "", errors.New("no content-type declared; use route.With{ContentType: route.Text}")
	}
	if body.kind != KindText {
		return "", errors.New("expected text body; declare route.With{ContentType: route.Text}")
	}
	body.onceText.Do(func() {
		body.text = body.jsObj.Call("textSync").String()
	})
	return body.text, body.textErr
}

func (body *Body) JSON() (Dict, error) {
	if body.kind == KindNone {
		return nil, errors.New("no content-type declared; use route.With{ContentType: route.JSON}")
	}
	if body.kind != KindJSON {
		return nil, errors.New("expected json body; declare route.With{ContentType: route.JSON}")
	}
	body.onceJSON.Do(func() {
		s := body.jsObj.Call("jsonSync").String()
		body.jsonRaw = []byte(s)
	})
	if body.jsonErr != nil {
		return nil, body.jsonErr
	}

	var data Dict
	dec := json.NewDecoder(bytesReader(body.jsonRaw))
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (body *Body) Form() (Dict, error) {
	if body.kind == KindNone {
		return nil, errors.New("no content-type declared; use route.With{ContentType: route.Form}")
	}
	if body.kind != KindForm {
		return nil, errors.New("expected form body; declare route.With{ContentType: route.Form}")
	}
	body.onceForm.Do(func() {
		s := body.jsObj.Call("formSync").String()
		body.formRaw = []byte(s)
	})
	if body.formErr != nil {
		return nil, body.formErr
	}

	var data Dict
	if err := json.Unmarshal(body.formRaw, &data); err != nil {
		return nil, err
	}

	return data, nil
}

type UploadFile struct {
	Field string
	Name  string
	Type  string
	Size  int64
	Bytes []byte
}

func (body *Body) Multipart() (Dict, []UploadFile, error) {
	if body.kind == KindNone {
		return nil, nil, errors.New("no content-type declared; use route.With{ContentType: route.Multipart}")
	}
	if body.kind != KindMultipart {
		return nil, nil, errors.New("expected multipart body; declare route.With{ContentType: route.Multipart}")
	}
	body.onceMultipart.Do(func() {
		s := body.jsObj.Call("formSync").String()
		body.multipartRaw = []byte(s)
	})
	if body.multipartErr != nil {
		return nil, nil, body.multipartErr
	}

	var form Dict
	if err := json.Unmarshal(body.multipartRaw, &form); err != nil {
		return nil, nil, err
	}

	arr := body.jsObj.Call("filesSync")
	var files []UploadFile
	if !arr.IsUndefined() && !arr.IsNull() {
		n := arr.Length()
		files = make([]UploadFile, 0, n)
		for i := range n {
			it := arr.Index(i)
			field := it.Get("field").String()
			name := it.Get("name").String()
			typ := it.Get("type").String()
			size := int64(it.Get("size").Int())
			u8 := it.Get("bytes")
			buf := make([]byte, u8.Get("length").Int())
			js.CopyBytesToGo(buf, u8)
			files = append(files, UploadFile{
				Field: field, Name: name, Type: typ, Size: size, Bytes: buf,
			})
		}
	}

	return form, files, nil
}

type Blob struct {
	Data []byte
	Type string
}

func (body *Body) Blob() (Blob, error) {
	if body.kind == KindNone {
		return Blob{}, errors.New("no content-type declared; use route.With{ContentType: route.Blob}")
	}
	if body.kind != KindBlob {
		return Blob{}, errors.New("expected blob body; declare route.With{ContentType: route.Blob}")
	}
	body.onceBlob.Do(func() {
		u8 := body.jsObj.Call("blobSync")
		n := u8.Get("length").Int()
		buf := make([]byte, n)
		js.CopyBytesToGo(buf, u8)
		body.blobData = buf
		body.blobType = body.jsObj.Call("blobTypeSync").String()
	})
	return Blob{Data: body.blobData, Type: body.blobType}, body.blobErr
}

type byteReader []byte

func (r byteReader) Read(p []byte) (int, error) {
	n := copy(p, r)
	if n < len(r) {
		return n, nil
	}
	return n, io.EOF
}

func bytesReader(b []byte) io.Reader { return byteReader(b) }
