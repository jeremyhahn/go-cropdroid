//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import "encoding/json"

type RaftProposal interface {
	Serialize() ([]byte, error)
}

type Proposal struct {
	Query int
	Data  []byte
	RaftProposal
}

func NewProposal() RaftProposal {
	return &Proposal{}
}

func CreateProposal(query int, data []byte) RaftProposal {
	return &Proposal{Query: query, Data: data}
}

func (proposal *Proposal) Serialize() ([]byte, error) {
	return json.Marshal(proposal)
}
