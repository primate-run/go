package session

import (
	"encoding/json"
	"syscall/js"

	"github.com/primate-run/go/core"
)

type SessionData = core.Dict

type SessionType struct {
	Id      string
	Exists  bool
	Create  func(SessionData)
	Get     func() SessionData
	Try     func() SessionData
	Set     func(SessionData)
	Destroy func()
}

func Session() SessionType {
	var session = js.Global().Get("PRMT_SESSION")

	return SessionType{
		Id:     session.Get("id").String(),
		Exists: session.Get("exists").Bool(),

		Create: func(data SessionData) {
			serialized, _ := json.Marshal(data)
			session.Get("create").Invoke(string(serialized))
		},

		Get: func() SessionData {
			data := make(core.Dict)
			raw := session.Get("get").Invoke().String()
			_ = json.Unmarshal([]byte(raw), &data)
			return data
		},

		Try: func() SessionData {
			data := make(core.Dict)
			raw := session.Get("try").Invoke().String()
			_ = json.Unmarshal([]byte(raw), &data)
			return data
		},

		Set: func(data SessionData) {
			serialized, _ := json.Marshal(data)
			session.Get("set").Invoke(string(serialized))
		},

		Destroy: func() {
			session.Get("destroy").Invoke()
		},
	}
}

func Id() string {
	return Session().Id
}

func Exists() bool {
	return Session().Exists
}

func Create(data SessionData) {
	Session().Create(data)
}

func Get() SessionData {
	return Session().Get()
}

func Try() SessionData {
	return Session().Try()
}

func Set(data SessionData) {
	Session().Set(data)
}

func Destroy() {
	Session().Destroy()
}
