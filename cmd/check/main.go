package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/eugenmayer/concourse-pipeline-resource/check"
	"github.com/eugenmayer/concourse-pipeline-resource/concourse"
	"github.com/eugenmayer/concourse-pipeline-resource/fly"
	"github.com/eugenmayer/concourse-pipeline-resource/logger"
	"github.com/eugenmayer/concourse-pipeline-resource/validator"
	"github.com/robdimsdale/sanitizer"
)

const (
	flyBinaryName        = "fly"
	atcExternalURLEnvKey = "ATC_EXTERNAL_URL"
)

var (
	l logger.Logger
)

func main() {
	checkDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln(err)
	}

	var input concourse.CheckRequest

	logFile, err := ioutil.TempFile("", "concourse-pipeline-resource-check.log")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Fprintf(os.Stderr, "Running check")
	fmt.Fprintf(os.Stderr, "Logging to %s\n", logFile.Name())

	err = json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		fmt.Fprintf(logFile, "Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	sanitized := concourse.SanitizedSource(input.Source)
	sanitizer := sanitizer.NewSanitizer(sanitized, logFile)

	l = logger.NewLogger(sanitizer)

	flyBinaryPath := filepath.Join(checkDir, flyBinaryName)

	if input.Source.Target == "" {
		input.Source.Target = os.Getenv(atcExternalURLEnvKey)
	}

	flyCommand := fly.NewCommand(input.Source.Target, l, flyBinaryPath)

	err = validator.ValidateCheck(input)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	command := check.NewCommand(l, logFile.Name(), flyCommand)
	response, err := command.Run(input)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	err = json.NewEncoder(os.Stdout).Encode(response)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}
}
