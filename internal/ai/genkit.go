// Package ai handles Genkit initialization and plugin configuration.
package ai

import (
	"context"
	"fmt"

	"github.com/FPT-OJT/minstant-ai.git/internal/config"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai"
)

// NewGenkit initializes and returns a Genkit instance configured with the
// OpenAI-compatible plugin. The plugin is set up using the provided AIConfig
// (API key, base URL, and default model).
func NewGenkit(ctx context.Context, cfg config.AIConfig) (*genkit.Genkit, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	plugin := compat_oai.OpenAICompatible{
		Provider: "openai-compat",
		APIKey:   cfg.APIKey,
		BaseURL:  cfg.BaseURL,
	}

	g := genkit.Init(ctx,
		genkit.WithPlugins(&plugin),
		genkit.WithDefaultModel(fmt.Sprintf("openai-compat/%s", cfg.Model)),
	)

	return g, nil
}
