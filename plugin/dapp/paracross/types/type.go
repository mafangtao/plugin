// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"reflect"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var (
	// ParaX paracross exec name
	ParaX = "paracross"
	glog  = log.New("module", ParaX)
	// ForkCommitTx main chain support paracross commit tx
	ForkCommitTx = "ForkParacrossCommitTx"
	// MainForkParacrossCommitTx 平行链配置项对应主链的ForkCommitTx 高度
	MainForkParacrossCommitTx = "MainForkParacrossCommitTx"
	// ParaSelfConsensForkHeight para self consens height string
	ParaSelfConsensForkHeight = "MainParaSelfConsensusForkHeight"
	// ForkLoopCheckCommitTxDone 循环检查共识交易done的fork
	ForkLoopCheckCommitTxDone = "ForkLoopCheckCommitTxDone"
	// MainLoopCheckCommitTxDoneForkHeight 平行链的配置项，对应主链的ForkLoopCheckCommitTxDone高度
	MainLoopCheckCommitTxDoneForkHeight = "MainLoopCheckCommitTxDoneForkHeight"
)

func init() {
	// init executor type
	types.AllowUserExec = append(types.AllowUserExec, []byte(ParaX))
	types.RegistorExecutor(ParaX, NewType())
	types.RegisterDappFork(ParaX, "Enable", 0)
	types.RegisterDappFork(ParaX, "ForkParacrossWithdrawFromParachain", 1298600)
	types.RegisterDappFork(ParaX, ForkCommitTx, 1850000)
	types.RegisterDappFork(ParaX, ForkLoopCheckCommitTxDone, 3230000)
}

// GetExecName get para exec name
func GetExecName() string {
	return types.ExecName(ParaX)
}

// ParacrossType base paracross type
type ParacrossType struct {
	types.ExecTypeBase
}

// NewType get paracross type
func NewType() *ParacrossType {
	c := &ParacrossType{}
	c.SetChild(c)
	return c
}

// GetName 获取执行器名称
func (p *ParacrossType) GetName() string {
	return ParaX
}

// GetLogMap get receipt log map
func (p *ParacrossType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogParacrossCommit:           {Ty: reflect.TypeOf(ReceiptParacrossCommit{}), Name: "LogParacrossCommit"},
		TyLogParacrossCommitDone:       {Ty: reflect.TypeOf(ReceiptParacrossDone{}), Name: "LogParacrossCommitDone"},
		TyLogParacrossCommitRecord:     {Ty: reflect.TypeOf(ReceiptParacrossRecord{}), Name: "LogParacrossCommitRecord"},
		TyLogParaAssetWithdraw:         {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogParaAssetWithdraw"},
		TyLogParaAssetTransfer:         {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogParaAssetTransfer"},
		TyLogParaAssetDeposit:          {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogParaAssetDeposit"},
		TyLogParacrossMiner:            {Ty: reflect.TypeOf(ReceiptParacrossMiner{}), Name: "LogParacrossMiner"},
		TyLogParaNodeConfig:            {Ty: reflect.TypeOf(ReceiptParaNodeConfig{}), Name: "LogParaNodeConfig"},
		TyLogParaNodeStatusUpdate:      {Ty: reflect.TypeOf(ReceiptParaNodeAddrStatUpdate{}), Name: "LogParaNodeAddrStatUpdate"},
		TyLogParaNodeGroupAddrsUpdate:  {Ty: reflect.TypeOf(types.ReceiptConfig{}), Name: "LogParaNodeGroupAddrsUpdate"},
		TyLogParaNodeVoteDone:          {Ty: reflect.TypeOf(ReceiptParaNodeVoteDone{}), Name: "LogParaNodeVoteDone"},
		TyLogParaNodeGroupConfig:       {Ty: reflect.TypeOf(ReceiptParaNodeGroupConfig{}), Name: "LogParaNodeGroupConfig"},
		TyLogParaNodeGroupStatusUpdate: {Ty: reflect.TypeOf(ReceiptParaNodeGroupConfig{}), Name: "LogParaNodeGroupStatusUpdate"},
	}
}

// GetTypeMap get action type
func (p *ParacrossType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Commit":          ParacrossActionCommit,
		"Miner":           ParacrossActionMiner,
		"AssetTransfer":   ParacrossActionAssetTransfer,
		"AssetWithdraw":   ParacrossActionAssetWithdraw,
		"Transfer":        ParacrossActionTransfer,
		"Withdraw":        ParacrossActionWithdraw,
		"TransferToExec":  ParacrossActionTransferToExec,
		"NodeConfig":      ParacrossActionNodeConfig,
		"NodeGroupConfig": ParacrossActionNodeGroupApply,
	}
}

// GetPayload paracross get action payload
func (p *ParacrossType) GetPayload() types.Message {
	return &ParacrossAction{}
}

// CreateTx paracross create tx by different action
func (p ParacrossType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	if action == "ParacrossCommit" {
		var param paracrossCommitTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			glog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		return createRawParacrossCommitTx(&param)
	} else if action == "ParacrossAssetTransfer" || action == "ParacrossAssetWithdraw" {
		var param types.CreateTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			glog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawAssetTransferTx(&param)

	} else if action == "ParacrossTransfer" || action == "Transfer" ||
		action == "ParacrossWithdraw" || action == "Withdraw" ||
		action == "ParacrossTransferToExec" || action == "TransferToExec" {

		return p.CreateRawTransferTx(action, message)
	} else if action == "NodeConfig" {
		if !types.IsPara() {
			return nil, types.ErrNotSupport
		}
		var param ParaNodeAddrConfig
		err := types.JSONToPB(message, &param)
		if err != nil {
			glog.Error("CreateTx.NodeConfig", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawNodeConfigTx(&param)
	} else if action == "NodeGroupConfig" {
		if !types.IsPara() {
			return nil, types.ErrNotSupport
		}
		var param ParaNodeGroupConfig
		err := types.JSONToPB(message, &param)
		//err := json.Unmarshal(message, &param)
		if err != nil {
			glog.Error("CreateTx.NodeGroupApply", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawNodeGroupApplyTx(&param)
	}

	return nil, types.ErrNotSupport
}
