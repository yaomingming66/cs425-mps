package transaction

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bamboovir/cs425/lib/mp1/router"
)

const (
	DepositEvent  = "DEPOSIT"
	TransferEvent = "TRANSFER"
)

type Deposit struct {
	Account string `json:"account"`
	Amount  int    `json:"amount"`
}

func (d *Deposit) Encode() (data []byte, err error) {
	data, err = json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *Deposit) Decode(data []byte) (*Deposit, error) {
	err := json.Unmarshal(data, d)
	if err != nil {
		return d, err
	}
	return d, nil
}

type Transfer struct {
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      int    `json:"amount"`
}

func (t *Transfer) Encode() (data []byte, err error) {
	data, err = json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (t *Transfer) Decode(data []byte) (*Transfer, error) {
	err := json.Unmarshal(data, t)
	if err != nil {
		return t, err
	}
	return t, nil
}

const (
	DepositEventTypeID = "deposit"
	TransferID         = "transfer"
	DepositPath        = "/transaction/deposit"
	TransferPath       = "/transaction/transfer"
)

func EncodeTransactionsMsg(msg string) (dmsg *router.Msg, err error) {
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

		dmsg := router.NewMsg(DepositPath, deposit)

		return dmsg, nil
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

		dmsg := router.NewMsg(TransferPath, transfer)

		return dmsg, nil
	default:
		errMsg := "unrecognized event type"
		return nil, fmt.Errorf(errMsg)
	}
}
