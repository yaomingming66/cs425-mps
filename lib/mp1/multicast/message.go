package multicast

import (
	"encoding/json"

	"github.com/google/uuid"
)

type BMsg struct {
	SrcID string `json:"src"`
	Path  string `json:"path"`
	Body  []byte `json:"body"`
}

func NewBMsg(srcID string, path string, v interface{}) (msg *BMsg, err error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return &BMsg{
		SrcID: srcID,
		Path:  path,
		Body:  data,
	}, nil
}

func (m *BMsg) Encode() (data []byte, err error) {
	data, err = json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *BMsg) Decode(data []byte) (msg *BMsg, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

type RMsg struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Body []byte `json:"body"`
}

func NewRMsg(path string, body []byte) *RMsg {
	id := uuid.New().String()
	return &RMsg{
		ID:   id,
		Path: path,
		Body: body,
	}
}

func (m *RMsg) Encode() (data []byte, err error) {
	data, err = json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *RMsg) Decode(data []byte) (msg *RMsg, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
