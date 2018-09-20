package validator

import (
	"fmt"

	"github.com/eugenmayer/concourse-pipeline-resource/concourse"
)

func ValidateTeams(teams []concourse.Team) error {
	if teams == nil || len(teams) == 0 {
		return fmt.Errorf("%s must be provided in source", "teams")
	}

	for i, team := range teams {
		if team.Name == "" {
			return fmt.Errorf("%s must be provided for team: %d", "name", i)
		}

		if team.Username == "" && team.Password != "" {
			return fmt.Errorf("%s must be provided for team: %s", "username", team.Name)
		}

		if team.Password == "" && team.Username != "" {
			return fmt.Errorf("%s must be provided for team: %s", "password", team.Name)
		}
	}

	return nil
}
