package transaction

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sync "github.com/sasha-s/go-deadlock"

	log "github.com/sirupsen/logrus"
)

var (
	logger = log.WithField("src", "transaction")
)

type Transaction struct {
	balances     map[string]int
	balancesLock *sync.Mutex
}

func NewTransaction() *Transaction {
	return &Transaction{
		balances:     map[string]int{},
		balancesLock: &sync.Mutex{},
	}
}

func (t *Transaction) Deposit(account string, amount int) (err error) {
	t.balancesLock.Lock()
	defer t.balancesLock.Unlock()

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

func (t *Transaction) Transfer(fromAccount string, toAccount string, amount int) (err error) {
	t.balancesLock.Lock()
	defer t.balancesLock.Unlock()

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
		logger.Infof("transfer failed, src account [%s] don't has enough funds, curr amount [%d], current amount [%d]", fromAccount, prevFromAccountAmount, amount)
		return fmt.Errorf("transfer failed, src account [%s] don't has enough funds, curr amount [%d], current amount [%d]", fromAccount, prevFromAccountAmount, amount)
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

func (t *Transaction) BalancesSnapshot() map[string]int {
	t.balancesLock.Lock()
	defer t.balancesLock.Unlock()
	balancesSnapshot := map[string]int{}
	for account, amount := range t.balances {
		balancesSnapshot[account] = amount
	}
	return balancesSnapshot
}

func (t *Transaction) BalancesSnapshotStdString() string {
	builder := &strings.Builder{}

	builder.WriteString("BALANCES")
	for account, amount := range t.BalancesSnapshot() {
		builder.WriteString(fmt.Sprintf(" %s:%d", account, amount))
	}

	return builder.String()
}

func (t *Transaction) BalancesSnapshotStdSortedString() string {
	builder := &strings.Builder{}

	builder.WriteString("BALANCES")

	balancesSnapshot := t.BalancesSnapshot()
	accounts := make([]string, 0)

	for account := range balancesSnapshot {
		accounts = append(accounts, account)
	}

	sort.Strings(accounts)

	for _, account := range accounts {
		builder.WriteString(fmt.Sprintf(" %s:%d", account, balancesSnapshot[account]))
	}

	return builder.String()
}
