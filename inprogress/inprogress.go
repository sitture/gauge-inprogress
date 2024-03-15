package inprogress

import (
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/sitture/gauge-inprogress/logger"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	EnvFileExtensions       = "gauge_spec_file_extensions"
	EnvGaugeSpecDirs        = "GAUGE_SPEC_DIRS"
	EnvInProgressTags       = "IN_PROGRESS_TAGS"
	EnvInProgressConsoleOut = "IN_PROGRESS_CONSOLE_OUTPUT"
	reportDirectory         = "reports"
	reportFile              = "inprogress.md"
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

var GetSpecDirs = func() []string {
	return strings.Split(strings.TrimSpace(os.Getenv(EnvGaugeSpecDirs)), "||")
}

var OutPutScenariosToConsole = func() bool {
	value := os.Getenv(EnvInProgressConsoleOut)
	if value == "true" {
		return true
	}
	return false
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

type InProgressSpec struct {
	spec      *gauge_messages.ProtoSpec
	scenarios []*gauge_messages.ProtoScenario
}

func (inProgressSpec *InProgressSpec) GetSpec() *gauge_messages.ProtoSpec {
	return inProgressSpec.spec
}

func (inProgressSpec *InProgressSpec) GetScenarios() []*gauge_messages.ProtoScenario {
	return inProgressSpec.scenarios
}

func getInProgressScenariosBySpec(spec *gauge_messages.ProtoSpec, tags []string) []*gauge_messages.ProtoScenario {
	inProgressScenarios := make([]*gauge_messages.ProtoScenario, 0)
	if containsInProgressTags(tags) {
		for _, itemType := range spec.GetItems() {
			scenario := itemType.GetScenario()
			if scenario != nil {
				inProgressScenarios = append(inProgressScenarios, scenario)
			}
		}
	}
	return inProgressScenarios
}

func GetInProgressScenarios(specs map[string]InProgressSpec) []*gauge_messages.ProtoScenario {
	inProgressScenarios := make([]*gauge_messages.ProtoScenario, 0)
	for _, spec := range specs {
		inProgressScenarios = append(inProgressScenarios, spec.GetScenarios()...)
	}
	return inProgressScenarios
}

func GetInProgressSpecs(specs []*gauge_messages.ProtoSpec) map[string]InProgressSpec {
	inProgressSpecs := make(map[string]InProgressSpec, 0)
	for _, spec := range specs {
		specKey := spec.GetSpecHeading()
		if containsInProgressTags(spec.GetTags()) {
			if _, exists := inProgressSpecs[specKey]; !exists {
				inProgressSpec := InProgressSpec{spec: spec, scenarios: getInProgressScenariosBySpec(spec, spec.GetTags())}
				inProgressSpecs[specKey] = inProgressSpec
			}
		} else {
			inProgressScenarios := make([]*gauge_messages.ProtoScenario, 0)
			if inProgressSpec, exists := inProgressSpecs[specKey]; exists {
				inProgressScenarios = append(inProgressScenarios, inProgressSpec.scenarios...)
			}
			for _, itemType := range spec.GetItems() {
				scenario := itemType.GetScenario()
				if scenario != nil && containsInProgressTags(scenario.GetTags()) {
					inProgressScenarios = append(inProgressScenarios, scenario)
				}
			}
			if len(inProgressScenarios) > 0 {
				delete(inProgressSpecs, specKey)
				inProgressSpec := InProgressSpec{spec: spec, scenarios: inProgressScenarios}
				inProgressSpecs[specKey] = inProgressSpec
			}
		}
	}
	return inProgressSpecs
}

func GetInProgressSpecsWithReason(specs map[string]InProgressSpec) map[string]InProgressSpec {
	inProgressSpecs := make(map[string]InProgressSpec, 0)
	for specKey, spec := range specs {
		for _, specItem := range spec.GetSpec().GetItems() {
			if specItem.GetItemType() == gauge_messages.ProtoItem_Comment && containsInProgressPrefix(specItem.GetComment().GetText()) {
				if _, exists := inProgressSpecs[specKey]; !exists {
					inProgressSpec := InProgressSpec{spec: spec.GetSpec(), scenarios: spec.GetScenarios()}
					inProgressSpecs[specKey] = inProgressSpec
				}
			}
		}
		for _, scenario := range spec.GetScenarios() {
			for _, scenItem := range scenario.GetScenarioItems() {
				if scenItem.GetItemType() == gauge_messages.ProtoItem_Comment && containsInProgressPrefix(scenItem.GetComment().GetText()) {
					if _, exists := inProgressSpecs[specKey]; !exists {
						inProgressSpec := InProgressSpec{spec: spec.GetSpec(), scenarios: spec.GetScenarios()}
						inProgressSpecs[specKey] = inProgressSpec
					}
				}
			}
		}
	}
	return inProgressSpecs
}

type ScenarioWithReason struct {
	Scenario *gauge_messages.ProtoScenario
	Reason   string
}

func getScenarioMapKey(spec *gauge_messages.ProtoSpec, scenario *gauge_messages.ProtoScenario) string {
	return fmt.Sprintf("%s_%s", spec.GetFileName(), scenario.GetScenarioHeading())
}

func GetInProgressScenariosWithReason(specs map[string]InProgressSpec) map[string]ScenarioWithReason {
	inProgressScenarios := make(map[string]ScenarioWithReason, 0)
	for _, spec := range specs {
		if containsInProgressTags(spec.GetSpec().GetTags()) {
			for _, specItem := range spec.GetSpec().GetItems() {
				if specItem.GetItemType() == gauge_messages.ProtoItem_Comment && containsInProgressPrefix(specItem.GetComment().GetText()) {
					for _, scenario := range spec.GetScenarios() {
						key := getScenarioMapKey(spec.GetSpec(), scenario)
						inProgressScenarios[key] = ScenarioWithReason{scenario, specItem.GetComment().GetText()}
					}
				}
			}
		} else {
			for _, scenario := range spec.GetScenarios() {
				key := getScenarioMapKey(spec.GetSpec(), scenario)
				for _, scenItem := range scenario.GetScenarioItems() {
					if scenItem.GetItemType() == gauge_messages.ProtoItem_Comment && containsInProgressPrefix(scenItem.GetComment().GetText()) {
						inProgressScenarios[key] = ScenarioWithReason{scenario, scenItem.GetComment().GetText()}
					}
				}
			}
		}
	}
	return inProgressScenarios
}

var GetReportPath = func() string {
	return path.Join(GetProjectRoot(), reportDirectory, reportFile)
}

func WriteToFile(inProgressSpecs map[string]InProgressSpec, inProgressScenariosWithReason map[string]ScenarioWithReason) (error error) {
	console := OutPutScenariosToConsole()
	file, err := os.Create(GetReportPath())
	if err != nil {
		logger.Fatalf("Unable to create data.js file")
	}
	for _, spec := range inProgressSpecs {
		specLine := fmt.Sprintf("# %s // %s", spec.GetSpec().GetSpecHeading(),
			filepath.Base(spec.GetSpec().GetFileName()))
		_, error = file.WriteString(specLine + "\n")
		if console {
			logger.Infof(specLine)
		}
		for _, scenario := range spec.GetScenarios() {
			scenarioLine := fmt.Sprintf("  ## %s", scenario.GetScenarioHeading())
			_, error = file.WriteString(scenarioLine + "\n")
			if console {
				logger.Infof(scenarioLine)
			}
			inProgressReason := inProgressScenariosWithReason[getScenarioMapKey(spec.GetSpec(), scenario)].Reason
			if len(strings.TrimSpace(inProgressReason)) > 0 {
				reasonLine := fmt.Sprintf("    - %s", inProgressReason)
				_, error = file.WriteString(reasonLine + "\n")
				if console {
					logger.Infof(reasonLine)
				}
			}
		}
	}
	error = file.Close()
	return
}

func containsInProgressPrefix(comment string) bool {
	var inProgressRegex = regexp.MustCompile(`^in.?progress|//.?in.?progress|^wip|^//.?wip`)
	return inProgressRegex.MatchString(strings.ToLower(strings.TrimSpace(comment)))
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
