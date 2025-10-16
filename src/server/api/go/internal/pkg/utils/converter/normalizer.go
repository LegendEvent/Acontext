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
	case FormatNone, "": // Internal format, no normalization needed
		return &NoOpNormalizer{}, nil
	case FormatOpenAI:
		return &OpenAINormalizer{}, nil
	case FormatAnthropic:
		return &AnthropicNormalizer{}, nil
	case FormatLangChain:
		return &LangChainNormalizer{}, nil
	default:
		return nil, fmt.Errorf("unsupported input format: %s", format)
	}
}

// NoOpNormalizer does no transformation (for internal format)
type NoOpNormalizer struct{}

func (n *NoOpNormalizer) Normalize(role string, parts []service.PartIn) (string, []service.PartIn, error) {
	return role, parts, nil
}

// OpenAINormalizer normalizes OpenAI format to internal format
type OpenAINormalizer struct{}

func (n *OpenAINormalizer) Normalize(role string, parts []service.PartIn) (string, []service.PartIn, error) {
	// OpenAI roles map directly: user, assistant, system, tool, function
	// These are already compatible with our internal format

	// Validate role
	validRoles := map[string]bool{
		"user": true, "assistant": true, "system": true,
		"tool": true, "function": true,
	}
	if !validRoles[role] {
		return "", nil, fmt.Errorf("invalid OpenAI role: %s", role)
	}

	// OpenAI parts are also compatible, but we need to ensure tool-call parts have correct structure
	normalizedParts := make([]service.PartIn, len(parts))
	for i, part := range parts {
		normalizedParts[i] = part

		// Ensure tool-call parts have 'id' field in meta
		if part.Type == "tool-call" && part.Meta != nil {
			// OpenAI format uses 'id', 'name' (tool name), 'arguments'
			// Our internal format expects 'id', 'tool_name', 'arguments'
			if name, ok := part.Meta["name"].(string); ok && part.Meta["tool_name"] == nil {
				normalizedParts[i].Meta["tool_name"] = name
			}
		}
	}

	return role, normalizedParts, nil
}

// AnthropicNormalizer normalizes Anthropic format to internal format
type AnthropicNormalizer struct{}

func (n *AnthropicNormalizer) Normalize(role string, parts []service.PartIn) (string, []service.PartIn, error) {
	// Anthropic only has "user" and "assistant" roles
	// System messages should be sent via system parameter, not as messages
	validRoles := map[string]bool{
		"user": true, "assistant": true,
	}
	if !validRoles[role] {
		return "", nil, fmt.Errorf("invalid Anthropic role: %s (only 'user' and 'assistant' are supported)", role)
	}

	normalizedParts := make([]service.PartIn, 0, len(parts))

	for _, part := range parts {
		switch part.Type {
		case "text", "image":
			// Direct mapping
			normalizedParts = append(normalizedParts, part)

		case "tool-call":
			// Anthropic calls this "tool_use" in content blocks
			// Ensure it has the required fields
			if part.Meta == nil {
				return "", nil, fmt.Errorf("tool-call part missing meta")
			}

			normalizedPart := part
			// Anthropic uses 'name' for tool name, we use 'tool_name'
			if name, ok := part.Meta["name"].(string); ok && part.Meta["tool_name"] == nil {
				normalizedPart.Meta["tool_name"] = name
			}
			// Anthropic uses 'input' for arguments, we use 'arguments'
			if input, ok := part.Meta["input"]; ok && part.Meta["arguments"] == nil {
				normalizedPart.Meta["arguments"] = input
			}
			normalizedParts = append(normalizedParts, normalizedPart)

		case "tool-result":
			// Anthropic calls this "tool_result"
			// Map tool_use_id to tool_call_id
			if part.Meta == nil {
				return "", nil, fmt.Errorf("tool-result part missing meta")
			}

			normalizedPart := part
			if toolUseID, ok := part.Meta["tool_use_id"].(string); ok && part.Meta["tool_call_id"] == nil {
				normalizedPart.Meta["tool_call_id"] = toolUseID
			}
			normalizedParts = append(normalizedParts, normalizedPart)

		default:
			// Other types pass through
			normalizedParts = append(normalizedParts, part)
		}
	}

	return role, normalizedParts, nil
}

// LangChainNormalizer normalizes LangChain format to internal format
type LangChainNormalizer struct{}

func (n *LangChainNormalizer) Normalize(role string, parts []service.PartIn) (string, []service.PartIn, error) {
	// LangChain uses: human->user, ai->assistant, system->system, tool->tool
	normalizedRole := role
	switch role {
	case "human":
		normalizedRole = "user"
	case "ai":
		normalizedRole = "assistant"
	case "user", "assistant", "system", "tool", "function":
		// Already valid
	default:
		return "", nil, fmt.Errorf("invalid LangChain role: %s", role)
	}

	// LangChain parts structure is similar to ours
	return normalizedRole, parts, nil
}
