package validator

import (
	"fmt"

	"github.com/eugenmayer/concourse-pipeline-resource/concourse"
)

func ValidateOut(input concourse.OutRequest) error {
	err := ValidateTeams(input.Source.Teams)
	if err != nil {
		return err
	}

	sourceTeamNames := []string{}
	for _, team := range input.Source.Teams {
		sourceTeamNames = append(sourceTeamNames, team.Name)
	}

	if input.Source.Target == "" {
		return fmt.Errorf("%s must be provided in source", "target")
	}

	var pipelinesFilePresent bool
	var pipelinesPresent bool

	if input.Params.PipelinesFile != "" {
		pipelinesFilePresent = true
	}

	if input.Params.Pipelines != nil && len(input.Params.Pipelines) > 0 {
		pipelinesPresent = true
	}

	if !(pipelinesPresent || pipelinesFilePresent) {
		return fmt.Errorf(
			"pipelines must be provided via either %s or %s",
			"pipelines",
			"pipelines_file",
		)
	}

	if pipelinesPresent && pipelinesFilePresent {
		return fmt.Errorf(
			"pipelines must be provided via one of either %s or %s",
			"pipelines",
			"pipelines_file",
		)
	}

	for i, p := range input.Params.Pipelines {
		if p.Name == "" {
			return fmt.Errorf("%s must be provided for pipeline[%d]", "name", i)
		}

		if p.ConfigFile == "" {
			return fmt.Errorf("%s must be provided for pipeline[%d]", "config_file", i)
		}

		if p.TeamName == "" {
			return fmt.Errorf("%s must be provided for pipeline[%d]", "team", i)
		}

		if !stringContains(sourceTeamNames, p.TeamName) {
			return fmt.Errorf("team name '%s' not found in source team names: %v", p.TeamName, sourceTeamNames)
		}

		// vars files can be nil as it is optional.
		if p.VarsFiles != nil {
			// However, if it is provided it must be non-empty
			if len(p.VarsFiles) == 0 {
				return fmt.Errorf("%s must be non-empty if provided for pipeline[%d]", "vars_files", i)
			}

			for j, v := range p.VarsFiles {
				if len(v) == 0 {
					return fmt.Errorf(
						"%s must be non-empty for pipeline[%d].vars_files[%d]",
						"vars file",
						i,
						j,
					)
				}
			}
		}
	}

	return nil
}

func stringContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}

	return false
}
