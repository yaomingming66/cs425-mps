package types

type Msg struct {
	From      string `json:"from"`
	TimeStamp string `json:"timestamp"`
	Payload   string `json:"payload"`
}

type Params struct {
	Delay     string `json:"delay from receiving"`
	TimeStamp string `json:"timestamp"`
	Size      string `json:"size of payload"`
}
