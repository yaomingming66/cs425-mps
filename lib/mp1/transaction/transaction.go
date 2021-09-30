package transaction

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	logger = log.WithField("src", "transaction")
)

type TransactionProcessor struct {
	balances map[string]int
}

func NewTransactionProcessor() *TransactionProcessor {
	return &TransactionProcessor{
		balances: map[string]int{},
	}
}

func (t *TransactionProcessor) Deposit(account string, amount int) (err error) {
	if amount < 0 {
		logger.Errorf("amount should be a integer greater or equal to zero")
		return errors.New("amount should be a integer greater or equal to zero")
	}
	prevAmount, ok := t.balances[account]
	if !ok {
		logger.Infof("account [%s] not exists, create new account [%s] with amount [%d]", account, account, amount)
		t.balances[account] = amount
		return nil
	}

	logger.Infof("account [%s] exists, add account [%s] with amount [%d]", account, account, amount)
	t.balances[account] = prevAmount + amount
	return nil
}

func (t *TransactionProcessor) Transfer(fromAccount string, toAccount string, amount int) (err error) {
	if amount < 0 {
		logger.Errorf("amount should be a integer greater or equal to zero")
		return errors.New("amount should be a integer greater or equal to zero")
	}
	prevFromAccountAmount, isFromAccountExist := t.balances[fromAccount]
	prevToAccountAmount, isToAccountExist := t.balances[toAccount]

	if !isFromAccountExist {
		logger.Infof("transfer failed, src account [%s] not exists", fromAccount)
		return fmt.Errorf("transfer failed, src account [%s] not exists", fromAccount)
	}

	if prevFromAccountAmount-amount < 0 {
		logger.Infof("transfer failed, src account [%s] don't has enough funds, curr amount [%d], need amount [%d]", fromAccount, prevFromAccountAmount, amount)
		return fmt.Errorf("transfer failed, src account [%s] don't has enough funds, curr amount [%d], need amount [%d]", fromAccount, prevFromAccountAmount, amount)
	}

	t.balances[fromAccount] = prevFromAccountAmount - amount

	if !isToAccountExist {
		logger.Infof("dst account [%s] not exists, create new account [%s] with amount [%d]", toAccount, toAccount, amount)
		t.balances[toAccount] = amount
		return nil
	}

	logger.Infof("account [%s] exists, add account [%s] with amount [%d]", toAccount, toAccount, amount)
	t.balances[toAccount] = prevToAccountAmount + amount
	return nil
}

func (t *TransactionProcessor) BalancesSnapshot() map[string]int {
	balancesSnapshot := map[string]int{}
	for account, amount := range t.balances {
		balancesSnapshot[account] = amount
	}
	return balancesSnapshot
}

func (t *TransactionProcessor) BalancesSnapshotStdString() string {
	builder := &strings.Builder{}

	builder.WriteString("BALANCES")
	for account, amount := range t.BalancesSnapshot() {
		builder.WriteString(fmt.Sprintf(" %s:%d", account, amount))
	}

	return builder.String()
}