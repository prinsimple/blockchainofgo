package core

import "github.com/prinsimple/goblock/types"

type Header struct {
	Version   uint32
	PrevHash  types.Hash
	TimeStamp uint64
	Height    uint32
	Nonce     uint64
}

type Block struct {
	Header
	Transaction []Transaction
}

func (h *Header) EncodeBinary() {

}

func (h *Header) DecodeBinary() {

}
