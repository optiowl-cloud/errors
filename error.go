package errors

import (
	"fmt"

	"log/slog"

	pkgerrors "github.com/pkg/errors"
)

type goError struct {
	pkgError   error
	attributes []ErrorAttribute
	stack      []ErrorAttribute
	class      string
}

type stackTracer interface {
	StackTrace() []uintptr
}

type errStackTracer interface {
	StackTrace() pkgerrors.StackTrace
}

type unwrapper interface {
	Unwrap() error
}

type slogAttrs interface {
	SlogAttrs() []slog.Attr
}

type causer interface {
	Cause() error
}

var _ stackTracer = &goError{}
var _ unwrapper = &goError{}
var _ error = &goError{}
var _ slogAttrs = &goError{}
var _ causer = &goError{}
var _ fmt.Formatter = &goError{}
