package config

import (
	"fmt"
	"runtime"
)

var (
	gitVersion = "dev"
	gitCommit  = ""
	buildDate  = ""
)

type BuildInfo struct {
	GitVersion string `json:"git_version"`
	GitCommit  string `json:"git_commit"`
	BuildDate  string `json:"build_date"`
	GoVersion  string `json:"go_version"`
	Compiler   string `json:"compiler"`
	Platform   string `json:"platform"`
}

func GetBuildInfo() *BuildInfo {
	return &BuildInfo{
		GitVersion: gitVersion,
		GitCommit:  gitCommit,
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		Compiler:   runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
