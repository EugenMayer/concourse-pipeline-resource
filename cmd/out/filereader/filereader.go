package filereader

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/eugenmayer/concourse-pipeline-resource/concourse"
	"gopkg.in/yaml.v2"
)

func PipelinesFromFile(pipelinesFilename string, sourcesDir string) ([]concourse.Pipeline, error) {
	if pipelinesFilename != "" {
		if sourcesDir == "" {
			return nil, fmt.Errorf("sourcesDir must be non-empty")
		}

		b, err := ioutil.ReadFile(filepath.Join(sourcesDir, pipelinesFilename))
		if err != nil {
			return nil, err
		}

		var fileContents concourse.OutParams
		err = yaml.Unmarshal(b, &fileContents)
		if err != nil {
			return nil, err
		}

		return fileContents.Pipelines, nil
	}

	return []concourse.Pipeline{}, nil
}
