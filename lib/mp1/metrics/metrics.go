package metrics

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	logger          = log.WithField("src", "metrics")
	bandwidthLogger = log.WithField("src", "metrics.bandwidth")
	delayLogger     = log.WithField("src", "metrics.delay")
	enableLog       = true
)

func SetupMetrics() {
	metricsENV := os.Getenv("METRICS")
	if metricsENV == "y" {
		enableLog = true
	} else {
		enableLog = false
	}
}

type BandwidthLogEntry struct {
	NodeID    string `json:"node_id"`
	Timestamp string `json:"timestamp"`
	BytesSize int    `json:"bytes_size"`
}

type DelayLogEntry struct {
	NodeID    string `json:"node_id"`
	MsgID     string `json:"msg_id"`
	Timestamp string `json:"timestamp"`
}

func NowUnixNana() string {
	timeNow := time.Now().UnixNano()
	return strconv.FormatInt(timeNow, 10)
}

func NewBandwidthLogEntry(nodeID string, bytesSize int) *BandwidthLogEntry {
	return &BandwidthLogEntry{
		NodeID:    nodeID,
		BytesSize: bytesSize,
		Timestamp: NowUnixNana(),
	}
}

func (b *BandwidthLogEntry) Encode() (data []byte, err error) {
	data, err = json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (b *BandwidthLogEntry) Log() {
	encoded, err := b.Encode()
	if err != nil {
		logger.Errorf("encode bandwidth log entry failed: %v", err)
		return
	}

	if enableLog {
		bandwidthLogger.Infof(string(encoded))
	}
}

func NewDelayLogEntry(nodeID string, msgID string) *DelayLogEntry {
	return &DelayLogEntry{
		NodeID:    nodeID,
		MsgID:     msgID,
		Timestamp: NowUnixNana(),
	}
}

func (d *DelayLogEntry) Encode() (data []byte, err error) {
	data, err = json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *DelayLogEntry) Log() {
	encoded, err := d.Encode()
	if err != nil {
		logger.Errorf("encode delay log entry failed: %v", err)
		return
	}

	if enableLog {
		delayLogger.Infof(string(encoded))
	}
}
