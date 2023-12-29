package radiclegithubactions

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"radicle-github-actions-adapter/app"
	"strings"
)

// listYAMLFiles lists all .yaml and .yml files from the given directory.
// It does not look into subdirectories.
func (rga *RadicleGitHubActions) listYAMLFiles(directory string) ([]string, error) {
	var yamlFiles []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			rga.logger.Debug("directory not found", "path", path)
			return err
		}
		if info.IsDir() {
			rga.logger.Debug("file is a directory", "path", path)
			return filepath.SkipDir
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), ".yaml") || strings.HasSuffix(strings.ToLower(info.Name()), ".yml") {
			yamlFiles = append(yamlFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return yamlFiles, nil
}

// getRadicleGitHubActionsSetup retrieves the GitHub Actions settings of the Radicle project.
// If no file found it returns an error.
func (rga *RadicleGitHubActions) getRadicleGitHubActionsSetup(filePath string) (*app.GitHubActionsSettings, error) {
	gitHubActionsSettings := app.GitHubActionsSettings{}
	yamlFileContent, err := os.ReadFile(filePath)
	if err != nil {
		rga.logger.Info("no Radicle GiHub Actions settings file found", "file", filePath)
		return nil, err
	}
	decoder := yaml.NewDecoder(strings.NewReader(string(yamlFileContent)))
	err = decoder.Decode(&gitHubActionsSettings)
	if err != nil {
		rga.logger.Info("could not decode Radicle GiHub Actions settings file", "error", err)
		return nil, err
	}
	return &gitHubActionsSettings, nil
}
