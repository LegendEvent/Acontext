package converter

import (
	"fmt"

	"github.com/memodb-io/Acontext/internal/modules/service"
)

// MessageNormalizer normalizes input messages from different formats to internal format
type MessageNormalizer interface {
	// Normalize converts format-specific role and parts to internal representation
	Normalize(role string, parts []service.PartIn) (string, []service.PartIn, error)
}

// GetNormalizer returns the appropriate normalizer for the given format
func GetNormalizer(format MessageFormat) (MessageNormalizer, error) {
	switch format {
	case FormatAcontext, "": // Internal format, no normalization needed
		return &NoOpNormalizer{}, nil
	case FormatOpenAI:
		return &OpenAINormalizer{}, nil
	case FormatAnthropic:
		return &AnthropicNormalizer{}, nil
	default:
		return nil, fmt.Errorf("unsupported input format: %s", format)
	}
}

// NoOpNormalizer does no transformation (for internal format)
type NoOpNormalizer struct{}

func (n *NoOpNormalizer) Normalize(role string, parts []service.PartIn) (string, []service.PartIn, error) {
	return role, parts, nil
}
