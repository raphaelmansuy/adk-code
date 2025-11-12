// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
