package helper

import (
	"github.com/getgauge/common"
	"github.com/sitture/gauge-inprogress/gauge_messages"
	"github.com/sitture/gauge-inprogress/logger"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	EnvFileExtensions = "gauge_spec_file_extensions"
	EnvInProgressTags = "IN_PROGRESS_TAGS"
)

func GetProjectRoot() string {
	return os.Getenv(common.GaugeProjectRootEnv)
}

var GetProjectDirName = func() string {
	return path.Base(GetProjectRoot())
}

var GetInProgressTags = func() []string {
	e := os.Getenv(EnvInProgressTags)
	if e == "" {
		e = "wip, in-progress" //this was earlier hardcoded, this is a failsafe if env isn't set
	}
	tags := strings.Split(strings.TrimSpace(e), ",")
	var inProgressTags []string
	for _, ext := range tags {
		e := strings.TrimSpace(ext)
		if e != "" {
			inProgressTags = append(inProgressTags, e)
		}
	}
	return inProgressTags
}

// findFilesIn Finds all the files in the directory of a given extension
func findFilesIn(dirRoot string, isValidFile func(path string) bool, shouldSkip func(path string, f os.FileInfo) bool) []string {
	absRoot, _ := filepath.Abs(dirRoot)
	files := common.FindFilesInDir(absRoot, isValidFile, shouldSkip)
	return files
}

func getProtoSpecs(m *gauge_messages.SpecDetails) []*gauge_messages.ProtoSpec {
	specs := make([]*gauge_messages.ProtoSpec, 0)
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

var GaugeSpecFileExtensions = func() []string {
	e := os.Getenv(EnvFileExtensions)
	if e == "" {
		e = ".spec, .md" //this was earlier hardcoded, this is a failsafe if env isn't set
	}
	exts := strings.Split(strings.TrimSpace(e), ",")
	var allowedExts = []string{}
	for _, ext := range exts {
		e := strings.TrimSpace(ext)
		if e != "" {
			allowedExts = append(allowedExts, e)
		}
	}
	return allowedExts
}

// IsValidSpecExtension Checks if the path has a spec file extension
func IsValidSpecExtension(path string) bool {
	for _, ext := range GaugeSpecFileExtensions() {
		if ext == strings.ToLower(filepath.Ext(path)) {
			return true
		}
	}
	return false
}

// FindSpecFilesIn Finds spec files in the given directory
var FindSpecFilesIn = func(dir string) []string {
	return findFilesIn(dir, IsValidSpecExtension, func(path string, f os.FileInfo) bool {
		return false
	})
}

// GetSpecFiles returns the list of spec files present at the given path.
// If the path itself represents a spec file, it returns the same.
var GetSpecFiles = func(paths []string) []string {
	var specFiles []string
	for _, path := range paths {
		if !common.FileExists(path) {
			logger.Fatalf("Specs directory %s does not exists.", path)
		}
		if common.DirExists(path) {
			specFilesInPath := FindSpecFilesIn(path)
			if len(specFilesInPath) < 1 {
				logger.Fatalf("No specifications found in %s.", path)
			}
			specFiles = append(specFiles, specFilesInPath...)
		} else if IsValidSpecExtension(path) {
			f, _ := filepath.Abs(path)
			specFiles = append(specFiles, f)
		}
	}
	return specFiles
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
		if containsInProgressTags(spec.GetTags()) {
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
				if scenario != nil && containsInProgressTags(scenario.GetTags()) {
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

func containsInProgressTags(tags []string) bool {
	for _, inProgressTag := range GetInProgressTags() {
		for _, tag := range tags {
			if strings.TrimSpace(tag) == inProgressTag {
				return true
			}
		}
	}
	return false
}

func PercentOf(part int, total int) float64 {
	return (float64(part) * float64(100)) / float64(total)
}
