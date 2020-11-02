package test

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"os"
)

func newFabSdk() *fabsdk.FabricSDK {
	dir, err := os.Getwd()
	if err != nil {
		panic("get current work directory failed!")
	}
	os.Setenv("FABRIC_RELAYER_PATH", dir + "/..")

	sdk, err := fabsdk.New(config.FromFile("./../config/config_e2e.yaml"))
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

func newChannelClient(sdk *fabsdk.FabricSDK, channelID string) *channel.Client {
	ccp := sdk.ChannelContext(channelID, fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1"))
	cc, err := channel.New(ccp)
	if err != nil {
		panic(err)
	}
	return cc
}

func newEventClient(sdk *fabsdk.FabricSDK, channelID string) *event.Client {
	ccp := sdk.ChannelContext(channelID, fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1"))
	eventClient, err := event.New(ccp, event.WithBlockEvents())
	if err != nil {
		panic(err)
	}
	return eventClient
}

func newLedgerClient(sdk *fabsdk.FabricSDK, channelID string) *ledger.Client {
	ccp := sdk.ChannelContext(channelID, fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1"))
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
