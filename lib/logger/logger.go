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
	logger.SetFormatter(&log.TextFormatter{
		PadLevelText:              true,
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		DisableTimestamp:          true,
	})
	logger.SetOutput(os.Stderr)
	return logger
}
