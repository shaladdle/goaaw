package rpc

import (
	"testing"

	anet "aaw/net"
)

type testServer struct{}

func (testServer) RPCNorm_Add(a, b int) int {
	return a + b
}

func TestNormalRpc(t *testing.T) {
	const (
		sName        = "Test"
		hostport     = "localhost:8080"
		arg1     int = 1
		arg2     int = 2
		want     int = 3
	)

	pnet := anet.NewPipeNet()

	rpcs := NewServer()
	rpcs.Register(sName, testServer{})

	//rpcs.TCPListen(hostport)
	go rpcs.Accept(pnet)

	//d := TcpDialer(hostport)
	d := pnet
	cli, err := NewClient(d)
	if err != nil {
		t.Errorf("client creation failed: %v", err)
	}

	methodName := sName + ".Add"
	var got int
	_, err = cli.Call(methodName, arg1, arg2, &got)
	if err != nil {
		t.Errorf("call failed: %v", err)
	}

	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}
