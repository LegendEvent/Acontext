package types

type PartIn struct {
	Type      string                 `json:"type"`                 // "text" | "image" | ...
	Text      string                 `json:"text,omitempty"`       // Text sharding
	FileField string                 `json:"file_field,omitempty"` // File field name in the form
	Meta      map[string]interface{} `json:"meta,omitempty"`       // [Optional] metadata
}
