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
	SearchParams map[string]any
	Hash         string
}

type Request struct {
	Url     URL
	Body    *Body
	Path    map[string]any
	Query   map[string]any
	Headers map[string]any
	Cookies map[string]any
}
