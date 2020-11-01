package tools

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/polynetwork/fabric-relayer/internal/github.com/hyperledger/fabric/protoutil"
	"github.com/polynetwork/poly/common"
	"strings"
)

type FabricSdk struct {
	sdk *fabsdk.FabricSDK
	channelClient *channel.Client
	eventClient *event.Client
	ledgerClient *ledger.Client
}

type CrossChainEvent struct {
	Data  []byte
	TxHash []byte
}

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

func newLedger(sdk *fabsdk.FabricSDK) *ledger.Client {
	ccp := sdk.ChannelContext("mychannel", fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1"))
	ledgerClient, err := ledger.New(ccp)
	if err != nil {
		panic(err)
	}
	return ledgerClient
}

func newChannelClient(sdk *fabsdk.FabricSDK) (*channel.Client, *event.Client, *ledger.Client) {
	ccp := sdk.ChannelContext("mychannel", fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1"))
	cc, err := channel.New(ccp)
	if err != nil {
		panic(err)
	}
	eventClient, err := event.New(ccp, event.WithBlockEvents())
	if err != nil {
		panic(err)
	}
	ledgerClient, err := ledger.New(ccp)
	if err != nil {
		panic(err)
	}
	return cc, eventClient,ledgerClient
}

func newEventClient(sdk *fabsdk.FabricSDK) *event.Client {
	ccp := sdk.ChannelContext("mychannel", fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1"))
	eventClient, err := event.New(ccp, event.WithBlockEvents())
	if err != nil {
		panic(err)
	}
	return eventClient
}

func NewFabricSdk() (*FabricSdk, error) {
	fabricSdk := &FabricSdk{}
	fabricSdk.sdk = newFabSdk()
	fabricSdk.channelClient, fabricSdk.eventClient, fabricSdk.ledgerClient = newChannelClient(fabricSdk.sdk)
	return fabricSdk, nil
}

func (sdk *FabricSdk) RegisterCrossChainEvent() (fab.Registration, <-chan *fab.CCEvent, error) {
	reg, notifier, err := sdk.eventClient.RegisterChaincodeEvent("ccm1", "to_ploy.*")
	if err != nil {
		return nil, nil, err
	} else {
		return reg, notifier, nil
	}
}

func (sdk *FabricSdk) Unregister(reg fab.Registration) {
	sdk.eventClient.Unregister(reg)
}

func (sdk *FabricSdk) GetLatestHeight() (uint64, error) {
	info, err := sdk.ledgerClient.QueryInfo()
	if err != nil {
		return 0, err
	}
	return info.BCI.Height, nil
}

func (sdk *FabricSdk) GetCrossChainEvent(height uint64) ([]*CrossChainEvent, error) {
	block, err := sdk.ledgerClient.QueryBlock(height)
	if err != nil {
		return nil, err
	}
	events := make([]*CrossChainEvent, 0)
	for _, v := range block.Data.Data {
		xx , err := protoutil.GetEnvelopeFromBlock(v)
		if err != nil {
			return nil, err	
		}
		cas, err := protoutil.GetActionsFromEnvelopeMsg(xx)
		//cas, err := protoutil.GetActionsFromEnvelope(v)
		if err != nil {
			return nil, err
		}

		for _, e := range cas {
			chaincodeEvent := &peer.ChaincodeEvent{}
			err = proto.Unmarshal(e.Events, chaincodeEvent)
			if err != nil {
				return nil, err
			}
			if strings.Contains(chaincodeEvent.EventName , "from_ccm") {
				txHash, _ := hex.DecodeString(chaincodeEvent.TxId)
				events = append(events, &CrossChainEvent{
					Data: chaincodeEvent.Payload,
					TxHash: txHash,
				})
			}
		}
	}
	return events, nil
}

func (sdk *FabricSdk) CrossChainTransfer(crossChainTxProof []byte, header []byte, headerProof []byte, anchor []byte) {
	req := channel.Request{
		ChaincodeID: "ccm1",
		Fcn: "verifyHeaderAndExecuteTx",
		Args: packArgs([]string{hex.EncodeToString(crossChainTxProof), hex.EncodeToString(header), hex.EncodeToString(headerProof), hex.EncodeToString(anchor)}),
	}
	fmt.Printf("proof: %s\nheader: %s\n", hex.EncodeToString(crossChainTxProof), hex.EncodeToString(header))
	response, err := sdk.channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response: %v\n", string(response.TransactionID))
}

func (sdk *FabricSdk) PolyHeader(header []byte) {
	req := channel.Request{
		ChaincodeID: "ccm1",
		Fcn: "changeBookKeeper",
		Args: [][]byte{header},
	}
	response, err := sdk.channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response: %v\n", string(response.TransactionID))
}

func (sdk *FabricSdk) GetLatestSyncHeight() uint32 {
	req := channel.Request{
		ChaincodeID: "ccm1",
		Fcn: "getPolyEpochHeight",
		Args: [][]byte{},
	}
	response, err := sdk.channelClient.Query(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	height := binary.LittleEndian.Uint32(response.Payload)
	return height
}

func (sdk *FabricSdk) Lock() {
	req := channel.Request{
		ChaincodeID: "peth",
		Fcn: "lock",
		Args: packArgs([]string{"2", "BC8F34783742ea552C7e8823a2A9e8f58052B4D4", "10"}),
	}
	response, err := sdk.channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response: %v\n", string(response.TransactionID))
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
