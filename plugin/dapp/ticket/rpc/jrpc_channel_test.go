// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

import (
	"strings"
	"testing"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	_ "github.com/33cn/plugin/plugin"
	"github.com/stretchr/testify/assert"
)

var mocker *testnode.Chain33Mock

func TestMain(m *testing.M) {
	cfg, sub := testnode.GetDefaultConfig()
	cfg.Consensus.Name = "ticket"
	mocker = testnode.NewWithConfig(cfg, sub, nil)
	mocker.Listen()
	m.Run()
	mocker.Close()
}

func TestJRPCChannel(t *testing.T) {
	jrpcClient := mocker.GetJSONC()

	testCases := []struct {
		fn func(*testing.T, *jsonclient.JSONClient) error
	}{
		{fn: testCountTicketCmd},
		{fn: testCloseTicketCmd},
		{fn: testGetColdAddrByMinerCmd},
	}
	for index, testCase := range testCases {
		err := testCase.fn(t, jrpcClient)
		if err == nil {
			continue
		}
		assert.NotEqualf(t, err, types.ErrActionNotSupport, "test index %d", index)
		if strings.Contains(err.Error(), "rpc: can't find") {
			assert.FailNowf(t, err.Error(), "test index %d", index)
		}
		t.Log(err.Error())
	}
}

func testCountTicketCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var res int64
	return jrpc.Call("ticket.GetTicketCount", nil, &res)
}

func testCloseTicketCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var res types.ReplyHashes
	return jrpc.Call("ticket.CloseTickets", nil, &res)
}

func testGetColdAddrByMinerCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &types.ReqString{}
	params.Execer = "ticket"
	params.FuncName = "MinerSourceList"
	params.Payload = types.MustPBToJSON(req)
	rep = &types.ReplyStrings{}
	return jrpc.Call("Chain33.Query", params, rep)
}
