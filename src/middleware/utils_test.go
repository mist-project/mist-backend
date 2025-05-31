package middleware_test

import "context"

type DummyRequest struct{}

var MockHandler = func(ctx context.Context, req interface{}) (interface{}, error) { return req, nil }
