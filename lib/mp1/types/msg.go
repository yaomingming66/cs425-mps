package types

import (
	"encoding/json"
	"fmt"
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

type Balances struct {
	Accounts []Account
}

func (b *Balances) Encode() string {
	builder := &strings.Builder{}

	builder.WriteString("BALANCES")
	for i := range b.Accounts {
		account := b.Accounts[i]
		builder.WriteString(fmt.Sprintf(" %s:%d", account.Account, account.Amount))
	}

	return builder.String()
}
