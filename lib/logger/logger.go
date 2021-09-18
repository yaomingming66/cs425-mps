package logger

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func SetupLogger(logger *log.Logger) *log.Logger {
	logger.SetLevel(log.TraceLevel)
	logEnv := os.Getenv("LOG")
	if logEnv == "trace" {
		logger.SetReportCaller(true)
	} else {
		logger.SetReportCaller(false)
	}

	if logEnv == "json" {
		log.SetFormatter(&log.JSONFormatter{
			DisableTimestamp: true,
			PrettyPrint:      false,
		})
	} else {
		logger.SetFormatter(&log.TextFormatter{
			PadLevelText:              true,
			ForceColors:               false,
			EnvironmentOverrideColors: true,
			DisableTimestamp:          true,
		})
	}

	logger.SetOutput(os.Stderr)
	return logger
}
