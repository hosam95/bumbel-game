package structs

type Message struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}
