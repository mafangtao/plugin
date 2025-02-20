// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"encoding/hex"

	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
)

var (
	clog                    = log.New("module", "execs.paracross")
	enableParacrossTransfer = true
	driverName              = pt.ParaX
)

// Paracross exec
type Paracross struct {
	drivers.DriverBase
}

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Paracross{}))
}

//Init paracross exec register
func Init(name string, sub []byte) {
	drivers.Register(GetName(), newParacross, types.GetDappFork(driverName, "Enable"))
	setPrefix()
}

//GetName return paracross name
func GetName() string {
	return newParacross().GetName()
}

func newParacross() drivers.Driver {
	c := &Paracross{}
	c.SetChild(c)
	c.SetExecutorType(types.LoadExecutorType(driverName))
	return c
}

// GetDriverName return paracross driver name
func (c *Paracross) GetDriverName() string {
	return pt.ParaX
}

func (c *Paracross) checkTxGroup(tx *types.Transaction, index int) ([]*types.Transaction, error) {
	if tx.GroupCount >= 2 {
		txs, err := c.GetTxGroup(index)
		if err != nil {
			clog.Error("ParacrossActionAssetTransfer", "get tx group failed", err, "hash", hex.EncodeToString(tx.Hash()))
			return nil, err
		}
		return txs, nil
	}
	return nil, nil
}

func (c *Paracross) saveLocalParaTxs(tx *types.Transaction, isDel bool) (*types.LocalDBSet, error) {

	var payload pt.ParacrossAction
	err := types.Decode(tx.Payload, &payload)
	if err != nil {
		return nil, err
	}
	if payload.Ty != pt.ParacrossActionCommit || payload.GetCommit() == nil {
		return nil, nil
	}

	commit := payload.GetCommit()
	crossTxHashs, crossTxResult, err := getCrossTxHashs(c.GetAPI(), commit.Status)
	if err != nil {
		return nil, err
	}

	return c.udpateLocalParaTxs(commit.Status.Title, commit.Status.Height, crossTxHashs, crossTxResult, isDel)

}

//无法获取到commit tx信息，从commitDone 结构里面构建
func (c *Paracross) saveLocalParaTxsFork(commitDone *pt.ReceiptParacrossDone, isDel bool) (*types.LocalDBSet, error) {
	status := &pt.ParacrossNodeStatus{
		MainBlockHash:   commitDone.MainBlockHash,
		MainBlockHeight: commitDone.MainBlockHeight,
		Title:           commitDone.Title,
		Height:          commitDone.Height,
		BlockHash:       commitDone.BlockHash,
		TxResult:        commitDone.TxResult,
	}

	crossTxHashs, crossTxResult, err := getCrossTxHashs(c.GetAPI(), status)
	if err != nil {
		return nil, err
	}

	return c.udpateLocalParaTxs(commitDone.Title, commitDone.Height, crossTxHashs, crossTxResult, isDel)

}

func (c *Paracross) udpateLocalParaTxs(paraTitle string, paraHeight int64, crossTxHashs [][]byte, crossTxResult []byte, isDel bool) (*types.LocalDBSet, error) {
	var set types.LocalDBSet

	if len(crossTxHashs) == 0 {
		return &set, nil
	}

	for i := 0; i < len(crossTxHashs); i++ {
		success := util.BitMapBit(crossTxResult, uint32(i))

		paraTx, err := GetTx(c.GetAPI(), crossTxHashs[i])
		if err != nil {
			clog.Crit("paracross.Commit Load Tx failed", "para title", paraTitle,
				"para height", paraHeight, "para tx index", i, "error", err, "txHash",
				hex.EncodeToString(crossTxHashs[i]))
			return nil, err
		}

		var payload pt.ParacrossAction
		err = types.Decode(paraTx.Tx.Payload, &payload)
		if err != nil {
			clog.Crit("paracross.Commit Decode Tx failed", "para title", paraTitle,
				"para height", paraHeight, "para tx index", i, "error", err, "txHash",
				hex.EncodeToString(crossTxHashs[i]))
			return nil, err
		}
		if payload.Ty == pt.ParacrossActionAssetTransfer {
			kv, err := c.updateLocalAssetTransfer(paraHeight, paraTx.Tx, success, isDel)
			if err != nil {
				return nil, err
			}
			set.KV = append(set.KV, kv)
		} else if payload.Ty == pt.ParacrossActionAssetWithdraw {
			kv, err := c.initLocalAssetWithdraw(paraHeight, paraTx.Tx, true, success, isDel)
			if err != nil {
				return nil, err
			}
			set.KV = append(set.KV, kv)
		}
	}

	return &set, nil
}

func (c *Paracross) initLocalAssetTransfer(tx *types.Transaction, success, isDel bool) (*types.KeyValue, error) {
	clog.Debug("para execLocal", "tx hash", hex.EncodeToString(tx.Hash()), "action name", log.Lazy{Fn: tx.ActionName})
	key := calcLocalAssetKey(tx.Hash())
	if isDel {
		c.GetLocalDB().Set(key, nil)
		return &types.KeyValue{Key: key, Value: nil}, nil
	}

	var payload pt.ParacrossAction
	err := types.Decode(tx.Payload, &payload)
	if err != nil {
		return nil, err
	}
	if payload.GetAssetTransfer() == nil {
		return nil, errors.New("GetAssetTransfer is nil")
	}
	exec := "coins"
	symbol := types.BTY
	if payload.GetAssetTransfer().Cointoken != "" {
		exec = "token"
		symbol = payload.GetAssetTransfer().Cointoken
	}

	var asset pt.ParacrossAsset
	amount, err := tx.Amount()
	if err != nil {
		return nil, err
	}

	asset = pt.ParacrossAsset{
		From:       tx.From(),
		To:         tx.To,
		Amount:     amount,
		IsWithdraw: false,
		TxHash:     tx.Hash(),
		Height:     c.GetHeight(),
		Exec:       exec,
		Symbol:     symbol,
	}

	err = c.GetLocalDB().Set(key, types.Encode(&asset))
	if err != nil {
		clog.Error("para execLocal", "set", hex.EncodeToString(tx.Hash()), "failed", err)
	}
	return &types.KeyValue{Key: key, Value: types.Encode(&asset)}, nil
}

func (c *Paracross) initLocalAssetWithdraw(paraHeight int64, tx *types.Transaction, isWithdraw, success, isDel bool) (*types.KeyValue, error) {
	key := calcLocalAssetKey(tx.Hash())
	if isDel {
		c.GetLocalDB().Set(key, nil)
		return &types.KeyValue{Key: key, Value: nil}, nil
	}

	var asset pt.ParacrossAsset

	amount, err := tx.Amount()
	if err != nil {
		return nil, err
	}
	asset.ParaHeight = paraHeight

	var payload pt.ParacrossAction
	err = types.Decode(tx.Payload, &payload)
	if err != nil {
		return nil, err
	}
	if payload.GetAssetWithdraw() == nil {
		return nil, errors.New("GetAssetWithdraw is nil")
	}
	exec := "coins"
	symbol := types.BTY
	if payload.GetAssetWithdraw().Cointoken != "" {
		exec = "token"
		symbol = payload.GetAssetWithdraw().Cointoken
	}

	asset = pt.ParacrossAsset{
		From:       tx.From(),
		To:         tx.To,
		Amount:     amount,
		IsWithdraw: isWithdraw,
		TxHash:     tx.Hash(),
		Height:     c.GetHeight(),
		Exec:       exec,
		Symbol:     symbol,
	}

	asset.CommitDoneHeight = c.GetHeight()
	asset.Success = success

	err = c.GetLocalDB().Set(key, types.Encode(&asset))
	if err != nil {
		clog.Error("para execLocal", "set", "", "failed", err)
	}
	return &types.KeyValue{Key: key, Value: types.Encode(&asset)}, nil
}

func (c *Paracross) updateLocalAssetTransfer(paraHeight int64, tx *types.Transaction, success, isDel bool) (*types.KeyValue, error) {
	clog.Debug("para execLocal", "tx hash", hex.EncodeToString(tx.Hash()))
	key := calcLocalAssetKey(tx.Hash())

	var asset pt.ParacrossAsset
	v, err := c.GetLocalDB().Get(key)
	if err != nil {
		return nil, err
	}
	err = types.Decode(v, &asset)
	if err != nil {
		panic(err)
	}
	if !isDel {
		asset.ParaHeight = paraHeight

		asset.CommitDoneHeight = c.GetHeight()
		asset.Success = success
	} else {
		asset.ParaHeight = 0
		asset.CommitDoneHeight = 0
		asset.Success = false
	}
	c.GetLocalDB().Set(key, types.Encode(&asset))
	return &types.KeyValue{Key: key, Value: types.Encode(&asset)}, nil
}

//IsFriend call exec is same seariase exec
func (c *Paracross) IsFriend(myexec, writekey []byte, tx *types.Transaction) bool {
	//不允许平行链
	if types.IsPara() {
		return false
	}
	//friend 调用必须是自己在调用
	if string(myexec) != c.GetDriverName() {
		return false
	}
	//只允许同系列的执行器（tx 也必须是 paracross）
	if string(types.GetRealExecName(tx.Execer)) != c.GetDriverName() {
		return false
	}
	//只允许跨链交易
	return c.allow(tx, 0) == nil
}

func (c *Paracross) allow(tx *types.Transaction, index int) error {
	// 增加新的规则: 在主链执行器带着title的 asset-transfer/asset-withdraw 交易允许执行
	// 1. user.p.${tilte}.${paraX}
	// 1. payload 的 actionType = t/w
	if !types.IsPara() && c.allowIsParaTx(tx.Execer) {
		var payload pt.ParacrossAction
		err := types.Decode(tx.Payload, &payload)
		if err != nil {
			return err
		}
		if payload.Ty == pt.ParacrossActionAssetTransfer || payload.Ty == pt.ParacrossActionAssetWithdraw {
			return nil
		}
		if types.IsDappFork(c.GetHeight(), pt.ParaX, pt.ForkCommitTx) {
			if payload.Ty == pt.ParacrossActionCommit || payload.Ty == pt.ParacrossActionNodeConfig ||
				payload.Ty == pt.ParacrossActionNodeGroupApply {
				return nil
			}
		}
	}
	return types.ErrNotAllow
}

// Allow add paracross allow rule
func (c *Paracross) Allow(tx *types.Transaction, index int) error {
	//默认规则
	err := c.DriverBase.Allow(tx, index)
	if err == nil {
		return nil
	}
	//paracross 添加的规则
	return c.allow(tx, index)
}

func (c *Paracross) allowIsParaTx(execer []byte) bool {
	if !bytes.HasPrefix(execer, types.ParaKey) {
		return false
	}
	count := 0
	index := 0
	s := len(types.ParaKey)
	for i := s; i < len(execer); i++ {
		if execer[i] == '.' {
			count++
			index = i
		}
	}
	if count == 1 && c.AllowIsSame(execer[index+1:]) {
		return true
	}
	return false
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (c *Paracross) CheckReceiptExecOk() bool {
	return true
}
