package model

type BodyMsg struct {
	Messages struct {
		Success []string `json:"success"`
		Error   []string `json:"error"`
	}
	Points []float64 `json:"points"`
}
