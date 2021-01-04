package main

import (
	"context"
	"github.com/sitture/gauge-inprogress/gauge_messages"
	"github.com/sitture/gauge-inprogress/helper"
	"github.com/sitture/gauge-inprogress/logger"
	"google.golang.org/grpc"
	"os"
	"strings"
)

type handler struct {
	server *grpc.Server
}

func (h *handler) GenerateDocs(c context.Context, m *gauge_messages.SpecDetails) (*gauge_messages.Empty, error) {
	logger.Debugf("\n\nIn progress tags are set to %s.", helper.GetInProgressTags())
	specDirs := strings.Split(strings.TrimSpace(os.Getenv("GAUGE_SPEC_DIRS")), "||")
	logger.Debugf("Analysing specs under %s", specDirs)
	allSpecs := helper.GetSpecs(m, helper.GetSpecFiles(specDirs))
	allScenarios := helper.GetScenarios(allSpecs)
	inProgressSpecs, inProgressScenarios := helper.GetInProgressSpecsScenarios(allSpecs)
	logger.Infof(
		"In progress Summary: %s\n",
		helper.GetProjectDirName(),
	)
	if helper.PercentOf(len(inProgressSpecs), len(allSpecs)) == 0 && helper.PercentOf(len(inProgressScenarios), len(allScenarios)) == 0 {
		logger.Infof("No in progress scenarios found.")
	} else {
		logger.Infof(
			"Specifications:\t%d of %d\t%0.0f%%",
			len(inProgressSpecs),
			len(allSpecs),
			helper.PercentOf(len(inProgressSpecs), len(allSpecs)),
		)
		logger.Infof(
			"Scenarios:\t%d of %d\t%0.0f%%",
			len(inProgressScenarios),
			len(allScenarios),
			helper.PercentOf(len(inProgressScenarios), len(allScenarios)),
		)
	}
	logger.Infof("")
	return &gauge_messages.Empty{}, nil
}

func (h *handler) Kill(c context.Context, m *gauge_messages.KillProcessRequest) (*gauge_messages.Empty, error) {
	defer h.stopServer()
	return &gauge_messages.Empty{}, nil
}

func (h *handler) stopServer() {
	h.server.Stop()
}
