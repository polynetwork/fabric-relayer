package manager

import (
	"fmt"

	"github.com/polynetwork/poly/common"
)

type CrossTransfer struct {
	txIndex string
	txId    []byte
	value   []byte
	toChain uint32
	height  uint64
}

func (c *CrossTransfer) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(c.txIndex)
	sink.WriteVarBytes(c.txId)
	sink.WriteVarBytes(c.value)
	sink.WriteUint32(c.toChain)
	sink.WriteUint64(c.height)
}

func (c *CrossTransfer) Deserialization(source *common.ZeroCopySource) error {
	txIndex, eof := source.NextString()
	if eof {
		return fmt.Errorf("Waiting deserialize txIndex error")
	}
	txId, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Waiting deserialize txId error")
	}
	value, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Waiting deserialize value error")
	}
	toChain, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("Waiting deserialize toChain error")
	}
	height, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("Waiting deserialize height error")
	}
	c.txIndex = txIndex
	c.txId = txId
	c.value = value
	c.toChain = toChain
	c.height = height
	return nil
}
