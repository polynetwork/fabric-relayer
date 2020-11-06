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
	poly_common "github.com/polynetwork/poly/common"
	common2 "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"strings"
	"testing"
	"time"
)

func TestCrossChainEvent(t *testing.T) {
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
		Fcn:         "lock",
		Args:        packArgs([]string{"2", "BC8F34783742ea552C7e8823a2A9e8f58052B4D4", "11"}),
	}
	response, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response: %s\n", string(response.TransactionID))

	select {
	case ccEvent := <-notifier:
		fmt.Printf("receive cc event:%v\n", ccEvent)
	case <-time.After(time.Second * 600):
		fmt.Printf("not receive cc event!")
	}
}

func TestCrossChainSyncHeight(t *testing.T) {
	sdk := newFabSdk()
	channelClient := newChannelClient(sdk, "mychannel")
	ledgerClient := newLedgerClient(sdk, "mychannel")
	req := channel.Request{
		ChaincodeID: "ccm",
		Fcn:         "getPolyEpochHeight",
		Args:        [][]byte{},
	}
	response, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	height := binary.LittleEndian.Uint32(response.Payload)
	fmt.Printf("txid: %s, height: %d\n", response.TransactionID, height)

	for true {
		tx, err := ledgerClient.QueryTransaction(response.TransactionID)
		if err != nil {
			panic(err)
		}
		fmt.Printf("code: %d\n", tx.ValidationCode)
		time.Sleep(time.Microsecond * 100)
	}
}

func TestLock(t *testing.T) {
	sdk := newFabSdk()
	channelClient := newChannelClient(sdk, "mychannel")
	ledgerClient := newLedgerClient(sdk, "mychannel")

	req := channel.Request{
		ChaincodeID: "peth",
		Fcn:         "lock",
		Args:        packArgs([]string{"2", "BC8F34783742ea552C7e8823a2A9e8f58052B4D4", "10"}),
	}

	response, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response: %v\n", string(response.TransactionID))

	tx, err := ledgerClient.QueryTransaction(response.TransactionID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("code: %d\n", tx.ValidationCode)
}

func TestFabricTxStatus(t *testing.T) {
	sdk := newFabSdk()
	channelClient := newChannelClient(sdk, "mychannel")
	ledgerClient := newLedgerClient(sdk, "mychannel")

	req := channel.Request{
		ChaincodeID: "lockproxy",
		Fcn:         "lock",
		Args:        packArgs([]string{"peth", "2", "BC8F34783742ea552C7e8823a2A9e8f58052B4D4", "1"}),
	}
	//
	{
		info, err := ledgerClient.QueryInfo()
		if err != nil {
			panic(err)
		}
		fmt.Printf("height: %d\n", info.BCI.Height)
	}
	response1, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response 1: %v\n", string(response1.TransactionID))

	{
		info, err := ledgerClient.QueryInfo()
		if err != nil {
			panic(err)
		}
		fmt.Printf("height: %d\n", info.BCI.Height)
	}
	response2, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response 2: %v\n", string(response2.TransactionID))

	{
		info, err := ledgerClient.QueryInfo()
		if err != nil {
			panic(err)
		}
		fmt.Printf("height: %d\n", info.BCI.Height)
	}
	response3, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response 3: %v\n", string(response3.TransactionID))

	{
		info, err := ledgerClient.QueryInfo()
		if err != nil {
			panic(err)
		}
		fmt.Printf("height: %d\n", info.BCI.Height)
	}
	response4, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response 4: %v\n", string(response4.TransactionID))

	{
		info, err := ledgerClient.QueryInfo()
		if err != nil {
			panic(err)
		}
		fmt.Printf("height: %d\n", info.BCI.Height)
	}
	//
	for true {
		tx1, err := ledgerClient.QueryTransaction(response1.TransactionID)
		if err != nil {
			panic(err)
		}
		fmt.Printf("transaction 1 code: %d\n", tx1.ValidationCode)

		tx2, err := ledgerClient.QueryTransaction(response2.TransactionID)
		if err != nil {
			panic(err)
		}
		fmt.Printf("transaction 2 code: %d\n", tx2.ValidationCode)

		tx3, err := ledgerClient.QueryTransaction(response3.TransactionID)
		if err != nil {
			panic(err)
		}
		fmt.Printf("transaction 3 code: %d\n", tx3.ValidationCode)

		tx4, err := ledgerClient.QueryTransaction(response4.TransactionID)
		if err != nil {
			panic(err)
		}
		fmt.Printf("transaction 4 code: %d\n", tx4.ValidationCode)
		time.Sleep(time.Microsecond * 500)
	}
}

func TestFabricTxStatus_1(t *testing.T) {
	sdk := newFabSdk()
	channelClient := newChannelClient(sdk, "mychannel")
	ledgerClient := newLedgerClient(sdk, "mychannel")

	req := channel.Request{
		ChaincodeID: "lockproxy",
		Fcn:         "lock",
		Args:        packArgs([]string{"peth", "2", "BC8F34783742ea552C7e8823a2A9e8f58052B4D4", "1"}),
	}
	var err error
	var response1 channel.Response
	go func() {
		//
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("before tx1, height: %d\n", info.BCI.Height)
		}
		response1, err = channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
		if err != nil {
			panic(err)
		}
		fmt.Printf("response of  tx1: %v\n", string(response1.TransactionID))
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("after tx1, height: %d\n", info.BCI.Height)
		}
	}()


	var response2 channel.Response
	go func() {
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("before tx2, height: %d\n", info.BCI.Height)
		}
		response2, err = channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
		if err != nil {
			panic(err)
		}
		fmt.Printf("response of tx2: %v\n", string(response2.TransactionID))
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("after tx2, height: %d\n", info.BCI.Height)
		}
	}()


	var response3 channel.Response
	go func() {
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("before tx3, height: %d\n", info.BCI.Height)
		}
		response3, err = channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
		if err != nil {
			panic(err)
		}
		fmt.Printf("response of tx3: %v\n", string(response3.TransactionID))
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("after tx3, height: %d\n", info.BCI.Height)
		}
	}()

	var response4 channel.Response
	go func() {
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("before tx4, height: %d\n", info.BCI.Height)
		}
		response4, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
		if err != nil {
			panic(err)
		}
		fmt.Printf("response of tx4: %v\n", string(response4.TransactionID))
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("after tx4, height: %d\n", info.BCI.Height)
		}
	}()

	for true {
		if response1.TransactionID != "" {
			tx1, err := ledgerClient.QueryTransaction(response1.TransactionID)
			if err != nil {
				panic(err)
			}
			fmt.Printf("transaction 1 code: %d\n", tx1.ValidationCode)
		}

		if response2.TransactionID != "" {
			tx2, err := ledgerClient.QueryTransaction(response2.TransactionID)
			if err != nil {
				panic(err)
			}
			fmt.Printf("transaction 2 code: %d\n", tx2.ValidationCode)
		}

		if response3.TransactionID != "" {
			tx3, err := ledgerClient.QueryTransaction(response3.TransactionID)
			if err != nil {
				panic(err)
			}
			fmt.Printf("transaction 3 code: %d\n", tx3.ValidationCode)
		}

		if response4.TransactionID != "" {
			tx4, err := ledgerClient.QueryTransaction(response4.TransactionID)
			if err != nil {
				panic(err)
			}
			fmt.Printf("transaction 4 code: %d\n", tx4.ValidationCode)
		}
		time.Sleep(time.Microsecond * 500)
	}
}


func TestFabricTxStatus_2(t *testing.T) {
	req := channel.Request{
		ChaincodeID: "lockproxy",
		Fcn:         "lock",
		Args:        packArgs([]string{"peth", "2", "BC8F34783742ea552C7e8823a2A9e8f58052B4D4", "1"}),
	}
	var err error
	var response1 channel.Response
	go func() {
		sdk := newFabSdk()
		channelClient := newChannelClient(sdk, "mychannel")
		ledgerClient := newLedgerClient(sdk, "mychannel")
		//
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("before tx1, height: %d\n", info.BCI.Height)
		}
		response1, err = channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
		if err != nil {
			fmt.Printf("error: %v", err)
		}
		fmt.Printf("response of  tx1: %v\n", string(response1.TransactionID))
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("after tx1, height: %d\n", info.BCI.Height)
		}
	}()


	var response2 channel.Response
	go func() {
		sdk := newFabSdk()
		channelClient := newChannelClient(sdk, "mychannel")
		ledgerClient := newLedgerClient(sdk, "mychannel")
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("before tx2, height: %d\n", info.BCI.Height)
		}
		response2, err = channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
		if err != nil {
			fmt.Printf("error: %v", err)
		}
		fmt.Printf("response of tx2: %v\n", string(response2.TransactionID))
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("after tx2, height: %d\n", info.BCI.Height)
		}
	}()


	var response3 channel.Response
	go func() {
		sdk := newFabSdk()
		channelClient := newChannelClient(sdk, "mychannel")
		ledgerClient := newLedgerClient(sdk, "mychannel")
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("before tx3, height: %d\n", info.BCI.Height)
		}
		response3, err = channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
		if err != nil {
			fmt.Printf("error: %v", err)
		}
		fmt.Printf("response of tx3: %v\n", string(response3.TransactionID))
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("after tx3, height: %d\n", info.BCI.Height)
		}
	}()

	var response4 channel.Response
	go func() {
		sdk := newFabSdk()
		channelClient := newChannelClient(sdk, "mychannel")
		ledgerClient := newLedgerClient(sdk, "mychannel")
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("before tx4, height: %d\n", info.BCI.Height)
		}
		response4, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
		if err != nil {
			fmt.Printf("error: %v", err)
		}
		fmt.Printf("response of tx4: %v\n", string(response4.TransactionID))
		{
			info, err := ledgerClient.QueryInfo()
			if err != nil {
				panic(err)
			}
			fmt.Printf("after tx4, height: %d\n", info.BCI.Height)
		}
	}()

	sdk := newFabSdk()
	ledgerClient := newLedgerClient(sdk, "mychannel")
	info, err := ledgerClient.QueryInfo()
	height := info.BCI.Height
	current := height - 1
	for true {
		if response1.TransactionID != "" {
			tx1, err := ledgerClient.QueryTransaction(response1.TransactionID)
			if err != nil {
				panic(err)
			}
			fmt.Printf("transaction 1 code: %d\n", tx1.ValidationCode)
		}

		if response2.TransactionID != "" {
			tx2, err := ledgerClient.QueryTransaction(response2.TransactionID)
			if err != nil {
				panic(err)
			}
			fmt.Printf("transaction 2 code: %d\n", tx2.ValidationCode)
		}

		if response3.TransactionID != "" {
			tx3, err := ledgerClient.QueryTransaction(response3.TransactionID)
			if err != nil {
				panic(err)
			}
			fmt.Printf("transaction 3 code: %d\n", tx3.ValidationCode)
		}

		if response4.TransactionID != "" {
			tx4, err := ledgerClient.QueryTransaction(response4.TransactionID)
			if err != nil {
				panic(err)
			}
			fmt.Printf("transaction 4 code: %d\n", tx4.ValidationCode)
		}

		info, err := ledgerClient.QueryInfo()
		if err != nil {
			panic(err)
		}
		height := info.BCI.Height - 1
		if height != current {
			block, err := ledgerClient.QueryBlock(height)
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
					fmt.Println(chaincodeEvent.String())
				}
			}
			current = height
		}
		time.Sleep(time.Microsecond * 1000)
	}
}


func TestBlockEvent(t *testing.T) {
	sdk := newFabSdk()
	ledgerClient := newLedgerClient(sdk, "mychannel")
	for i := uint64(73); i < 74; i++ {
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
