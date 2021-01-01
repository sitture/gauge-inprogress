package main

import (
	"github.com/sitture/gauge-inprogress/helper"
	"os"
)

func main() {
	os.Chdir(helper.GetProjectRoot())
}