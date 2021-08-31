package main

import (
	"os"

	"github.com/bamboovir/cs425/cmd/mp0/client"
	"github.com/bamboovir/cs425/lib/logger"
	log "github.com/sirupsen/logrus"
)

func main() {
	logger.SetupLogger(log.StandardLogger())
	rootCMD := client.NewRootCMD()
	if err := rootCMD.Execute(); err != nil {
		log.Errorf("%v\n", err)
		os.Exit(1)
	}
}
