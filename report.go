package errors

import (
	"context"

	"log/slog"
)

type Severity string

func (s Severity) String() string {
	return string(s)
}

const (
	SeverityWarning Severity = "Warning"
	SeverityError   Severity = "Error"
)

func ReportWithSeverity(ctx context.Context, err error, severity Severity) {
	if severity != SeverityError && severity != SeverityWarning {
		severity = SeverityError
	}

	err = Wrap(
		err,
		String("severity", severity.String()),
	)

	slog.ErrorContext(
		ctx, "error: "+err.Error(),
		SlogErr(err), slog.String("severity", severity.String()),
	)
}

func Report(ctx context.Context, err error) {
	ReportWithSeverity(ctx, err, SeverityError)
}
