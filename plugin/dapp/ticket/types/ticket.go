// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"errors"
	"reflect"
	"time"

	//log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

const (
	//log for ticket

	//TyLogNewTicket new ticket log type
	TyLogNewTicket = 111
	// TyLogCloseTicket close ticket log type
	TyLogCloseTicket = 112
	// TyLogMinerTicket miner ticket log type
	TyLogMinerTicket = 113
	// TyLogTicketBind bind ticket log type
	TyLogTicketBind = 114
)

//ticket
const (
	// TicketActionGenesis action type
	TicketActionGenesis = 11
	// TicketActionOpen action type
	TicketActionOpen = 12
	// TicketActionClose action type
	TicketActionClose = 13
	// TicketActionList  action type
	TicketActionList = 14 //读的接口不直接经过transaction
	// TicketActionInfos action type
	TicketActionInfos = 15 //读的接口不直接经过transaction
	// TicketActionMiner action miner
	TicketActionMiner = 16
	// TicketActionBind action bind
	TicketActionBind = 17
)

// TicketOldParts old tick type
const TicketOldParts = 3

// TicketCountOpenOnce count open once
const TicketCountOpenOnce = 1000

// ErrOpenTicketPubHash err type
var ErrOpenTicketPubHash = errors.New("ErrOpenTicketPubHash")

// TicketX dapp name
var TicketX = "ticket"

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(TicketX))
	types.RegistorExecutor(TicketX, NewType())
	types.RegisterDappFork(TicketX, "Enable", 0)
	types.RegisterDappFork(TicketX, "ForkTicketId", 1062000)
	types.RegisterDappFork(TicketX, "ForkTicketVrf", 1770000)
}

// TicketType ticket exec type
type TicketType struct {
	types.ExecTypeBase
}

// NewType new type
func NewType() *TicketType {
	c := &TicketType{}
	c.SetChild(c)
	return c
}

// GetPayload get payload
func (ticket *TicketType) GetPayload() types.Message {
	return &TicketAction{}
}

// GetLogMap get log map
func (ticket *TicketType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogNewTicket:   {Ty: reflect.TypeOf(ReceiptTicket{}), Name: "LogNewTicket"},
		TyLogCloseTicket: {Ty: reflect.TypeOf(ReceiptTicket{}), Name: "LogCloseTicket"},
		TyLogMinerTicket: {Ty: reflect.TypeOf(ReceiptTicket{}), Name: "LogMinerTicket"},
		TyLogTicketBind:  {Ty: reflect.TypeOf(ReceiptTicketBind{}), Name: "LogTicketBind"},
	}
}

// Amount get amount
func (ticket TicketType) Amount(tx *types.Transaction) (int64, error) {
	var action TicketAction
	err := types.Decode(tx.GetPayload(), &action)
	if err != nil {
		return 0, types.ErrDecode
	}
	if action.Ty == TicketActionMiner && action.GetMiner() != nil {
		ticketMiner := action.GetMiner()
		return ticketMiner.Reward, nil
	}
	return 0, nil
}

// GetName get name
func (ticket *TicketType) GetName() string {
	return TicketX
}

// GetTypeMap get type map
func (ticket *TicketType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Genesis": TicketActionGenesis,
		"Topen":   TicketActionOpen,
		"Tbind":   TicketActionBind,
		"Tclose":  TicketActionClose,
		"Miner":   TicketActionMiner,
	}
}

// TicketMinerParam
type TicketMinerParam struct {
	CoinDevFund              int64
	CoinReward               int64
	FutureBlockTime          int64
	TicketPrice              int64
	TicketFrozenTime         int64
	TicketWithdrawTime       int64
	TicketMinerWaitTime      int64
	TargetTimespan           time.Duration
	TargetTimePerBlock       time.Duration
	RetargetAdjustmentFactor int64
}

// GetTicketMinerParam 获取ticket miner config params
func GetTicketMinerParam(height int64) *TicketMinerParam {
	conf := types.Conf("mver.consensus.ticket")
	c := &TicketMinerParam{}
	c.CoinDevFund = conf.MGInt("coinDevFund", height) * types.Coin
	c.CoinReward = conf.MGInt("coinReward", height) * types.Coin
	c.FutureBlockTime = conf.MGInt("futureBlockTime", height)
	c.TicketPrice = conf.MGInt("ticketPrice", height) * types.Coin
	c.TicketFrozenTime = conf.MGInt("ticketFrozenTime", height)
	c.TicketWithdrawTime = conf.MGInt("ticketWithdrawTime", height)
	c.TicketMinerWaitTime = conf.MGInt("ticketMinerWaitTime", height)
	c.TargetTimespan = time.Duration(conf.MGInt("targetTimespan", height)) * time.Second
	c.TargetTimePerBlock = time.Duration(conf.MGInt("targetTimePerBlock", height)) * time.Second
	c.RetargetAdjustmentFactor = conf.MGInt("retargetAdjustmentFactor", height)
	return c
}
