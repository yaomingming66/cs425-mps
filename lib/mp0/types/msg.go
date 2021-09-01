package types

type Msg struct {
	From      string `json:"from"`
	TimeStamp string `json:"timestamp"`
	Payload   string `json:"payload"`
}
