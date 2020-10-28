package tools

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/polynetwork/poly/common"
)

type FabricSdk struct {
	sdk *fabsdk.FabricSDK
	channelClient *channel.Client
	eventClient *event.Client
}

func newFabSdk() *fabsdk.FabricSDK {
	sdk, err := fabsdk.New(config.FromFile("./config_e2e.yaml"))
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
	ccp := sdk.ChannelContext("mychannel", fabsdk.WithUser("User1"), fabsdk.WithOrg("Org1"))
	cc, err := channel.New(ccp)
	if err != nil {
		panic(err)
	}
	return cc
}

func newEventClient(sdk *fabsdk.FabricSDK) *event.Client {
	ccp := sdk.ChannelContext("mychannel", fabsdk.WithUser("User1"), fabsdk.WithOrg("Org1"))
	eventClient, err := event.New(ccp)
	if err != nil {
		panic(err)
	}
	return eventClient
}

func NewFabricSdk() (*FabricSdk, error) {
	fabricSdk := &FabricSdk{}
	fabricSdk.sdk = newFabSdk()
	fabricSdk.channelClient = newChannelClient(fabricSdk.sdk)
	fabricSdk.eventClient = newEventClient(fabricSdk.sdk)
	return fabricSdk, nil
}

func (sdk *FabricSdk) RegisterCrossChainEvent() (fab.Registration, <-chan *fab.CCEvent, error) {
	reg, notifier, err := sdk.eventClient.RegisterChaincodeEvent("basic", "eventID")
	if err != nil {
		return nil, nil, err
	} else {
		return reg, notifier, nil
	}
}

func (sdk *FabricSdk) Unregister(reg fab.Registration) {
	sdk.eventClient.Unregister(reg)
}

func (sdk *FabricSdk) CrossChainTransfer() {
	req := channel.Request{
		ChaincodeID: "basic",
		Fcn: "TransferAsset",
		Args: packArgs([]string{"asset6","Christopher"}),
	}
	response, err := sdk.channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response: %v\n", string(response.TransactionID))
}

func (sdk *FabricSdk) PolyHeader() {

}


func packArgs(args []string) [][]byte {
	ret := make([][]byte, 0)
	for _, arg := range args {
		ret = append(ret, []byte(arg))
	}
	return ret
}

func ParseAuditpath(path []byte) ([]byte, []byte, [][32]byte, error) {
	source := common.NewZeroCopySource(path)
	/*
		l, eof := source.NextUint64()
		if eof {
			return nil, nil, nil, nil
		}
	*/
	value, eof := source.NextVarBytes()
	if eof {
		return nil, nil, nil, nil
	}
	size := int((source.Size() - source.Pos()) / common.UINT256_SIZE)
	pos := make([]byte, 0)
	hashs := make([][32]byte, 0)
	for i := 0; i < size; i++ {
		f, eof := source.NextByte()
		if eof {
			return nil, nil, nil, nil
		}
		pos = append(pos, f)

		v, eof := source.NextHash()
		if eof {
			return nil, nil, nil, nil
		}
		var onehash [32]byte
		copy(onehash[:], (v.ToArray())[0:32])
		hashs = append(hashs, onehash)
	}

	return value, pos, hashs, nil
}
