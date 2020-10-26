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
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/polynetwork/eth-contracts/go_abi/eccm_abi"
	"github.com/polynetwork/fabric-relayer/config"
	"github.com/polynetwork/fabric-relayer/db"
	"github.com/polynetwork/fabric-relayer/log"
	"github.com/polynetwork/fabric-relayer/tools"
	sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native/service/cross_chain_manager/eth"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	autils "github.com/polynetwork/poly/native/service/utils"
)

type EthereumManager struct {
	config         *config.ServiceConfig
	restClient     *tools.RestClient
	client         *ethclient.Client
	currentHeight  uint64
	forceHeight    uint64
	lockerContract *bind.BoundContract
	polySdk        *sdk.PolySdk
	polySigner     *sdk.Account
	exitChan       chan int
	header4sync    [][]byte
	crosstx4sync   []*CrossTransfer
	db             *db.BoltDB
}

func (e *EthereumManager) init() error {
	latestHeight := e.findLatestHeight()
	if latestHeight == 0 {
		return fmt.Errorf("init - the genesis block has not synced!")
	}

	if e.forceHeight > 0 && e.forceHeight < latestHeight {
		e.currentHeight = e.forceHeight
	} else {
		e.currentHeight = latestHeight
	}

	log.Infof("EthereumManager init - start height: %d", e.currentHeight)
	return nil
}

func NewEthereumManager(
	servconfig *config.ServiceConfig,
	startheight,
	startforceheight uint64,
	ontsdk *sdk.PolySdk,
	client *ethclient.Client,
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
		forceHeight:   startforceheight,
		restClient:    tools.NewRestClient(),
		client:        client,
		polySdk:       ontsdk,
		polySigner:    signer,
		header4sync:   make([][]byte, 0),
		crosstx4sync:  make([]*CrossTransfer, 0),
		db:            boltDB,
	}

	if err = mgr.init(); err != nil {
		return nil, err
	}
	return
}

func (e *EthereumManager) MonitorChain() {

	var (
		backtrace        = uint64(1)
		fetchBlockTicker = time.NewTicker(config.ETH_MONITOR_INTERVAL)
	)

	for {
		select {
		case <-fetchBlockTicker.C:
			height, err := tools.GetNodeHeight(e.config.ETHConfig.RestURL, e.restClient)
			if err != nil {
				log.Infof("MonitorChain - cannot get node height, err: %s", err)
				continue
			}
			if height-e.currentHeight <= config.ETH_USEFUL_BLOCK_NUM {
				continue
			}
			log.Infof("MonitorChain - eth height is %d", height)

			if !e.handleBlocks(height) {
				continue
			}

			e.commitLatestHeader(&backtrace)
		case <-e.exitChan:
			return
		}
	}
}

func (e *EthereumManager) findLatestHeight() uint64 {
	// try to get key
	var sideChainIdBytes [8]byte
	binary.LittleEndian.PutUint64(sideChainIdBytes[:], e.config.ETHConfig.SideChainId)
	contractAddress := autils.HeaderSyncContractAddress
	key := append([]byte(scom.CURRENT_HEADER_HEIGHT), sideChainIdBytes[:]...)

	// try to get storage
	result, err := e.polySdk.GetStorage(contractAddress.ToHexString(), key)
	if err != nil {
		return 0
	}
	if result == nil || len(result) == 0 {
		return 0
	} else {
		return binary.LittleEndian.Uint64(result)
	}
}

func (e *EthereumManager) handleBlocks(height uint64) bool {
	for e.currentHeight < height-config.ETH_USEFUL_BLOCK_NUM {
		if !e.handleNewBlock(e.currentHeight + 1) {
			return false
		}

		e.currentHeight++
		// try to commit header if more than 50 headers needed to be syned
		if len(e.header4sync) >= e.config.ETHConfig.HeadersPerBatch {
			if e.commitHeader() != 0 {
				log.Errorf("MonitorChain - commit header failed.")
				return false
			}
			e.header4sync = make([][]byte, 0)
		}
	}

	return true
}

// commitLatestHeader try to commit latest header when we are at latest height
func (e *EthereumManager) commitLatestHeader(backtrace *uint64) {
	commitHeaderResult := e.commitHeader()
	if commitHeaderResult > 0 {
		log.Errorf("MonitorChain - commit header failed.")
		return
	}

	if commitHeaderResult == 0 {
		*backtrace = 1
		e.header4sync = make([][]byte, 0)
		return
	}

	latestHeight := e.findLatestHeight()
	if latestHeight == 0 {
		return
	}

	e.currentHeight = latestHeight - *backtrace
	*backtrace += 1
	log.Errorf("MonitorChain - back to height: %d", e.currentHeight)
	e.header4sync = make([][]byte, 0)
}

func (e *EthereumManager) handleNewBlock(height uint64) bool {
	if !e.handleBlockHeader(height) {
		log.Errorf("handleNewBlock - handleBlockHeader on height :%d failed", height)
		return false
	}
	if !e.fetchLockDepositEvents(height, e.client) {
		log.Errorf("handleNewBlock - fetchLockDepositEvents on height :%d failed", height)
	}
	return true
}

func (e *EthereumManager) handleBlockHeader(height uint64) bool {
	header, err := tools.GetNodeHeader(e.config.ETHConfig.RestURL, e.restClient, height)
	if err != nil {
		log.Errorf("handleBlockHeader - GetNodeHeader on height :%d failed", height)
		return false
	}
	e.header4sync = append(e.header4sync, header)
	return true
}

func (e *EthereumManager) fetchLockDepositEvents(height uint64, client *ethclient.Client) bool {
	lockAddress := ethcommon.HexToAddress(e.config.ETHConfig.ECCMContractAddress)
	lockContract, err := eccm_abi.NewEthCrossChainManager(lockAddress, client)
	if err != nil {
		return false
	}
	opt := &bind.FilterOpts{
		Start:   height,
		End:     &height,
		Context: context.Background(),
	}
	events, err := lockContract.FilterCrossChainEvent(opt, nil)
	if err != nil {
		log.Errorf("fetchLockDepositEvents - FilterCrossChainEvent error :%s", err.Error())
		return false
	}
	if events == nil {
		log.Infof("fetchLockDepositEvents - no events found on FilterCrossChainEvent")
		return false
	}

	e.handleEvents(events, height)

	return true
}

func (e *EthereumManager) handleEvents(events *eccm_abi.EthCrossChainManagerCrossChainEventIterator, height uint64) {
	for events.Next() {
		evt := events.Event
		if !e.isTargetContract(evt) {
			continue
		}

		index := big.NewInt(0)
		index.SetBytes(evt.TxId)
		crossTx := &CrossTransfer{
			txIndex: tools.EncodeBigInt(index),
			txId:    evt.Raw.TxHash.Bytes(),
			toChain: uint32(evt.ToChainId),
			value:   []byte(evt.Rawdata),
			height:  height,
		}
		sink := common.NewZeroCopySink(nil)
		crossTx.Serialization(sink)

		if err := e.db.PutRetry(sink.Bytes()); err != nil {
			log.Errorf("fetchLockDepositEvents - e.db.PutRetry error: %s", err)
		}
		log.Infof("fetchLockDepositEvent -  height: %d", height)
	}
}

func (e *EthereumManager) isTargetContract(evt *eccm_abi.EthCrossChainManagerCrossChainEvent) bool {
	if len(e.config.TargetContracts) == 0 {
		return false
	}
	toContractStr := evt.ProxyOrAssetContract.String()
	for _, v := range e.config.TargetContracts {
		toChainIdArr, ok := v[toContractStr]
		if !ok {
			continue
		}
		if len(toChainIdArr["outbound"]) == 0 {
			return true
		}
		for _, id := range toChainIdArr["outbound"] {
			if id == evt.ToChainId {
				return true
			}
		}
	}
	return false
}

func (e *EthereumManager) commitHeader() int {
	tx, err := e.polySdk.Native.Hs.SyncBlockHeader(
		e.config.ETHConfig.SideChainId,
		e.polySigner.Address,
		e.header4sync,
		e.polySigner,
	)
	if err != nil {
		log.Warnf("commitHeader - send transaction to poly chain err: %s!", err.Error())

		if strings.Contains(err.Error(), "get the parent block failed") ||
			strings.Contains(err.Error(), "missing required field") {
			return -1
		} else {
			return 1
		}
	}

	tick := time.NewTicker(100 * time.Millisecond)
	var h uint32
	for range tick.C {
		h, _ = e.polySdk.GetBlockHeightByTxHash(tx.ToHexString())
		curr, _ := e.polySdk.GetCurrentBlockHeight()
		if h > 0 && curr > h {
			break
		}
	}

	log.Infof("commitHeader - send transaction %s to poly chain and confirmed on height %d", tx.ToHexString(), h)
	return 0
}

func (e *EthereumManager) MonitorDeposit() {
	monitorTicker := time.NewTicker(config.ETH_MONITOR_INTERVAL)
	for {
		select {
		case <-monitorTicker.C:
			height, err := tools.GetNodeHeight(e.config.ETHConfig.RestURL, e.restClient)
			if err != nil {
				log.Infof("MonitorChain - cannot get node height, err: %s", err)
				continue
			}
			snycheight := e.findLatestHeight()
			if snycheight > height-config.ETH_PROOF_USERFUL_BLOCK {
				// try to handle deposit event when we are at latest height
				e.handleLockDepositEvents(snycheight)
			}
		case <-e.exitChan:
			return
		}
	}
}

func (e *EthereumManager) handleLockDepositEvents(refHeight uint64) error {
	retryList, err := e.db.GetAllRetry()
	if err != nil {
		return fmt.Errorf("handleLockDepositEvents - e.db.GetAllRetry error: %s", err)
	}

	for _, v := range retryList {
		time.Sleep(time.Second * 1)
		crosstx := new(CrossTransfer)
		err := crosstx.Deserialization(common.NewZeroCopySource(v))
		if err != nil {
			log.Errorf("handleLockDepositEvents - retry.Deserialization error: %s", err)
			continue
		}
		//1. decode events
		key := crosstx.txIndex
		keyBytes, err := eth.MappingKeyAt(key, "01")
		if err != nil {
			log.Errorf("handleLockDepositEvents - MappingKeyAt error:%s\n", err.Error())
			continue
		}
		if refHeight <= crosstx.height+e.config.ETHConfig.BlockConfig {
			continue
		}
		height := int64(refHeight - e.config.ETHConfig.BlockConfig)
		heightHex := hexutil.EncodeBig(big.NewInt(height))
		proofKey := hexutil.Encode(keyBytes)
		//2. get proof
		proof, err := tools.GetProof(e.config.ETHConfig.RestURL, e.config.ETHConfig.ECCDContractAddress, proofKey, heightHex, e.restClient)
		if err != nil {
			log.Errorf("handleLockDepositEvents - error :%s\n", err.Error())
			continue
		}
		//3. commit proof to poly
		txHash, err := e.commitProof(uint32(height), proof, crosstx.value, crosstx.txId)
		if err != nil {
			if strings.Contains(err.Error(), "chooseUtxos, current utxo is not enough") {
				log.Infof("handleLockDepositEvents - invokeNativeContract error: %s", err)
				continue
			} else {
				if err := e.db.DeleteRetry(v); err != nil {
					log.Errorf("handleLockDepositEvents - e.db.DeleteRetry error: %s", err)
				}
				if strings.Contains(err.Error(), "tx already done") {
					log.Debugf("handleLockDepositEvents - eth_tx %s already on poly", ethcommon.BytesToHash(crosstx.txId).String())
				} else {
					log.Errorf("handleLockDepositEvents - invokeNativeContract error for eth_tx %s: %s", ethcommon.BytesToHash(crosstx.txId).String(), err)
				}
				continue
			}
		}
		//4. put to check db for checking
		if err := e.db.PutCheck(txHash, v); err != nil {
			log.Errorf("handleLockDepositEvents - e.db.PutCheck error: %s", err)
		}
		if err := e.db.DeleteRetry(v); err != nil {
			log.Errorf("handleLockDepositEvents - e.db.PutCheck error: %s", err)
		}
		log.Infof("handleLockDepositEvents - syncProofToAlia txHash is %s", txHash)
	}
	return nil
}

func (e *EthereumManager) commitProof(height uint32, proof []byte, value []byte, txhash []byte) (string, error) {
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

func (e *EthereumManager) parserValue(value []byte) []byte {
	source := common.NewZeroCopySource(value)
	txHash, eof := source.NextVarBytes()
	if eof {
		fmt.Printf("parserValue - deserialize txHash error")
	}
	return txHash
}

func (e *EthereumManager) CheckDeposit() {
	checkTicker := time.NewTicker(config.ETH_MONITOR_INTERVAL)
	for {
		select {
		case <-checkTicker.C:
			e.checkLockDepositEvents()
		case <-e.exitChan:
			return
		}
	}
}

func (e *EthereumManager) checkLockDepositEvents() {
	checkMap, err := e.db.GetAllCheck()
	if err != nil {
		log.Errorf("checkLockDepositEvents - e.db.GetAllCheck error: %s", err)
		return
	}

	for k, v := range checkMap {
		event, err := e.polySdk.GetSmartContractEvent(k)
		if err != nil {
			log.Errorf("checkLockDepositEvents - e.aliaSdk.GetSmartContractEvent error: %s", err)
			continue
		}

		if event.State != 1 {
			log.Infof("checkLockDepositEvents - state of poly tx %s is not success", k)
			if err := e.db.PutRetry(v); err != nil {
				log.Errorf("checkLockDepositEvents - e.db.PutRetry error:%s", err)
			}
		}

		if err := e.db.DeleteCheck(k); err != nil {
			log.Errorf("checkLockDepositEvents - e.db.DeleteRetry error:%s", err)
		}
	}
}
