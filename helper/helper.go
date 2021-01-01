package helper

import "github.com/getgauge/common"

func GetProjectRoot() string {
	projectRoot, _ := common.GetProjectRoot()
	return projectRoot
}
