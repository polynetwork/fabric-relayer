package fabric_relayer

import (
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/polynetwork/fabric-relayer/internal/github.com/hyperledger/fabric/protoutil"
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
		Fcn:         "GetAllAssets",
		Args:        packArgs([]string{}),
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
		Fcn:         "TransferAsset",
		Args:        packArgs([]string{"asset6", "Christopher"}),
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

	eventID := ".*"
	reg, notifier, err := eventClient.RegisterChaincodeEvent("mycc", eventID)
	if err != nil {
		panic(err)
	}
	defer eventClient.Unregister(reg)

	req := channel.Request{
		ChaincodeID: "mycc",
		Fcn:         "query",
		Args:        packArgs([]string{"a"}),
	}
	response, err := channelClient.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		panic(err)
	}
	fmt.Printf("response: %s\n", string(response.TransactionID))

	select {
	case ccEvent := <-notifier:
		fmt.Printf("receive cc event:%v\n", ccEvent)
	case <-time.After(time.Second * 60):
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

	eventID := ".*"
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
	//fmt.Printf("transaction: %s\n", string(tx.TransactionEnvelope.Payload))

	pl := &common.Payload{}
	err = proto.Unmarshal(tx.TransactionEnvelope.Payload, pl)
	if err != nil {
		t.Fatal(err)
	}

	txn := &peer.Transaction{}
	err = proto.Unmarshal(pl.Data, txn)
	if err != nil {
		t.Fatal(err)
	}

	ac := &peer.TransactionAction{}
	err = proto.Unmarshal(txn.Actions[0].Payload, ac)
	if err != nil {
		t.Fatal(err)
	}

	capl := &peer.ChaincodeActionPayload{}
	err = proto.Unmarshal(ac.Payload, capl)
	if err != nil {
		t.Fatal(err)
	}

	hdr := &common.ChannelHeader{}
	err = proto.Unmarshal(pl.Header.ChannelHeader, hdr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBlock(t *testing.T) {
	//dir, err := os.Getwd()
	//if err != nil {
	//	panic("startServer - get current work directory failed!")
	//	return
	//}
	//os.Setenv("FABRIC_RELAYER_PATH", dir)
	//
	//sdk := newFabSdk()
	//ledgerClient := newLedger(sdk)
	//for i := uint64(20); i < 51; i++ {
	//	fmt.Println(i)
	//	block, err := ledgerClient.QueryBlock(i)
	//	if err != nil {
	//		panic(err)
	//	}
	//	for _, v := range block.Data.Data {
	//		cas, err := protoutil.GetActionsFromEnvelope(v)
	//		if err != nil {
	//			t.Fatal(err)
	//		}
	//
	//		for _, e := range cas {
	//			chaincodeEvent := &peer.ChaincodeEvent{}
	//			err = proto.Unmarshal(e.Events, chaincodeEvent)
	//			if err != nil {
	//				t.Fatal(err)
	//			}
	//			if chaincodeEvent.EventName == "ERC20TokenImpltransfer" {
	//				te := &TransferEvent{}
	//				err := json.Unmarshal(chaincodeEvent.Payload, te)
	//				if err != nil {
	//					t.Fatal(err)
	//				}
	//				fmt.Println("amount", big.NewInt(0).SetBytes(te.Amount).String())
	//			}
	//
	//			fmt.Println(chaincodeEvent.String())
	//		}
	//	}
	//}
	raw, _ := hex.DecodeString("0abf070a6708031a0c089bae84fd0510b3bfcdd40222096d796368616e6e656c2a40333165393038313164303232313566303833323964373331303638363235636561356564373538643664623337313131646639343264366332636163636337303a081206120463636d3112d3060ab6060a074f7267314d535012aa062d2d2d2d2d424547494e2043455254494649434154452d2d2d2d2d0a4d4949434b6a4343416443674177494241674952414c3977693432436763676f6379426a2b347a2f73744d77436759494b6f5a497a6a304541774977637a454c0a4d416b474131554542684d4356564d78457a415242674e5642416754436b4e6862476c6d62334a7561574578466a415542674e564241635444564e68626942470a636d467559326c7a593238784754415842674e5642416f54454739795a7a45755a586868625842735a53356a623230784844416142674e5642414d5445324e680a4c6d39795a7a45755a586868625842735a53356a623230774868634e4d6a41784d444d784d4449304e7a41775768634e4d7a41784d4449354d4449304e7a41770a576a42724d517377435159445651514745774a56557a45544d4245474131554543424d4b5132467361575a76636d3570595445574d4251474131554542784d4e0a5532467549455a795957356a61584e6a627a454f4d4177474131554543784d465957527461573478487a416442674e5642414d4d466b466b62576c75514739790a5a7a45755a586868625842735a53356a623230775754415442676371686b6a4f5051494242676771686b6a4f50514d4242774e4341415236774e6e7a503230370a7447423679426b5535643255344c63694a51384943334b504f776838667959724671716d485241634c6851374875753476797a793147726b7377476d444c6f620a334144684e2b576a695a4c4d6f303077537a414f42674e56485138424166384542414d434234417744415944565230544151482f424149774144417242674e560a48534d454a44416967434351366a4643535262696d6950676c704952337a68484e354452536e3869304f786564464d666b6a7036477a414b42676771686b6a4f0a5051514441674e49414442464169454136387036346c586559617a32624235557132776b3469435149787847485578654f5236557a6178645a56734349454e370a3446744a393566344e4d3963736b395233305a73526c566a747a2b564764424b79754632676a4b510a2d2d2d2d2d454e442043455254494649434154452d2d2d2d2d0a1218a4e4b47935810a9a46d6161445732c6f3b09eece57518b0a12cb0f0ac80f0ac50f1206120463636d311aba0f0a18766572696679486561646572416e644578656375746554780aa00363663230346662363361383433613232333833333836373062653134353332646138323863336233643663626438613832616433366433343236343463623235323230373032303030303030303030303030303032303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030383632306534353239373738376539393865316236326433623231343530363534333134666239663930616333623461656632333665313432616362646236326336303431343265656133343939343766393363336239623734666263663134316531303261646435313065636530373030303030303030303030303030303437303635373436383036373536653663366636333662336130343730363537343638313439613432306137653537653630373036666666363961653930313132656365633336393966313164383039363938303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030300af60b30303030303030306462303536646431303030303030303033346631666230303863633538326132306366613564373239326530353563643462373962306238616663333636363531386337646362613336653465643536303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303762653936353039383262383134386366393939333338363536653165373936643130343735646662373237636264306437333964376432633662346131333861353166643331343464323036666335353739336662356263313530333838633234316238303063643136346638363165333362623263656638613636313633666231366131356639323235303330306164643539353233393632313961343366643131303137623232366336353631363436353732323233613332326332323736373236363566373636313663373536353232336132323432343134343262343734653664366136623463363333393461346134393463333036623737346437353634333436353736353233343532366133313463373334623538373534313737363334393435363434383335363335353463353733333561373634313333373634393464353334653535333234643463363235613433366332623664373737333332373737613539363334393664333636653435366632623631373634363734356137333364323232633232373637323636356637303732366636663636323233613232333634633739356133363531326637363534346436353537346637353761373335363634363136313665373836623737366634653536366436363536373432663662373534333534373437333533353535393336353636373435326237363639353834373639343337613435356136313732326236393734366232663339363932623336363936353732373636623639373435393664346336383336363936363736373435613334346337373364336432323263323236633631373337343566363336663665363636393637356636323663366636333662356636653735366432323361333133373330333533363335326332323665363537373566363336383631363936653566363336663665363636393637323233613665373536633663376430303030303030303030303030303030303030303030303030303030303030303030303030303030303332333132303530333862386166363231306563666463626361623232353532656638643863663431633666383666396366396162353364383635373431636664623833336630366232333132303530323831373239313835343062326235313265616531383732613261326533613238643938396336306439356461623838323961646137643764643730366436353832333132303530323637393933306134326161663363363937393863613861336631326531333463303139343035383138643738336431313734386530333964653835313539383830333432303131623236623231396634323037303938383836623266303662646333323832343738646335376239366437616635343631373632333036313761643230383630646232313266343462626133356665326561316134363533343264626432343439333562313731333234633761393839343637303935393065376537303662306332343230313162306330626332366131643236646664303034646538663166396534633134383133353266346366633464653538616630343630396665626530383139366632653038316163343232333766333736353635616436646137326536653731643039333139646463383939363431643239666337623865626337393161386338373734323031316330393639326366336461623966343462393532643338623733356633313133363632373937663565346335663061323238633364323736613939616465313934343539353363363139626665373335376634613663316335616634383833333861643333363736656166643538353338636164383532306134353838333233620a000a00")
	p, err := utils.GetProposal(raw)
	if err != nil {

	}
	p, err := protoutil.UnmarshalPayload(raw)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := protoutil.UnmarshalTransaction(p.Data)
	if err != nil {
		t.Fatal(err)
	}

	res := make([]*peer.ChaincodeAction, len(tx.Actions))
	for i, v := range tx.Actions {
		_, res[i], err = protoutil.GetPayloads(v)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(res[i].ChaincodeId.Name)
	}
	fmt.Println("")
}

type TransferEvent struct {
	From   []byte `json:"from"`
	To     []byte `json:"to"`
	Amount []byte `json:"amount"`
}
