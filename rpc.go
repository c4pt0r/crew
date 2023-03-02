package main

import (
	"context"

	jsonrpc "github.com/ybbus/jsonrpc/v3"
)

/* RPC helper functions */
type RpcRender interface {
	Render(url string, params map[string]string) ([]byte, error)
}

type JsonRPCRender struct {
	addr string
}

func NewJsonRPCRender(addr string) *JsonRPCRender {
	return &JsonRPCRender{
		addr: addr,
	}
}

func (r *JsonRPCRender) Render(url string, params map[string]string) ([]byte, error) {
	// TODO: use cache
	rpcClient := jsonrpc.NewClient(r.addr)
	resp, err := rpcClient.Call(context.Background(), "Render", url, params)
	if err != nil {
		return nil, err
	}
	s, err := resp.GetString()
	if err != nil {
		return nil, err
	}
	return []byte(s), nil

}
