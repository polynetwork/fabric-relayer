package fabric_relayer

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"os"
	"testing"
	"time"
)

func newFabSdk() *fabsdk.FabricSDK {
	sdk, err := fabsdk.New(config.FromFile("./config/config_e2e.yaml"))
	if err != nil {
		panic(err)
	}
	return sdk
}

func newResMgmt(sdk *fabsdk.FabricSDK) *resmgmt.Client {
	rcp := sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1"))
	rc, err := resmgmt.New(rcp)
	if err != nil {
		panic(err)
	}
	return rc
}

func newChannelClient(sdk *fabsdk.FabricSDK) *channel.Client {
	ccp := sdk.ChannelContext("mychannel", fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1"))
	cc, err := channel.New(ccp)
	if err != nil {
		panic(err)
	}
	return cc
}

func newEventClient(sdk *fabsdk.FabricSDK) *event.Client {
	ccp := sdk.ChannelContext("mychannel", fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1"))
	eventClient, err := event.New(ccp, event.WithBlockEvents())
	if err != nil {
		panic(err)
	}
	return eventClient
}

func newLedger(sdk *fabsdk.FabricSDK) *ledger.Client {
	ccp := sdk.ChannelContext("mychannel", fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1"))
	ledgerClient, err := ledger.New(ccp)
	if err != nil {
		panic(err)
	}
	return ledgerClient
}

func packArgs(args []string) [][]byte {
	ret := make([][]byte, 0)
	for _, arg := range args {
		ret = append(ret, []byte(arg))
	}
	return ret
}

func TestCCQuery(t *testing.T) {
	sdk := newFabSdk()
	channelClient := newChannelClient(sdk)
	req := channel.Request{
		ChaincodeID: "basic",
		Fcn: "GetAllAssets",
		Args: packArgs([]string{}),
	}
	response, err := channelClient.Query(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response: %s\n", string(response.Payload))
}

func TestCCInvoke(t *testing.T) {
	sdk := newFabSdk()
	channelClient := newChannelClient(sdk)
	req := channel.Request{
		ChaincodeID: "basic",
		Fcn: "TransferAsset",
		Args: packArgs([]string{"asset6","Christopher"}),
	}
	response, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response: %v\n", string(response.TransactionID))
}

func TestCCEvent(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		panic("startServer - get current work directory failed!")
		return
	}
	os.Setenv("FABRIC_RELAYER_PATH", dir)

	sdk := newFabSdk()
	channelClient := newChannelClient(sdk)
	eventClient := newEventClient(sdk)

	eventID := "test"
	reg, notifier, err := eventClient.RegisterChaincodeEvent("mycc", eventID)
	if err != nil {
		panic(err)
	}
	defer eventClient.Unregister(reg)

	req := channel.Request{
		ChaincodeID: "mycc",
		Fcn: "query",
		Args: packArgs([]string{"a"}),
	}
	response, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response: %s\n", string(response.TransactionID))

	select {
	case ccEvent := <- notifier:
		fmt.Printf("receive cc event:%v\n", ccEvent)
	case <- time.After(time.Second * 60):
		fmt.Printf("not receive cc event!")
	}
}

func TestCCEvent1(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		panic("startServer - get current work directory failed!")
		return
	}
	os.Setenv("FABRIC_RELAYER_PATH", dir)

	sdk := newFabSdk()
	channelClient := newChannelClient(sdk)
	eventClient := newEventClient(sdk)

	eventID := "test"
	reg, notifier, err := eventClient.RegisterChaincodeEvent("peth", eventID)
	if err != nil {
		panic(err)
	}
	defer eventClient.Unregister(reg)

	req := channel.Request{
		ChaincodeID: "peth",
		Fcn: "lock",
		Args: packArgs([]string{"2", "BC8F34783742ea552C7e8823a2A9e8f58052B4D4", "10"}),
	}
	response, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response: %s\n", string(response.TransactionID))

	select {
	case ccEvent := <- notifier:
		fmt.Printf("receive cc event:%v\n", ccEvent)
	case <- time.After(time.Second * 60):
		fmt.Printf("not receive cc event!")
	}
}

func TestTransaction(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		panic("startServer - get current work directory failed!")
		return
	}
	os.Setenv("FABRIC_RELAYER_PATH", dir)

	sdk := newFabSdk()
	ledgerClient := newLedger(sdk)
	tx, err := ledgerClient.QueryTransaction("5c69313e45b78a951a5ea01ad66de45ed11b198eeb3cd8f06bc968c0ff8e0cc9")
	if err != nil {
		panic(err)
	}
	fmt.Printf("transaction: %s\n", string(tx.TransactionEnvelope.Payload))
}

func TestBlock(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		panic("startServer - get current work directory failed!")
		return
	}
	os.Setenv("FABRIC_RELAYER_PATH", dir)

	sdk := newFabSdk()
	ledgerClient := newLedger(sdk)
	block, err := ledgerClient.QueryBlock(1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("block: %s\n", string(block.Data.String()))
}

