// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
)

type Base struct {
	// Timestamp is the expiry of the transaction (inclusive). Once this time passes and the
	// transaction is not included in a block, it is safe to regenerate it.
	Timestamp int64 `json:"nonce"`

	// ChainID protects against replay attacks on different VM instances.
	ChainID ids.ID `json:"chainId"`

	// Unit price is the value per unit to spend on this transaction.
	UnitPrice uint64 `json:"unitPrice"`
}

func (b *Base) Execute(chainID ids.ID, r Rules, timestamp int64) error {
	switch {
	case b.Timestamp%consts.MillisecondsPerSecond != 0:
		// TODO: make this modulus configurable
		return ErrMisalignedTime
	case b.Timestamp < timestamp: // tx: 100 block: 110
		return ErrTimestampTooLate
	case b.Timestamp > timestamp+r.GetValidityWindow(): // tx: 100 block 10
		return ErrTimestampTooEarly
	case b.ChainID != chainID:
		return ErrInvalidChainID
	case b.UnitPrice < r.GetMinUnitPrice():
		return ErrInvalidUnitPrice
	default:
		return nil
	}
}

func (*Base) Size() int {
	return consts.Uint64Len*2 + consts.IDLen
}

func (b *Base) Marshal(p *codec.Packer) {
	p.PackInt64(b.Timestamp)
	p.PackID(b.ChainID)
	p.PackUint64(b.UnitPrice)
}

func UnmarshalBase(p *codec.Packer) (*Base, error) {
	var base Base
	base.Timestamp = p.UnpackInt64(true)
	if base.Timestamp%consts.MillisecondsPerSecond != 0 {
		// TODO: make this modulus configurable
		return nil, ErrMisalignedTime
	}
	p.UnpackID(true, &base.ChainID)
	base.UnitPrice = p.UnpackUint64(true)
	return &base, p.Err()
}
