package app

import (
	"encoding/json"
	"fmt"
	"io"
)

var marshalJSON = json.Marshal //nolint:gochecknoglobals // JSON encode test seam

func writeJSON(out io.Writer, value any) error {
	data, err := loadSeam(&marshalJSON)(value)
	if err != nil {
		return fmt.Errorf("编码 JSON: %w", err)
	}

	data = append(data, '\n')
	if _, err := out.Write(data); err != nil {
		return fmt.Errorf("写入 stdout: %w", err)
	}

	return nil
}

func writeText(out io.Writer, line string) error {
	if _, err := fmt.Fprintln(out, line); err != nil {
		return fmt.Errorf("写入 stdout: %w", err)
	}

	return nil
}

func writeErr(stderr io.Writer, err error, code int) int {
	if _, writeErr := fmt.Fprintf(stderr, "nsk: %v\n", err); writeErr != nil {
		return exitFailure
	}

	return code
}
