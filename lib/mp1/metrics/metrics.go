package metrics

type Bandwith struct {
	nodeID    string
	timeStamp string
	size      int
}

type Delay struct {
	messageID string
	timestamp string
}
