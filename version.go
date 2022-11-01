package main

import (
	"fmt"
	"github.com/common-nighthawk/go-figure"
)

var (
	Version      string
	GoVersion    string
	GitUrl       string
	GitBranch    string
	GitCommit    string
	GitLatestTag string
	BuildTime    string
	Title        = figure.NewFigure("Otter Alter", "doom", true).String()
)

func VersionIfo() string {
	return fmt.Sprintf("\nVersion: %s\nGoVersion: %s\nGitUrl: %s\nGitBranch: %s\nGitCommit: %s\nGitLatestTag: %s\nBuildTime: %s",
		Version, GoVersion, GitUrl, GitBranch, GitCommit, GitLatestTag, BuildTime)
}
