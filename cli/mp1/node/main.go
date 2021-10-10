package main

import (
	"os"
	"time"

	"github.com/bamboovir/cs425/cmd/mp1"
	"github.com/bamboovir/cs425/lib/logger"
	sync "github.com/sasha-s/go-deadlock"
	log "github.com/sirupsen/logrus"
)

func main() {
	sync.Opts.DeadlockTimeout = time.Second * 100
	logger.SetupLogger(log.StandardLogger())
	rootCMD := mp1.NewRootCMD()
	if err := rootCMD.Execute(); err != nil {
		log.Errorf("%v\n", err)
		os.Exit(1)
	}
}
