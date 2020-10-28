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
package manager

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/polynetwork/fabric-relayer/config"
	"github.com/polynetwork/fabric-relayer/db"
	"github.com/polynetwork/fabric-relayer/log"
	"github.com/polynetwork/fabric-relayer/tools"
	sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/common"
)

type EthereumManager struct {
	config         *config.ServiceConfig
	client         *tools.FabricSdk
	currentHeight  uint64
	lockerContract *bind.BoundContract
	polySdk        *sdk.PolySdk
	polySigner     *sdk.Account
	exitChan       chan int
	header4sync    [][]byte
	crosstx4sync   []*CrossTransfer
	db             *db.BoltDB
}

func NewEthereumManager(
	servconfig *config.ServiceConfig,
	startheight uint64,
	ontsdk *sdk.PolySdk,
	client *tools.FabricSdk,
	boltDB *db.BoltDB,
) (mgr *EthereumManager, err error) {

	var (
		wallet *sdk.Wallet
		signer *sdk.Account

		wf  = servconfig.PolyConfig.WalletFile
		pwd = []byte(servconfig.PolyConfig.WalletPwd)
	)

	// open wallet with poly sdk
	if !common.FileExisted(wf) {
		wallet, err = ontsdk.CreateWallet(wf)
	} else {
		wallet, err = ontsdk.OpenWallet(wf)
	}
	if err != nil {
		log.Errorf("NewETHManager - wallet open error: %s", err.Error())
		return
	}

	// get account
	if signer, err = wallet.GetDefaultAccount(pwd); err != nil || signer == nil {
		signer, err = wallet.NewDefaultSettingAccount(pwd)
	}
	if err != nil {
		log.Errorf("NewETHManager - wallet password error")
		return nil, err
	}

	// save wallet
	if err = wallet.Save(); err != nil {
		return
	}

	log.Infof("NewETHManager - poly address: %s", signer.Address.ToBase58())

	mgr = &EthereumManager{
		config:        servconfig,
		exitChan:      make(chan int),
		currentHeight: startheight,
		client:        client,
		polySdk:       ontsdk,
		polySigner:    signer,
		db:            boltDB,
	}
	return mgr, nil
}

func (e *EthereumManager) MonitorChain() {
	reg, notifier, err := e.client.RegisterCrossChainEvent()
	if err != nil {
		log.Errorf("failed to register cc event!")
	}
	defer e.client.Unregister(reg)

	select {
	case ccEvent := <- notifier:
		fmt.Printf("receive cc event:%v\n", ccEvent)
	}
}

func (e *EthereumManager) commitCrossChainEvent(height uint32, proof []byte, value []byte, txhash []byte) (string, error) {
	log.Debugf("commit proof, height: %d, proof: %s, value: %s, txhash: %s", height, string(proof), hex.EncodeToString(value), hex.EncodeToString(txhash))
	tx, err := e.polySdk.Native.Ccm.ImportOuterTransfer(
		e.config.ETHConfig.SideChainId,
		value,
		height,
		proof,
		ethcommon.Hex2Bytes(e.polySigner.Address.ToHexString()),
		[]byte{},
		e.polySigner)
	if err != nil {
		return "", err
	} else {
		log.Infof("commitProof - send transaction to poly chain: ( poly_txhash: %s, eth_txhash: %s, height: %d )",
			tx.ToHexString(), ethcommon.BytesToHash(txhash).String(), height)
		return tx.ToHexString(), nil
	}
}

