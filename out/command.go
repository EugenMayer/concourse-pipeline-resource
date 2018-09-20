package out

import (
	"crypto/md5"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/eugenmayer/concourse-pipeline-resource/concourse"
	"github.com/eugenmayer/concourse-pipeline-resource/fly"
	"github.com/eugenmayer/concourse-pipeline-resource/logger"
)

const (
	apiPrefix = "/api/v1"
)

type Command struct {
	logger     logger.Logger
	flyCommand fly.Command
	sourcesDir string
}

func NewCommand(
	logger logger.Logger,
	flyCommand fly.Command,
	sourcesDir string,
) *Command {
	return &Command{
		logger:     logger,
		flyCommand: flyCommand,
		sourcesDir: sourcesDir,
	}
}

func (c *Command) Run(input concourse.OutRequest) (concourse.OutResponse, error) {
	c.logger.Debugf("Received input: %+v\n", input)

	insecure := false
	if input.Source.Insecure != "" {
		var err error
		insecure, err = strconv.ParseBool(input.Source.Insecure)
		if err != nil {
			return concourse.OutResponse{}, err
		}
	}

	teams := make(map[string]concourse.Team)

	for _, team := range input.Source.Teams {
		teams[team.Name] = team
	}

	pipelines := input.Params.Pipelines

	c.logger.Debugf("Input pipelines: %+v\n", pipelines)

	c.logger.Debugf("Setting pipelines\n")
	for _, p := range pipelines {
		team, found := teams[p.TeamName]
		if !found {
			return concourse.OutResponse{}, fmt.Errorf("team (%s) configuration not found for pipeline (%s)", p.TeamName, p.Name)
		}

		c.logger.Debugf("Performing login\n")

		_, err:= c.flyCommand.Login(input.Source.Target,p.TeamName,team.Username,team.Password, insecure, team.OAuth)
		if err != nil {
			return concourse.OutResponse{}, err
		}

		c.logger.Debugf("Login successful\n")

		configFilepath := filepath.Join(c.sourcesDir, p.ConfigFile)

		var varsFilepaths []string
		for _, v := range p.VarsFiles {
			varFilepath := filepath.Join(c.sourcesDir, v)
			varsFilepaths = append(varsFilepaths, varFilepath)
		}

		_, err = c.flyCommand.SetPipeline(p.Name, configFilepath, varsFilepaths)
		if err != nil {
			return concourse.OutResponse{}, err
		}

		if p.Unpaused {
			_, err = c.flyCommand.UnpausePipeline(p.Name)
			if err != nil {
				return concourse.OutResponse{}, err
			}
		}
	}
	c.logger.Debugf("Setting pipelines complete\n")

	pipelineVersions := make(map[string]string)

	for teamName, team := range teams {
		c.logger.Debugf("Performing login\n")
		_, err := c.flyCommand.Login(
			input.Source.Target,
			teamName,
			team.Username,
			team.Password,
			insecure,
			team.OAuth,
		)
		if err != nil {
			return concourse.OutResponse{}, err
		}

		c.logger.Debugf("Login successful\n")

		pipelines, err := c.flyCommand.Pipelines()
		if err != nil {
			return concourse.OutResponse{}, err
		}
		c.logger.Debugf("Found pipelines (%s): %+v\n", teamName, pipelines)

		for _, pipelineName := range pipelines {
			c.logger.Debugf("Getting pipeline: %s\n", pipelineName)
			outBytes, err := c.flyCommand.GetPipeline(pipelineName)
			if err != nil {
				return concourse.OutResponse{}, err
			}

			version := fmt.Sprintf(
				"%x",
				md5.Sum(outBytes),
			)
			pipelineVersions[pipelineName] = version
		}
	}

	response := concourse.OutResponse{
		Version:  pipelineVersions,
		Metadata: []concourse.Metadata{},
	}

	return response, nil
}
