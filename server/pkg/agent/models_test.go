package agent

import (
	"context"
	"testing"
)

func TestListModelsStaticProviders(t *testing.T) {
	ctx := context.Background()
	for _, provider := range []string{"claude", "codex", "gemini", "cursor", "copilot"} {
		got, err := ListModels(ctx, provider, "")
		if err != nil {
			t.Fatalf("ListModels(%q) error: %v", provider, err)
		}
		if len(got) == 0 {
			t.Errorf("ListModels(%q) returned no models", provider)
		}
		for i, m := range got {
			if m.ID == "" {
				t.Errorf("ListModels(%q)[%d] has empty ID", provider, i)
			}
			if m.Label == "" {
				t.Errorf("ListModels(%q)[%d] has empty Label", provider, i)
			}
		}
	}
}

func TestGeminiStaticModelsExposesAliasesAndGemini3(t *testing.T) {
	// Gemini CLI has no `models list` subcommand, so we expose the
	// CLI's own aliases (auto / pro / flash / flash-lite) plus
	// explicit version pins including Gemini 3. Regression guard
	// for multica-ai/multica#1503 — Gemini 3 must be selectable.
	models := geminiStaticModels()
	ids := map[string]Model{}
	for _, m := range models {
		ids[m.ID] = m
	}
	for _, want := range []string{
		"auto", "auto-gemini-2.5",
		"pro", "flash", "flash-lite",
		"gemini-3-pro-preview", "gemini-3-flash-preview",
		"gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.5-flash-lite",
	} {
		if _, ok := ids[want]; !ok {
			t.Errorf("missing expected Gemini model %q in: %+v", want, models)
		}
	}
	auto, ok := ids["auto"]
	if !ok || !auto.Default {
		t.Errorf("expected `auto` to be the default Gemini entry, got %+v", auto)
	}
	for _, m := range models {
		if m.Provider != "google" {
			t.Errorf("all Gemini entries must carry Provider=google, got %+v", m)
		}
	}
}

func TestListModelsUnknownProvider(t *testing.T) {
	ctx := context.Background()
	_, err := ListModels(ctx, "nonexistent", "")
	if err == nil {
		t.Fatal("ListModels(unknown) expected error")
	}
}

func TestStaticCatalogsHaveAtMostOneDefault(t *testing.T) {
	// Each catalog should tag at most one entry as the display
	// default so the UI badge is unambiguous. More than one
	// usually means a copy/paste slip when adding new models.
	catalogs := map[string][]Model{
		"claude":  claudeStaticModels(),
		"codex":   codexStaticModels(),
		"gemini":  geminiStaticModels(),
		"cursor":  cursorStaticModels(),
		"copilot": copilotStaticModels(),
	}
	for provider, models := range catalogs {
		count := 0
		for _, m := range models {
			if m.Default {
				count++
			}
		}
		if count > 1 {
			t.Errorf("%s: %d models marked Default, want 0 or 1", provider, count)
		}
	}
}

func TestParseOpenCodeModels(t *testing.T) {
	input := `PROVIDER/MODEL                     CONTEXT  MAX_OUT
openai/gpt-4o                      128000   16384
anthropic/claude-sonnet-4-6        200000   8192
openai/gpt-4o                      128000   16384
nonprefixed-line
`
	models := parseOpenCodeModels(input)
	if len(models) != 2 {
		t.Fatalf("expected 2 models (header skipped, duplicate deduped, non-slash skipped), got %d: %+v", len(models), models)
	}
	if models[0].ID != "openai/gpt-4o" || models[0].Provider != "openai" {
		t.Errorf("unexpected first model: %+v", models[0])
	}
	if models[1].ID != "anthropic/claude-sonnet-4-6" || models[1].Provider != "anthropic" {
		t.Errorf("unexpected second model: %+v", models[1])
	}
}

func TestParsePiModels(t *testing.T) {
	input := `openai:gpt-4o
anthropic:claude-opus-4-7
openai:gpt-4o
bareword
`
	models := parsePiModels(input)
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d: %+v", len(models), models)
	}
	if models[0].ID != "openai/gpt-4o" {
		t.Errorf("expected colon normalized to slash: %+v", models[0])
	}
}

func TestParseCursorModels(t *testing.T) {
	input := `Available models

auto - Auto
composer-2-fast - Composer 2 Fast (current, default)
composer-2 - Composer 2
claude-4.6-sonnet-medium - Sonnet 4.6 1M
claude-opus-4-7-high - Opus 4.7 1M
gemini-3.1-pro - Gemini 3.1 Pro
`
	models := parseCursorModels(input)
	if len(models) != 6 {
		t.Fatalf("expected 6 models, got %d: %+v", len(models), models)
	}
	ids := map[string]Model{}
	for _, m := range models {
		ids[m.ID] = m
	}
	for _, want := range []string{"auto", "composer-2-fast", "composer-2", "claude-4.6-sonnet-medium", "claude-opus-4-7-high", "gemini-3.1-pro"} {
		if _, ok := ids[want]; !ok {
			t.Errorf("missing expected model %q in: %+v", want, models)
		}
	}
	if def := ids["composer-2-fast"]; !def.Default {
		t.Errorf("composer-2-fast should be marked default, got %+v", def)
	}
	if def := ids["composer-2-fast"]; def.Label != "Composer 2 Fast" {
		t.Errorf("default label should be stripped of parenthetical, got %q", def.Label)
	}
	// Non-default entry should not carry Default=true.
	if auto := ids["auto"]; auto.Default {
		t.Errorf("non-default entry should not be flagged default: %+v", auto)
	}
}

func TestParseCursorModelsSkipsHeaderAndBlankLines(t *testing.T) {
	input := `Available models

composer-2 - Composer 2
`
	models := parseCursorModels(input)
	if len(models) != 1 || models[0].ID != "composer-2" {
		t.Fatalf("unexpected: %+v", models)
	}
}

func TestParseACPSessionNewModels(t *testing.T) {
	// Mirrors the model catalog shape emitted by an ACP session/new response.
	raw := []byte(`{
      "sessionId": "ses_123",
      "models": {
        "availableModels": [
          {"modelId": "nous:moonshotai/kimi-k2.5", "name": "moonshotai/kimi-k2.5", "description": "Provider: Nous"},
          {"modelId": "nous:anthropic/claude-opus-4.7", "name": "anthropic/claude-opus-4.7", "description": "Provider: Nous • current"},
          {"modelId": "nous:moonshotai/kimi-k2.5", "name": "duplicate", "description": "dup"}
        ],
        "currentModelId": "nous:anthropic/claude-opus-4.7"
      }
    }`)
	models := parseACPSessionNewModels(raw)
	if len(models) != 2 {
		t.Fatalf("expected 2 models (duplicate deduped), got %d: %+v", len(models), models)
	}
	if models[0].ID != "nous:moonshotai/kimi-k2.5" || models[0].Provider != "nous" {
		t.Errorf("unexpected first model: %+v", models[0])
	}
	if models[0].Default {
		t.Errorf("non-current entry must not be marked default: %+v", models[0])
	}
	if !models[1].Default {
		t.Errorf("current entry must be marked default: %+v", models[1])
	}
	if models[1].ID != "nous:anthropic/claude-opus-4.7" {
		t.Errorf("expected current model second: %+v", models[1])
	}
}

func TestParseACPSessionNewModelsMissingField(t *testing.T) {
	// session/new without the models field should yield nil so the caller can
	// distinguish "no catalog" from "empty catalog".
	raw := []byte(`{"sessionId": "ses_123"}`)
	if got := parseACPSessionNewModels(raw); got != nil && len(got) != 0 {
		t.Errorf("expected nil/empty, got %+v", got)
	}
}

func TestParseACPSessionNewModelsGarbage(t *testing.T) {
	if got := parseACPSessionNewModels([]byte("not json")); got != nil {
		t.Errorf("expected nil for non-JSON, got %+v", got)
	}
}

func TestCachedDiscovery(t *testing.T) {
	calls := 0
	fn := func() ([]Model, error) {
		calls++
		return []Model{{ID: "x", Label: "x"}}, nil
	}
	// First call populates the cache; reset for isolation.
	modelCacheMu.Lock()
	delete(modelCache, "testkey")
	modelCacheMu.Unlock()

	if _, err := cachedDiscovery("testkey", fn); err != nil {
		t.Fatal(err)
	}
	if _, err := cachedDiscovery("testkey", fn); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("expected 1 underlying call due to cache, got %d", calls)
	}
}
