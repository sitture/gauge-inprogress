package main

import (
	"context"
	"fmt"
	"github.com/sitture/gauge-inprogress/gauge_messages"
	"google.golang.org/grpc"
)

type handler struct {
	server *grpc.Server
}

func (h *handler) GenerateDocs(c context.Context, m *gauge_messages.SpecDetails) (*gauge_messages.Empty, error) {
	fmt.Printf("Succesfully converted specs to html")
	return &gauge_messages.Empty{}, nil
}

func (h *handler) Kill(c context.Context, m *gauge_messages.KillProcessRequest) (*gauge_messages.Empty, error) {
	defer h.stopServer()
	return &gauge_messages.Empty{}, nil
}

func (h *handler) stopServer() {
	h.server.Stop()
}
