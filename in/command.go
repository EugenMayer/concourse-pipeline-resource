package in

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/eugenmayer/concourse-pipeline-resource/concourse"
	"github.com/eugenmayer/concourse-pipeline-resource/fly"
	"github.com/eugenmayer/concourse-pipeline-resource/logger"
)

const (
	apiPrefix = "/api/v1"
)

type Command struct {
	logger      logger.Logger
	flyCommand  fly.Command
	downloadDir string
}

func NewCommand(
	logger logger.Logger,
	flyCommand fly.Command,
	downloadDir string,
) *Command {
	return &Command{
		logger:      logger,
		flyCommand:  flyCommand,
		downloadDir: downloadDir,
	}
}

func (c *Command) Run(input concourse.InRequest) (concourse.InResponse, error) {
	c.logger.Debugf("Received input: %+v\n", input)

	insecure := false
	if input.Source.Insecure != "" {
		var err error
		insecure, err = strconv.ParseBool(input.Source.Insecure)
		if err != nil {
			return concourse.InResponse{}, err
		}
	}

	teams := make(map[string]concourse.Team)

	for _, team := range input.Source.Teams {
		teams[team.Name] = team
	}

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
			return concourse.InResponse{}, err
		}

		c.logger.Debugf("Login successful\n")

		pipelines, err := c.flyCommand.Pipelines()
		if err != nil {
			return concourse.InResponse{}, err
		}
		c.logger.Debugf("Found pipelines (%s): %+v\n", teamName, pipelines)

		var wg sync.WaitGroup
		wg.Add(len(pipelines))

		errChan := make(chan error, len(pipelines))

		for _, p := range pipelines {
			go func(pipelineName string) {
				defer wg.Done()

				outContents, err := c.flyCommand.GetPipeline(pipelineName)
				if err != nil {
					errChan <- err
				}
				pipelineContentsFilepath := filepath.Join(
					c.downloadDir,
					fmt.Sprintf(
						"%s-%s.yml",
						teamName,
						pipelineName,
					),
				)
				c.logger.Debugf(
					"Writing pipeline contents to: %s\n",
					pipelineContentsFilepath,
				)
				err = ioutil.WriteFile(pipelineContentsFilepath, outContents, os.ModePerm)
				// Untested as it is too hard to force ioutil.WriteFile to error
				if err != nil {
					errChan <- err
				}
			}(p)
		}

		c.logger.Debugf("Waiting for all pipelines\n")
		wg.Wait()
		c.logger.Debugf("Waiting for all pipelines complete\n")

		close(errChan)
		for err := range errChan {
			if err != nil {
				return concourse.InResponse{}, err
			}
		}
	}

	response := concourse.InResponse{
		Version:  input.Version,
		Metadata: []concourse.Metadata{},
	}

	return response, nil
}

type pipelineWithContent struct {
	name     string
	contents []byte
}
