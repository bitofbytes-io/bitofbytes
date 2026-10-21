package metrics

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestDisabledAddressVariantsDoNotBind(t *testing.T) {
	for _, addr := range []string{"", " ", "off", "OFF", " OfF ", "\toFf\n"} {
		var logs bytes.Buffer
		New().Serve(context.Background(), addr, slog.New(slog.NewTextHandler(&logs, nil)))
		if !strings.Contains(logs.String(), "Metrics listener disabled") || strings.Contains(logs.String(), "ERROR") {
			t.Fatalf("%q did not take the disabled path: %s", addr, logs.String())
		}
	}
}
