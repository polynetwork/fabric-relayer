package test

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/polynetwork/fabric-relayer/internal/github.com/hyperledger/fabric/protoutil"
	common2 "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	poly_common "github.com/polynetwork/poly/common"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCrossChainEvent(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		panic("startServer - get current work directory failed!")
		return
	}
	os.Setenv("FABRIC_RELAYER_PATH", dir)

	sdk := newFabSdk()
	channelClient := newChannelClient(sdk, "mychannel")
	eventClient := newEventClient(sdk, "mychannel")

	eventID := "to.*"
	reg, notifier, err := eventClient.RegisterChaincodeEvent("ccm1", eventID)
	if err != nil {
		panic(err)
	}
	defer eventClient.Unregister(reg)

	req := channel.Request{
		ChaincodeID: "peth",
		Fcn: "lock",
		Args: packArgs([]string{"2", "BC8F34783742ea552C7e8823a2A9e8f58052B4D4", "11"}),
	}
	response, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response: %s\n", string(response.TransactionID))

	select {
	case ccEvent := <- notifier:
		fmt.Printf("receive cc event:%v\n", ccEvent)
	case <- time.After(time.Second * 600):
		fmt.Printf("not receive cc event!")
	}
}

func TestCrossChainSyncHeight(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		panic("startServer - get current work directory failed!")
		return
	}
	os.Setenv("FABRIC_RELAYER_PATH", dir)

	sdk := newFabSdk()
	channelClient := newChannelClient(sdk, "mychannel")
	req := channel.Request{
		ChaincodeID: "ccm1",
		Fcn: "getPolyEpochHeight",
		Args: [][]byte{},
	}
	response, err := channelClient.Query(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	height := binary.LittleEndian.Uint32(response.Payload)
	fmt.Printf("txid: %s, height: %d\n", response.TransactionID, height)
}


func TestBlockEvent(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		panic("startServer - get current work directory failed!")
		return
	}
	os.Setenv("FABRIC_RELAYER_PATH", dir)

	sdk := newFabSdk()
	ledgerClient := newLedgerClient(sdk, "mychannel")
	for i := uint64(3); i < 50; i++ {
		block, err := ledgerClient.QueryBlock(i)
		if err != nil {
			panic(err)
		}
		for _, v := range block.Data.Data {
			xx, err := protoutil.GetEnvelopeFromBlock(v)
			if err != nil {
				t.Fatal(err)
			}
			//cas, err := protoutil.GetActionsFromEnvelope(v)
			cas, err := protoutil.GetActionsFromEnvelopeMsg(xx)
			if err != nil {
				t.Fatal(err)
			}

			for _, e := range cas {
				chaincodeEvent := &peer.ChaincodeEvent{}
				err = proto.Unmarshal(e.Events, chaincodeEvent)
				if err != nil {
					t.Fatal(err)
				}
				fmt.Printf("event, height: %d, chaincode: %s, txid: %s, name: %s\n", i, chaincodeEvent.ChaincodeId, chaincodeEvent.TxId, chaincodeEvent.EventName)
				if strings.Contains(chaincodeEvent.EventName, "from_ccm") {
					source := poly_common.NewZeroCopySource(chaincodeEvent.Payload)
					param := &common2.MakeTxParam{}
					param.Deserialization(source)
					fmt.Printf("param: %s\n", hex.EncodeToString(param.CrossChainID))
				}
			}
		}
	}
}