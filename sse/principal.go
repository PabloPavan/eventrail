package sse

import "net/http"

type Principal struct {
	UserID  int64
	ScopeID int64
}

type PrincipalResolver interface {
	Resolve(r *http.Request) (*Principal, error)
}
