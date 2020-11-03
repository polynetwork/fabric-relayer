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
	fabricSdk, err := tools.NewFabricSdk(servConfig.FabricConfig.Channel, servConfig.FabricConfig.Chaincode, servConfig.FabricConfig.SdkConfFile)
	if err != nil {
		log.Errorf("startServer - create fabric sdk, err: %s", err)
		return
	}

	var boltDB *db.BoltDB
	if servConfig.BoltDbPath == "" {
		boltDB, err = db.NewBoltDB("boltdb")
	} else {
		boltDB, err = db.NewBoltDB(servConfig.BoltDbPath)
	}
	if err != nil {
		log.Fatalf("db.NewWaitingDB error:%s", err)
		return
	}

	initPolyServer(servConfig, polySdk, fabricSdk, boltDB)
	initFabricServer(servConfig, polySdk, fabricSdk, boltDB)
	waitToExit()
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
