/*
* Copyright (C) 2020 The poly network Authors
* This file is part of The poly network library.
*
* The poly network is free software: you can redistribute it and/or modify
* it under the terms of the GNU Lesser General Public License as published by
* the Free Software Foundation, either version 3 of the License, or
* (at your option) any later version.
*
* The poly network is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
* GNU Lesser General Public License for more details.
* You should have received a copy of the GNU Lesser General Public License
* along with The poly network . If not, see <http://www.gnu.org/licenses/>.
 */
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/polynetwork/fabric-relayer/tools"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/polynetwork/fabric-relayer/config"
	"github.com/polynetwork/fabric-relayer/db"
	"github.com/polynetwork/fabric-relayer/log"
	"github.com/polynetwork/fabric-relayer/manager"
	sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/urfave/cli"
)

var (
	ConfigPath       string
	LogDir           string
)

func setupApp() *cli.App {
	app := cli.NewApp()
	app.Usage = "Fabric relayer Service"
	app.Action = startServer
	app.Version = config.Version
	app.Copyright = "Copyright in 2019 The Ontology Authors"
	app.Flags = []cli.Flag{
		logLevelFlag,
		configPathFlag,
		logDirFlag,
	}
	app.Commands = []cli.Command{}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	return app
}

func startServer(ctx *cli.Context) {
	// get all cmd flag
	logLevel := ctx.GlobalInt(getFlagName(logLevelFlag))
	ld := ctx.GlobalString(getFlagName(logDirFlag))
	log.InitLog(logLevel, ld, log.Stdout)
	ConfigPath = ctx.GlobalString(getFlagName(configPathFlag))

	dir, err := os.Getwd()
	if err != nil {
		log.Errorf("startServer - get current work directory failed!")
		return
	}
	os.Setenv("FABRIC_RELAYER_PATH", dir)

	// read config
	servConfig := config.NewServiceConfig(ConfigPath)
	if servConfig == nil {
		log.Errorf("startServer - read config failed!")
		return
	}

	// create poly sdk
	polySdk := sdk.NewPolySdk()
	err = setUpPoly(polySdk, servConfig.PolyConfig.RestURL)
	if err != nil {
		log.Errorf("startServer - failed to setup poly sdk: %v", err)
		return
	}

	// create fabric sdk
	fabricSdk, err := tools.NewFabricSdk(servConfig.FabricConfig.Channel,
		servConfig.FabricConfig.Chaincode, servConfig.FabricConfig.SdkConfFile,
		servConfig.FabricConfig.UserName, servConfig.FabricConfig.OrgName)
	if err != nil {
		log.Errorf("startServer - create fabric sdk, err: %s", err)
		return
	}

	if logLevel > 4 {
		height, err := fabricSdk.GetLatestSyncHeight()
		if err != nil {
			panic(err)
		}
		fmt.Printf("latest sync height: %d\n", height)
	} else {
		aa, _ := hex.DecodeString("a620ceb0c6d229def7d25421deb341d1a9d429abb8363e3a771a7989b2b431dad9ae060000000000000020000000000000000000000000000000000000000000000000000000000000004d20f3efe3549f6f902c2595676eb8c00c43d8066c754ea0876c2b5f76c83272ca381469d0ba0866ee3d9abd19b06ad8ac6f49023e19b8080000000000000016050341a2c2c7ace8df73838591284da26b1a9771fff804686561720148012b2d92fa85b0d94331dda6d1ff508fa8d5a60fe69545c07f2d335ef802cadd21")
		bb, _ := hex.DecodeString("00000000db056dd100000000f944e577a3336ca686081540533798edd35bc487493b92f9d7b59fed5090748b632a40c99150adeee1f3eb78eb2cc6b08405a4ef0bdcd974b4988e5849f681c5d4077dcf02da4fbebb36a5322837b430fcdcd361864f79edeb15062bfc103ff2b251b7079192a472d56b2b33e6f6e90d96f3bd2ed24c42185435a16ce763856ef860a25fc0be0000cbd2c97ea0a145fffd0c017b226c6561646572223a342c227672665f76616c7565223a22424338754a486e644a55614f5a7946306c697069567234385646544f46596567435a7350596c454b2b31576c6c5652626b424e367151715045504a7876546f4646586d532b484b6951495366444e74436a7658343368553d222c227672665f70726f6f66223a223672504e6e45653574307230467055465059777a6a526f354e4d6f41723375757576314a575a69306d59706c6c427657664e516b584d7749343645386e536c4b2f427969426b5942413374492f2b2f66306b794b6d773d3d222c226c6173745f636f6e6669675f626c6f636b5f6e756d223a302c226e65775f636861696e5f636f6e666967223a6e756c6c7d00000000000000000000000000000000000000000423120502679930a42aaf3c69798ca8a3f12e134c019405818d783d11748e039de8515988231205028172918540b2b512eae1872a2a2e3a28d989c60d95dab8829ada7d7dd706d658231205038b8af6210ecfdcbcab22552ef8d8cf41c6f86f9cf9ab53d865741cfdb833f06b23120502482acb6564b19b90653f6e9c806292e8aa83f78e7a9382a24a6efe41c0c06f390442011c5f78e4f988892f8dd31b6faef31f4e6a4a5dfd283d9ae6228d11ede1f7f044f17914baf04e02d8d32e7a7ae35af58ae61508faa12e6aebb740ef1bdcdeec48bd42011cf084bdbd297bb8786b0e0c4df9fca04df43a6c6fb52d8baade6883a4e27bcd3f2d3056db2cb54104eb7c83d32e38fe96eb049fb5fcc74502dde88d70c026eb8942011ca900e3cbb3d7dfa6a343df7d6ba99f96c4155259f5fa2164072e4dca854c47e66bd0bfdd5d4e399c3e1e783821eb1ef20e4fb15f452880c6417ff0f79b7d0c2942011b8743c33109da321447d0bad19d2816e06c65627f3661d371f62adad8ef6b8c8d2e1ad208fb52e53f4f7d9170cb1bab7ee3d151bcdce22172f9846e2379b5f804")
		fabricSdk.CrossChainTransfer(aa, bb, []byte{}, []byte{})
	}
}

func setUpPoly(poly *sdk.PolySdk, RpcAddr string) error {
	poly.NewRpcClient().SetAddress(RpcAddr)
	hdr, err := poly.GetHeaderByHeight(0)
	if err != nil {
		return err
	}
	poly.SetChainId(hdr.ChainID)
	return nil
}

func waitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("waitToExit - Fabric relayer received exit signal:%v.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}

func initFabricServer(servConfig *config.ServiceConfig, polysdk *sdk.PolySdk, fabricsdk *tools.FabricSdk, boltDB *db.BoltDB) {
	mgr, err := manager.NewFabricManager(servConfig, polysdk, fabricsdk, boltDB)
	if err != nil {
		log.Error("initFabricServer - fabric service start err: %s", err.Error())
		return
	}
	go mgr.MonitorChain()
}

func initPolyServer(servConfig *config.ServiceConfig, polysdk *sdk.PolySdk, fabricsdk *tools.FabricSdk, boltDB *db.BoltDB) {
	mgr, err := manager.NewPolyManager(servConfig, polysdk, fabricsdk, boltDB)
	if err != nil {
		log.Error("initPolyServer - poly service start failed: %v", err)
		return
	}
	go mgr.MonitorChain()
}

func main() {
	log.Infof("main - Fabric relayer starting...")
	if err := setupApp().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
