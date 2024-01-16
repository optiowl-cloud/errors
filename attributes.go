package errors

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"

	"log/slog"
)

type AttributeType uint8

const (
	UnknownAttributeType AttributeType = iota
	StringAttributeType
	Uint64AttributeType
	Int64AttributeType
	StackAttributeType
	StringPtrAttributeType
	NoOpStackAttributeType
)

type ErrorAttributes []ErrorAttribute

func (attrs ErrorAttributes) SlogAttrs() []slog.Attr {
	fields := make([]slog.Attr, 0, len(attrs)+1)
	for _, attr := range attrs {
		switch attr.attrType {
		case UnknownAttributeType:
			fields = append(fields, slog.Any(attr.key, attr.value))
		case StringAttributeType:
			fields = append(fields, slog.String(attr.key, attr.str))
		case Uint64AttributeType:
			fields = append(fields, slog.Uint64(attr.key, attr.number))
		case Int64AttributeType:
			fields = append(fields, slog.Int64(attr.key, int64(attr.number)))
		case StringPtrAttributeType:
			fields = append(fields, slog.Any(attr.key, attr.strPtr))
		}
	}
	errContextStack := &bytes.Buffer{}
	for _, attr := range attrs {
		if attr.attrType == StackAttributeType {
			errContextStack.WriteString(attr.str)
			errContextStack.WriteByte('\n')
		}
	}
	fields = append(fields, slog.String("error_context_stack", errContextStack.String()))
	return fields
}

func (attrs ErrorAttributes) ErrorAttributes() map[string]interface{} {
	attrmap := make(map[string]interface{}, len(attrs))
	for _, attr := range attrs {
		switch attr.attrType {
		case UnknownAttributeType:
			attrmap[attr.key] = attr.value
		case StringAttributeType:
			attrmap[attr.key] = attr.str
		case StringPtrAttributeType:
			attrmap[attr.key] = attr.strPtr
		case Uint64AttributeType:
			attrmap[attr.key] = attr.number
		case Int64AttributeType:
			attrmap[attr.key] = int64(attr.number)
		}
	}
	return attrmap
}

type ErrorAttribute struct {
	value    interface{}
	key      string
	str      string
	strPtr   *string
	number   uint64
	attrType AttributeType
}

var _ slogAttrs = ErrorAttributes{}

func Stringf(key, format string, args ...interface{}) ErrorAttribute {
	return String(key, fmt.Sprintf(format, args...))
}

func KeyedStackf(key, format string, args ...interface{}) ErrorAttribute {
	return ErrorAttribute{
		attrType: StackAttributeType,
		key:      key,
		str:      fmt.Sprintf(format, args...),
	}
}

func Stackf(format string, args ...interface{}) ErrorAttribute {
	pc, _, _, _ := runtime.Caller(1)
	funcObj := runtime.FuncForPC(pc)
	key := "unknown"
	if funcObj != nil {
		key = funcObj.Name()
		if idx := strings.LastIndex(key, "/"); idx != -1 {
			key = key[idx+1:]
		}
	}
	return KeyedStackf(key, format, args...)
}

func String(key, value string) ErrorAttribute {
	return ErrorAttribute{
		attrType: StringAttributeType,
		key:      key,
		str:      value,
	}
}

func StringPtr(key string, value *string) ErrorAttribute {
	return ErrorAttribute{
		attrType: StringPtrAttributeType,
		key:      key,
		strPtr:   value,
	}
}

func Int64(key string, value int64) ErrorAttribute {
	return ErrorAttribute{
		attrType: Int64AttributeType,
		key:      key,
		number:   uint64(value), // cast to uint64 and uncast when used
	}
}

func Uint64(key string, value uint64) ErrorAttribute {
	return ErrorAttribute{
		attrType: Uint64AttributeType,
		key:      key,
		number:   value,
	}
}

func Int(key string, value int) ErrorAttribute {
	return ErrorAttribute{
		attrType: Int64AttributeType,
		key:      key,
		number:   uint64(value),
	}
}
