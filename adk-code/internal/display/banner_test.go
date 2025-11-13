package display

import (
	"strings"
	"testing"
)

func TestNewBannerRenderer(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	br := NewBannerRenderer(r)
	if br == nil {
		t.Fatal("expected non-nil banner renderer")
	}
}

func TestBannerRenderer_RenderWelcome(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	br := NewBannerRenderer(r)
	result := br.RenderWelcome()
	if result == "" {
		t.Fatal("expected non-empty welcome message")
	}
}

func TestBannerRenderer_RenderStartBanner(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	br := NewBannerRenderer(r)
	result := br.RenderStartBanner("1.0.0", "test-model", "/test/dir")
	if result == "" {
		t.Fatal("expected non-empty start banner")
	}
	if !strings.Contains(result, "1.0.0") {
		t.Fatalf("expected version in banner, got: %s", result)
	}
}

func TestBannerRenderer_RenderStartBanner_WithModel(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	br := NewBannerRenderer(r)
	result := br.RenderStartBanner("2.0.0", "gpt-4", "/home/user")
	if !strings.Contains(result, "2.0.0") || !strings.Contains(result, "gpt-4") {
		t.Fatalf("expected version and model in banner, got: %s", result)
	}
}

func TestBannerRenderer_RenderStartBanner_WithPath(t *testing.T) {
	r, _ := NewRenderer(OutputFormatPlain)
	br := NewBannerRenderer(r)
	result := br.RenderStartBanner("1.5.0", "test", "/mydir")
	if !strings.Contains(result, "/mydir") {
		t.Fatalf("expected path in banner, got: %s", result)
	}
}
