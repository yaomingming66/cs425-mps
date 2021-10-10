package transaction

import (
	"bufio"
	"io"
	"os"

	"github.com/bamboovir/cs425/lib/mp1/router"
	log "github.com/sirupsen/logrus"
)

var (
	transactionEventListenerLogger = log.WithField("src", "event_listener")
)

func TransactionEventListenerPipeline(reader io.Reader) <-chan *router.Msg {
	out := make(chan *router.Msg, 100000)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			eventMsg, err := EncodeTransactionsMsg(line)
			if err != nil {
				transactionEventListenerLogger.Errorf("encode input msg failed with err :%v, skip", err)
				continue
			}
			out <- eventMsg
		}

		err := scanner.Err()

		if err != nil {
			transactionEventListenerLogger.Errorf("read err: %v", err)
		} else {
			transactionEventListenerLogger.Info("reach EOF")
		}
		close(out)
	}()

	return out
}
