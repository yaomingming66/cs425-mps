package transaction

import (
	"github.com/bamboovir/cs425/lib/mp1/multicast"
)

type Processor struct {
	transaction *Transaction
}

func NewProcessor() *Processor {
	return &Processor{
		transaction: NewTransaction(),
	}
}

func toProcessWrapper(f func(msg []byte)) func(msg multicast.TOMsg) {
	return func(msg multicast.TOMsg) {
		f(msg.Body)
	}
}

func (p *Processor) RegisteTransactionHandler(d *multicast.TODispatcher) {
	d.Bind(DepositPath, toProcessWrapper(p.processDeposit))
	d.Bind(TransferPath, toProcessWrapper(p.processTransfer))
}

func (p *Processor) processDeposit(msg []byte) {
	deposit := &Deposit{}
	_, err := deposit.Decode(msg)
	if err != nil {
		logger.Errorf("deposit err: %v", err)
		return
	}

	logger.Infof("deposit: %s -> %d", deposit.Account, deposit.Amount)
	err = p.transaction.Deposit(deposit.Account, deposit.Amount)
	if err != nil {
		logger.Errorf("deposit err: %v", err)
		return
	}
	logger.Infof(p.transaction.BalancesSnapshotStdSortedString())
}

func (p *Processor) processTransfer(msg []byte) {
	transfer := &Transfer{}
	_, err := transfer.Decode(msg)
	if err != nil {
		logger.Errorf("transfer err: %v", err)
		return
	}

	logger.Infof("tranfer: %s -> %s %d", transfer.FromAccount, transfer.ToAccount, transfer.Amount)
	err = p.transaction.Transfer(transfer.FromAccount, transfer.ToAccount, transfer.Amount)
	if err != nil {
		logger.Errorf("transfer err: %v", err)
		return
	}
	logger.Info(p.transaction.BalancesSnapshotStdSortedString())
}
