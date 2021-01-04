package helper

import (
	"fmt"
	"github.com/getgauge/common"
	"github.com/sitture/gauge-inprogress/gauge_messages"
	"github.com/sitture/gauge-inprogress/logger"
	"os"
	"path"
	"path/filepath"
)

func GetProjectRoot() string {
	return os.Getenv(common.GaugeProjectRootEnv)
}

func getAbsPath(path string) string {
	f, err := filepath.Abs(path)
	if err != nil {
		logger.Fatalf("Cannot get absolute path")
	}
	return f
}

func GetProjectSpecsDir() string {
	return getAbsPath(os.Getenv("gauge_specs_dir"))
}

func init() {
	AcceptedExtensions[".spec"] = true
	AcceptedExtensions[".md"] = true
}

var AcceptedExtensions = make(map[string]bool)
var GetProjectDirName = func() string {
	return path.Base(GetProjectRoot())
}

func isValidSpecExtension(path string) bool {
	return AcceptedExtensions[filepath.Ext(path)]
}

func findFilesInDir(dirPath string, isValidFile func(path string) bool) []string {
	var files []string
	filepath.Walk(dirPath, func(path string, f os.FileInfo, err error) error {
		if err == nil && !f.IsDir() && isValidFile(path) {
			files = append(files, path)
		}
		return err
	})
	return files
}

func findFilesIn(dirRoot string, isValidFile func(path string) bool) []string {
	absRoot := getAbsPath(dirRoot)
	files := findFilesInDir(absRoot, isValidFile)
	return files
}

func dirExists(dirPath string) bool {
	stat, err := os.Stat(dirPath)
	if err == nil && stat.IsDir() {
		return true
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

func GetFiles(path string) []string {
	var specFiles []string
	if dirExists(path) {
		specFiles = append(specFiles, findFilesIn(path, isValidSpecExtension)...)
	} else if fileExists(path) && isValidSpecExtension(path) {
		specFiles = append(specFiles, getAbsPath(path))
	}
	return specFiles
}

func getProtoSpecs(m *gauge_messages.SpecDetails) []*gauge_messages.ProtoSpec {
	specs := make([]*gauge_messages.ProtoSpec, 0)
	fmt.Println(len(m.GetDetails()))
	for _, detail := range m.GetDetails() {
		if detail.GetSpec() != nil {
			specs = append(specs, detail.GetSpec())
		}
	}
	return specs
}

func GetSpecs(m *gauge_messages.SpecDetails, files []string) []*gauge_messages.ProtoSpec {
	specs := getProtoSpecs(m)
	sortedSpecs := make([]*gauge_messages.ProtoSpec, 0)
	specsMap := make(map[string]*gauge_messages.ProtoSpec, 0)
	for _, spec := range specs {
		specsMap[spec.GetFileName()] = spec
	}
	for _, file := range files {
		spec, ok := specsMap[file]
		if !ok {
			spec = &gauge_messages.ProtoSpec{FileName: file, Tags: make([]string, 0), Items: make([]*gauge_messages.ProtoItem, 0)}
		}
		sortedSpecs = append(sortedSpecs, spec)
	}
	return sortedSpecs
}

func GetScenarios(specs []*gauge_messages.ProtoSpec) []*gauge_messages.ProtoScenario {
	scenarios := make([]*gauge_messages.ProtoScenario, 0)
	for _, spec := range specs {
		for _, itemType := range spec.GetItems() {
			if itemType.GetScenario() != nil {
				scenarios = append(scenarios, itemType.GetScenario())
			}
		}
	}
	return scenarios
}

func GetInProgressSpecsScenarios(specs []*gauge_messages.ProtoSpec) (
	inProgressSpecs map[string]*gauge_messages.ProtoSpec,
	inProgressScenarios map[string]*gauge_messages.ProtoScenario,
) {
	inProgressSpecs = make(map[string]*gauge_messages.ProtoSpec, 0)
	inProgressScenarios = make(map[string]*gauge_messages.ProtoScenario, 0)
	for _, spec := range specs {
		if find(spec.GetTags(), "wip") {
			if _, exists := inProgressSpecs[spec.GetFileName()]; !exists {
				inProgressSpecs[spec.GetFileName()] = spec
			}
			for _, itemType := range spec.GetItems() {
				scenario := itemType.GetScenario()
				if scenario != nil {
					inProgressScenarios[scenario.GetScenarioHeading()] = scenario
				}
			}
		} else {
			for _, itemType := range spec.GetItems() {
				scenario := itemType.GetScenario()
				if scenario != nil && find(scenario.GetTags(), "wip") {
					inProgressScenarios[scenario.GetScenarioHeading()] = scenario
					if _, exists := inProgressSpecs[spec.GetFileName()]; !exists {
						inProgressSpecs[spec.GetFileName()] = spec
					}
				}

			}
		}
	}
	return
}

func find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
