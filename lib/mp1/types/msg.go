package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Hi struct {
	From string `json:"from"`
}

type Deposit struct {
	Account string `json:"account"`
	Amount  int    `json:"amount"`
}

type Transfer struct {
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      int    `json:"amount"`
}

type PolymorphicMsg struct {
	TypeID string `json:"type"`
	Body   string `json:"body"`
}

const (
	DepositEventTypeID = "deposit"
	TransferID         = "transfer"
)

var (
	PolymorphicMessageDecodeTable = map[string]func() interface{}{
		DepositEventTypeID: func() interface{} { return &Deposit{} },
		TransferID:         func() interface{} { return &Transfer{} },
	}
)

func EncodePolymorphicMessage(v interface{}, typeID string) (data []byte, err error) {
	data, err = json.Marshal(v)
	if err != nil {
		return nil, err
	}

	data, err = json.Marshal(
		&PolymorphicMsg{
			TypeID: typeID,
			Body:   string(data),
		},
	)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func DecodePolymorphicMessage(data []byte) (v interface{}, err error) {
	msg := &PolymorphicMsg{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}

	v, ok := PolymorphicMessageDecodeTable[msg.TypeID]
	if !ok {
		return nil, fmt.Errorf("unrecognized type: %s", msg.TypeID)
	}

	vf, _ := v.(func() interface{})
	o := vf()

	err = json.Unmarshal([]byte(msg.Body), o)
	if err != nil {
		return nil, err
	}

	return o, nil
}

type Account struct {
	Account string `json:"account"`
	Amount  int    `json:"amount"`
}

func NewHi(from string) *Hi {
	return &Hi{
		From: from,
	}
}

func (h *Hi) Encode() (data []byte, err error) {
	return json.Marshal(h)
}

const (
	DepositEvent  = "DEPOSIT"
	TransferEvent = "TRANSFER"
)

func EncodeTransactionsMsg(msg string) (encoded []byte, err error) {
	fields := strings.Fields(msg)
	if len(fields) == 0 {
		errMsg := "invalid event format, empty fields"
		return nil, fmt.Errorf(errMsg)
	}

	eventType := fields[0]

	switch eventType {
	case DepositEvent:
		if len(fields) != 3 {
			errMsg := "invalid deposit event format"
			return nil, fmt.Errorf(errMsg)
		}

		amount, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}

		deposit := Deposit{
			Account: fields[1],
			Amount:  amount,
		}

		encoded, err = EncodePolymorphicMessage(deposit, DepositEventTypeID)
		if err != nil {
			return nil, err
		}
		return encoded, nil
	case TransferEvent:
		if len(fields) != 5 {
			errMsg := "invalid transfer event format"
			return nil, fmt.Errorf(errMsg)
		}

		amount, err := strconv.Atoi(fields[4])
		if err != nil {
			return nil, err
		}

		transfer := Transfer{
			FromAccount: fields[1],
			ToAccount:   fields[3],
			Amount:      amount,
		}

		encoded, err = EncodePolymorphicMessage(transfer, TransferID)

		if err != nil {
			return nil, err
		}

		return encoded, nil
	default:
		errMsg := "unrecognized event type"
		return nil, fmt.Errorf(errMsg)
	}
}
