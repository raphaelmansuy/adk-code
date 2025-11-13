package formatters_test

import (
	"testing"

	"adk-code/internal/display/formatters"
	"adk-code/internal/display/styles"
	pkgerrors "adk-code/pkg/errors"
)

// A minimal MarkdownRenderer stub for registering formatters
type stubMarkdownRenderer struct{}

func (s *stubMarkdownRenderer) Render(str string) (string, error) { return str, nil }

func TestRegisterCustomFormatterDuplicate(t *testing.T) {
	fr := formatters.NewFormatterRegistry("text", &styles.Styles{}, &styles.Formatter{}, &stubMarkdownRenderer{})
	// register first
	if err := fr.RegisterCustomFormatter("duplicate", nil); err != nil {
		t.Fatalf("unexpected error registering formatter: %v", err)
	}
	// register duplicate
	err := fr.RegisterCustomFormatter("duplicate", nil)
	if err == nil {
		t.Fatalf("expected error when registering duplicate formatter")
	}
	if !pkgerrors.Is(err, pkgerrors.CodeInvalidInput) {
		t.Fatalf("expected CodeInvalidInput error, got: %v", err)
	}
}
