package multicast

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
)

type BMsg struct {
	SrcID string `json:"src"`
	Path  string `json:"path"`
	Body  []byte `json:"body"`
}

// SHA1 hashes using sha1 algorithm
func SHA1(text string) string {
	algorithm := sha1.New()
	algorithm.Write([]byte(text))
	return hex.EncodeToString(algorithm.Sum(nil))
}

func NewBMsg(srcID string, path string, v interface{}) (msg *BMsg, err error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return &BMsg{
		SrcID: srcID,
		Path:  path,
		Body:  data,
	}, nil
}

func (m *BMsg) Encode() (data []byte, err error) {
	data, err = json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *BMsg) Decode(data []byte) (msg *BMsg, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

type RMsg struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Body []byte `json:"body"`
}

func NewRMsg(path string, v interface{}) (*RMsg, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String() + "-" + SHA1(string(data))

	return &RMsg{
		ID:   id,
		Path: path,
		Body: data,
	}, nil
}

func (m *RMsg) Encode() (data []byte, err error) {
	data, err = json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *RMsg) Decode(data []byte) (msg *RMsg, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

type TOAskProposalSeqMsg struct {
	SrcID string `json:"src"`
	MsgID string `json:"msg_id"`
	Body  []byte `json:"body"`
}

func NewTOAskProposalSeqMsg(srcID string, body []byte) *TOAskProposalSeqMsg {
	msgID := uuid.New().String() + SHA1(string(body))
	return &TOAskProposalSeqMsg{
		SrcID: srcID,
		MsgID: msgID,
		Body:  body,
	}
}

func (m *TOAskProposalSeqMsg) Encode() (data []byte, err error) {
	data, err = json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *TOAskProposalSeqMsg) Decode(data []byte) (msg *TOAskProposalSeqMsg, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

type TOReplyProposalSeqMsg struct {
	ProcessID   string `json:"pid"`
	MsgID       string `json:"msg_id"`
	ProposalSeq uint64 `json:"proposal_seq"`
}

func NewTOReplyProposalSeqMsg(processID string, msgID string, proposalSeqNum uint64) *TOReplyProposalSeqMsg {
	return &TOReplyProposalSeqMsg{
		ProcessID:   processID,
		MsgID:       msgID,
		ProposalSeq: proposalSeqNum,
	}
}

func (m *TOReplyProposalSeqMsg) Encode() (data []byte, err error) {
	data, err = json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *TOReplyProposalSeqMsg) Decode(data []byte) (msg *TOReplyProposalSeqMsg, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

type TOAnnounceAgreementSeqMsg struct {
	ProcessID    string `json:"pid"`
	AgreementSeq uint64 `json:"agreement_seq"`
	MsgID        string `json:"msg_id"`
}

func NewTOAnnounceAgreementSeqMsg(processID string, agreementSeq uint64, msgID string) *TOAnnounceAgreementSeqMsg {
	return &TOAnnounceAgreementSeqMsg{
		ProcessID:    processID,
		AgreementSeq: agreementSeq,
		MsgID:        msgID,
	}
}

func (m *TOAnnounceAgreementSeqMsg) Encode() (data []byte, err error) {
	data, err = json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *TOAnnounceAgreementSeqMsg) Decode(data []byte) (msg *TOAnnounceAgreementSeqMsg, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

type TOMsg struct {
	Path string `json:"path"`
	Body []byte `json:"body"`
}

func NewTOMsg(path string, v interface{}) (msg *TOMsg, err error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return &TOMsg{
		Path: path,
		Body: data,
	}, nil
}

func (m *TOMsg) Encode() (data []byte, err error) {
	data, err = json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *TOMsg) Decode(data []byte) (msg *TOMsg, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
