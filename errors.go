package errors

import (
	"errors"
	"fmt"
	"io"
	"runtime"

	"log/slog"

	"github.com/pborman/indent"
	pkgerrors "github.com/pkg/errors"
)

func New(message string, attrs ...ErrorAttribute) error {
	return &goError{
		pkgError:   pkgerrors.New(message),
		attributes: attrs,
	}
}

var noOpStack = ErrorAttribute{
	attrType: NoOpStackAttributeType,
}

func Join(errs ...error) error {
	return Wrap(errors.Join(errs...))
}

func Wrap(err error, attrs ...ErrorAttribute) error {
	if err == nil {
		return nil
	}
	return wrapInternal(err, attrs)
}

func WrapStackf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return wrapInternal(err, []ErrorAttribute{
		Stackf(format, args...),
	})
}

func WrapWithStack(err error, stackattr ErrorAttribute, attrs ...ErrorAttribute) error {
	temp := make([]ErrorAttribute, 0, len(attrs)+1)
	temp = append(temp, stackattr)
	temp = append(temp, attrs...)
	return Wrap(
		err,
		temp...,
	)
}

type WrapConfig struct {
	Error      error
	Stack      ErrorAttribute
	Attributes []ErrorAttribute
}

func wrapInternal(
	err error,
	attrs []ErrorAttribute,
) *goError {
	var class string = ""
	stackattrs := make([]ErrorAttribute, 0, 1)
	for _, attr := range attrs {
		if attr.attrType == StackAttributeType && attr != noOpStack {
			stackattrs = append(stackattrs, attr)
		}
	}
	if myErr, ok := err.(*goError); ok {
		myErr = myErr.addAttrs(attrs...)
		myErr.stack = append(myErr.stack, stackattrs...)
		return myErr
	}
	return &goError{
		pkgError:   pkgerrors.WithStack(err),
		attributes: attrs,
		class:      class,
		stack:      stackattrs,
	}
}

func Errorf(format string, args ...interface{}) error {
	return &goError{
		pkgError: pkgerrors.WithStack(fmt.Errorf(format, args...)),
	}
}

func (e *goError) Error() string {
	return e.pkgError.Error()
}

func (e *goError) Unwrap() error {
	return e.pkgError
}

func (e *goError) addAttrs(attrs ...ErrorAttribute) *goError {
	e.attributes = append(e.attributes, attrs...)
	return e
}

func (e *goError) ErrorClass() string {
	if e.class != "" {
		return e.class
	}
	type errorClasser interface {
		ErrorClass() string
	}
	cause := pkgerrors.Cause(e.pkgError)
	if ec, ok := cause.(errorClasser); ok {
		return ec.ErrorClass()
	}
	return fmt.Sprintf("%T", cause)
}

func (e *goError) ErrorAttributes() map[string]interface{} {
	attrs := ErrorAttributes(e.attributes).ErrorAttributes()
	return attrs
}

func (e *goError) SlogAttrs() []slog.Attr {
	fields := ErrorAttributes(e.attributes).SlogAttrs()
	return fields
}

func (e *goError) Cause() error {
	return e.pkgError
}

func (e *goError) StackTrace() []uintptr {
	if stErr, ok := e.pkgError.(errStackTracer); ok {
		frames := stErr.StackTrace()
		st := make([]uintptr, 0, len(frames)+1)
		for _, f := range frames {
			st = append(st, uintptr(f))
		}
		return st
	}
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[:n]
}

func SlogErr(err error) slog.Attr {
	return slog.Group(
		"error",
		slog.String("message", err.Error()),
		slog.String("verbose", fmt.Sprintf("%+v", err)),
	)
}

func (e *goError) Format(s fmt.State, verb rune) {
	// todo:p1: add stack trace that was appeneded

	_, _ = fmt.Fprintf(s, "Attributes:\n")
	si := indent.New(s, "\t")
	for key, value := range e.ErrorAttributes() {
		valstr := fmt.Sprintf("%v", value)
		if len(valstr) > 1024 {
			valstr = valstr[:1024] + "..."
		}
		_, _ = fmt.Fprintf(si, "%s=%v \n", key, valstr)
	}
	if len(e.stack) > 0 {
		_, _ = fmt.Fprintf(s, "Stack:\n")
		_, _ = fmt.Fprintf(si, "%s\n", e.Error())
		for _, attr := range e.stack {
			_, _ = fmt.Fprintf(si, "%s: %s\n", attr.key, attr.str)
		}
	}

	if fe, ok := e.pkgError.(fmt.Formatter); ok {
		fe.Format(s, verb)
	} else {
		// note: pkgError always implements fmt.Formatter
		_, _ = io.WriteString(s, e.Error())
	}
}
