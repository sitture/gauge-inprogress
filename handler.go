package main

import (
	"context"
	"fmt"
	"github.com/sitture/gauge-inprogress/gauge_messages"
	"github.com/sitture/gauge-inprogress/helper"
	"github.com/sitture/gauge-inprogress/logger"
	"google.golang.org/grpc"
	"os"
)

type handler struct {
	server *grpc.Server
}

func (h *handler) GenerateDocs(c context.Context, m *gauge_messages.SpecDetails) (*gauge_messages.Empty, error) {
	specDirs := os.Getenv("GAUGE_SPEC_DIRS")
	logger.Debugf("Analysing specs under %s", specDirs)
	allSpecs := helper.GetSpecs(m, helper.GetFiles(specDirs))
	allScenarios := helper.GetScenarios(allSpecs)
	inProgressSpecs, inProgressScenarios := helper.GetInProgressSpecsScenarios(allSpecs)

	fmt.Printf("\nIn progress summary: %s", helper.GetProjectDirName())
	fmt.Printf("\nSpecifications:\t%d of %d\t%d%%", len(inProgressSpecs), len(allSpecs), 10)
	fmt.Printf("\nScenarios:\t%d of %d\t%d%%", len(inProgressScenarios), len(allScenarios), 14)

	return &gauge_messages.Empty{}, nil
}

func (h *handler) Kill(c context.Context, m *gauge_messages.KillProcessRequest) (*gauge_messages.Empty, error) {
	defer h.stopServer()
	return &gauge_messages.Empty{}, nil
}

func (h *handler) stopServer() {
	h.server.Stop()
}
