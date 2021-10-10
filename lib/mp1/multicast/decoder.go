package multicast

func BMsgDecoder(data []byte) (interface{}, error) {
	msg := &BMsg{}
	return msg.Decode(data)
}

func RMsgDecoder(data []byte) (interface{}, error) {
	msg := &RMsg{}
	return msg.Decode(data)
}
