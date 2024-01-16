package errors

type ErrorMessage string

var _ error = ErrorMessage("")

func (e ErrorMessage) Error() string {
	if e == "" {
		return "[empty error message]"
	}
	return string(e)
}
