package transaction

import (
	"fmt"

	"github.com/bamboovir/cs425/lib/mp1/multicast"
	"github.com/pkg/errors"
)

type Processor struct {
	transaction *Transaction
}

func NewProcessor() *Processor {
	return &Processor{
		transaction: NewTransaction(),
	}
}

func (p *Processor) RegisteTransactionHandler(d *multicast.TotalOrding) {
	d.Bind(DepositPath, p.processDeposit)
	d.Bind(TransferPath, p.processTransfer)
}

func (p *Processor) processDeposit(msg *multicast.TOMsg) error {
	deposit := &Deposit{}
	_, err := deposit.Decode(msg.Body)
	if err != nil {
		return errors.Wrap(err, "process deposit failed")
	}

	// logger.Infof("deposit: %s -> %d", deposit.Account, deposit.Amount)
	fmt.Printf("DEPOSIT %s %d\n", deposit.Account, deposit.Amount)
	err = p.transaction.Deposit(deposit.Account, deposit.Amount)
	if err != nil {
		return errors.Wrap(err, "process deposit failed")
	}
	snapshot := p.transaction.BalancesSnapshotStdSortedString()
	// logger.Info(snapshot)
	fmt.Printf("%s\n", snapshot)
	return nil
}

func (p *Processor) processTransfer(msg *multicast.TOMsg) error {
	transfer := &Transfer{}
	_, err := transfer.Decode(msg.Body)
	if err != nil {
		return errors.Wrap(err, "process transfer failed")
	}

	// logger.Infof("tranfer: %s -> %s %d", transfer.FromAccount, transfer.ToAccount, transfer.Amount)
	fmt.Printf("TRANSFER %s %s %d\n", transfer.FromAccount, transfer.ToAccount, transfer.Amount)
	err = p.transaction.Transfer(transfer.FromAccount, transfer.ToAccount, transfer.Amount)
	if err != nil {
		return errors.Wrap(err, "process transfer failed")
	}
	snapshot := p.transaction.BalancesSnapshotStdSortedString()
	logger.Info(snapshot)
	// fmt.Printf("%s\n", snapshot)
	return nil
}
