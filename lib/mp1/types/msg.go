package types

import (
	"encoding/json"
)

type Hi struct {
	From string `json:"from"`
}

func NewHi(from string) *Hi {
	return &Hi{
		From: from,
	}
}

func (h *Hi) Encode() (data []byte, err error) {
	return json.Marshal(h)
}
