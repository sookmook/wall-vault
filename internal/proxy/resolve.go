package proxy

import (
	"errors"
	"fmt"

	"github.com/sookmook/wall-vault/internal/vault"
)

// ErrModelNotAllowed is returned when Client.ModelOverride is non-empty and
// the service has a non-empty AllowedModels whitelist that does NOT include
// the override.
var ErrModelNotAllowed = errors.New("model_override not in service.allowed_models")

// ResolveModel picks which model name to send to a service for a given
// client. Per spec §2.1:
//
//   override == ""                              → service.DefaultModel
//   override != "" && len(allowed_models) == 0  → override (unrestricted)
//   override != "" && override ∈ allowed_models → override (whitelisted)
//   override != "" && override ∉ allowed_models → ErrModelNotAllowed
func ResolveModel(c vault.Client, s vault.Service) (string, error) {
	if c.ModelOverride == "" {
		return s.DefaultModel, nil
	}
	if len(s.AllowedModels) == 0 {
		return c.ModelOverride, nil
	}
	for _, m := range s.AllowedModels {
		if m == c.ModelOverride {
			return c.ModelOverride, nil
		}
	}
	return "", fmt.Errorf("%w: override=%q service=%q",
		ErrModelNotAllowed, c.ModelOverride, s.ID)
}
