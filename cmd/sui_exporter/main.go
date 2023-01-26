package main

import (
	"context"
	"github.com/ybbus/jsonrpc/v3"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	rpcClient := jsonrpc.NewClient("https://testnet.sui.rpcpool.com")

	var state *SuiSystemState
	err := rpcClient.CallFor(context.Background(), &state, "sui_getSuiSystemState")
	if err != nil {
		spew.Dump(err)
	}

	spew.Dump(state)
}
