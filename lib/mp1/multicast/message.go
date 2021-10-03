package multicast

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Msg struct {
	SrcID string `json:"src"`
	Body  []byte `json:"body"`
}

func EncodeMsg(msg *Msg) (data []byte, err error) {
	data, err = json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func DecodeMsg(data []byte) (msg *Msg, err error) {
	msg = &Msg{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

type RMsg struct {
	ID   string `json:"id"`
	Body []byte `json:"body"`
}

func NewRMsg(body []byte) *RMsg {
	id := uuid.New().String()
	return &RMsg{
		ID:   id,
		Body: body,
	}
}

func EncodeRMsg(msg *RMsg) (data []byte, err error) {
	data, err = json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func DecodeRMsg(data []byte) (msg *RMsg, err error) {
	msg = &RMsg{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
