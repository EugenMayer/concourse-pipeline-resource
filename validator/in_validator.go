package validator

import (
	"fmt"

	"github.com/eugenmayer/concourse-pipeline-resource/concourse"
)

func ValidateIn(input concourse.InRequest) error {
	if input.Source.Target == "" {
		return fmt.Errorf("%s must be provided in source", "target")
	}

	return ValidateTeams(input.Source.Teams)
}
