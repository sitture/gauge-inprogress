package main

import (
	"context"
	"github.com/sitture/gauge-inprogress/gauge_messages"
	"github.com/sitture/gauge-inprogress/inprogress"
	"github.com/sitture/gauge-inprogress/logger"
	"google.golang.org/grpc"
)

type handler struct {
	server *grpc.Server
}

func (h *handler) GenerateDocs(c context.Context, m *gauge_messages.SpecDetails) (*gauge_messages.Empty, error) {
	logger.Debugf("In progress tags are set to %s.", inprogress.GetInProgressTags())
	specDirs := inprogress.GetSpecDirs()
	logger.Debugf("Analysing specs under %s", specDirs)
	allSpecs := inprogress.GetSpecs(m, inprogress.GetSpecFiles(specDirs))
	allScenarios := inprogress.GetScenarios(allSpecs)
	inProgressSpecs := inprogress.GetInProgressSpecs(allSpecs)
	inProgressScenarios := inprogress.GetInProgressScenarios(inProgressSpecs)
	inProgressSpecsWithReason := inprogress.GetInProgressSpecsWithReason(inProgressSpecs)
	inProgressScenariosWithReason := inprogress.GetInProgressScenariosWithReason(inProgressSpecsWithReason)

	if err := inprogress.WriteToFile(inProgressSpecs, inProgressScenariosWithReason); err != nil {
		logger.Debugf("Could not generate the inprogress report file.")
	}

	logger.Infof(
		"\nIn progress Summary: %s\n",
		inprogress.GetProjectDirName(),
	)
	if inprogress.PercentOf(len(inProgressSpecs), len(allSpecs)) == 0 && inprogress.PercentOf(len(inProgressScenarios), len(allScenarios)) == 0 {
		logger.Infof("No in progress scenarios found.")
	} else {
		logger.Infof(
			"Specifications: %d/%d (%0.0f%%)",
			len(inProgressSpecs),
			len(allSpecs),
			inprogress.PercentOf(len(inProgressSpecs), len(allSpecs)),
		)
		specsWithoutReason := len(inProgressSpecs) - len(inProgressSpecsWithReason)
		if specsWithoutReason > 0 {
			logger.Infof(
				"  - No inprogress comments: %d/%d (%0.0f%%)",
				specsWithoutReason,
				len(inProgressSpecs),
				inprogress.PercentOf(specsWithoutReason, len(inProgressSpecs)),
			)
		}
		logger.Infof(
			"Scenarios: %d/%d (%0.0f%%)",
			len(inProgressScenarios),
			len(allScenarios),
			inprogress.PercentOf(len(inProgressScenarios), len(allScenarios)),
		)
		scenariosWithoutReason := len(inProgressScenarios) - len(inProgressScenariosWithReason)
		if scenariosWithoutReason > 0 {
			logger.Infof(
				"  - No inprogress comments: %d/%d (%0.0f%%)",
				scenariosWithoutReason,
				len(inProgressScenarios),
				inprogress.PercentOf(scenariosWithoutReason, len(inProgressScenarios)),
			)
		}

	}
	logger.Infof("")
	logger.Infof("Successfully generated inprogress report to => %s\n", inprogress.GetReportPath())
	return &gauge_messages.Empty{}, nil
}

func (h *handler) Kill(c context.Context, m *gauge_messages.KillProcessRequest) (*gauge_messages.Empty, error) {
	defer h.stopServer()
	return &gauge_messages.Empty{}, nil
}

func (h *handler) stopServer() {
	h.server.Stop()
}
