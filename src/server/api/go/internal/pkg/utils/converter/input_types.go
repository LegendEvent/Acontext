package converter

// Input part types for different message formats

// OpenAI Message Content Types (for input normalization)
// Reference: https://platform.openai.com/docs/api-reference/messages

type OpenAIPartIn struct {
	Type string `json:"type" validate:"required,oneof=text image_url input_audio tool_call tool_result"`

	// For text type
	Text string `json:"text,omitempty"`

	// For image_url type
	ImageURL *OpenAIImageURL `json:"image_url,omitempty"`

	// For input_audio type
	InputAudio *OpenAIInputAudio `json:"input_audio,omitempty"`

	// For tool_call type (assistant messages)
	ID       string              `json:"id,omitempty"`
	Function *OpenAIFunctionCall `json:"function,omitempty"`

	// For tool_result type (user messages)
	ToolCallID string `json:"tool_call_id,omitempty"`
	Output     string `json:"output,omitempty"`

	// For file uploads (multipart)
	FileField string `json:"file_field,omitempty"`
}

type OpenAIInputAudio struct {
	Data   string `json:"data,omitempty"`   // base64 encoded audio
	Format string `json:"format,omitempty"` // "wav", "mp3"
}

// Anthropic Content Block Types (for input normalization)
// Reference: https://docs.anthropic.com/en/api/messages

type AnthropicPartIn struct {
	Type string `json:"type" validate:"required,oneof=text image tool_use tool_result"`

	// For text type
	Text string `json:"text,omitempty"`

	// For image type
	Source *AnthropicImageSource `json:"source,omitempty"`

	// For tool_use type
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`

	// For tool_result type
	ToolUseID string      `json:"tool_use_id,omitempty"`
	Content   interface{} `json:"content,omitempty"` // string or array of content blocks
	IsError   bool        `json:"is_error,omitempty"`

	// For file uploads (multipart)
	FileField string `json:"file_field,omitempty"`
}
