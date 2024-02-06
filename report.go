package errors

import (
	"context"
	"fmt"
	"os"

	"log/slog"
)

var reportErrorToHuman bool = os.Getenv("REPORT_ERROR_TO_HUMAN") == "true"

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

	if reportErrorToHuman {
		fmt.Printf("error: %s\nseverity: %s\n%+v\n", err.Error(), severity.String(), err)
	}
}

func Report(ctx context.Context, err error) {
	ReportWithSeverity(ctx, err, SeverityError)
}
