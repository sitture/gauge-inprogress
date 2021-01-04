package main

import (
	"github.com/sitture/gauge-inprogress/gauge_messages"
	"github.com/sitture/gauge-inprogress/helper"
	"github.com/sitture/gauge-inprogress/logger"
	"google.golang.org/grpc"
	"net"
	"os"
)

const (
	msgSize         = 1024 * 1024 * 1024
	inProgress      = "inprogress"
	PluginActionEnv = inProgress + "_action"
	DocsAction      = "documentation"
	GaugeHost       = "127.0.0.1:0"
)

func main() {
	os.Chdir(helper.GetProjectRoot())
	address, err := net.ResolveTCPAddr("tcp", GaugeHost)
	if err != nil {
		logger.Fatalf("failed to start server.", err)
	}
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		logger.Fatalf("failed to start server.", err)
	}
	server := grpc.NewServer(grpc.MaxRecvMsgSize(msgSize))
	handler := &handler{server: server}
	gauge_messages.RegisterDocumenterServer(server, handler)
	logger.Infof("Listening on port:%d", listener.Addr().(*net.TCPAddr).Port)
	server.Serve(listener)
}
