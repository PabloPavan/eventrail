package sse

import "net/http"

type Principal struct {
	UserID    string
	UserAgent string
}

type PrincipalResolver interface {
	Resolve(r *http.Request) (*Principal, error)
}
