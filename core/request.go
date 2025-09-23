package core

type URL struct {
	Href         string
	Origin       string
	Protocol     string
	Username     string
	Password     string
	Host         string
	Hostname     string
	Port         string
	Pathname     string
	Search       string
	SearchParams Dict
	Hash         string
}

type Request struct {
	Url     URL
	Body    *Body
	Path    *RequestBag
	Query   *RequestBag
	Headers *RequestBag
	Cookies *RequestBag
}
