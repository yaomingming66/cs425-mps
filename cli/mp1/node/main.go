package main

import (
	"os"

	"github.com/bamboovir/cs425/cmd/mp1"
	"github.com/bamboovir/cs425/lib/logger"
	log "github.com/sirupsen/logrus"
)

func main() {
	logger.SetupLogger(log.StandardLogger())
	rootCMD := mp1.NewRootCMD()
	if err := rootCMD.Execute(); err != nil {
		log.Errorf("%v\n", err)
		os.Exit(1)
	}
}
