package fly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"crypto/tls"
	"net/http"

	"github.com/eugenmayer/concourse-pipeline-resource/logger"
	"os"
	"errors"
	"strconv"
	"strings"
)

//go:generate counterfeiter . Command

type Command interface {
	Login(url string, teamName string, username string, password string, insecure bool, oauth bool) ([]byte, error)
	Pipelines() ([]string, error)
	GetPipeline(pipelineName string) ([]byte, error)
	SetPipeline(pipelineName string, configFilepath string, varsFilepaths []string) ([]byte, error)
	DestroyPipeline(pipelineName string) ([]byte, error)
	UnpausePipeline(pipelineName string) ([]byte, error)
}

type command struct {
	target        string
	logger        logger.Logger
	flyBinaryPath string
}

func NewCommand(target string, logger logger.Logger, flyBinaryPath string) Command {
	return &command{
		target:        target,
		logger:        logger,
		flyBinaryPath: flyBinaryPath,
	}
}

func (f command) Login(
	url string,
	teamName string,
	username string,
	password string,
	insecure bool,
	oauth bool,
) ([]byte, error) {
	if oauth {
		f.logger.Debugf("Using OAuth login")
		return f.loginOAuth(url, teamName, username, password, insecure)
	} else {
		f.logger.Debugf("Using legacy login")
		return f.loginLegacy(url, teamName, username, password, insecure, oauth)
	}
}

func (f command) loginLegacy(
	url string,
	teamName string,
	username string,
	password string,
	insecure bool,
	oauth bool,
) ([]byte, error) {
	args := []string{
		"login",
		"-c", url,
		"-n", teamName,
	}

	if username != "" && password != "" {
		args = append(args, "-u", username, "-p", password)
	}

	if insecure {
		args = append(args, "-k")
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
		}
		http.DefaultClient.Transport = tr
	}

	loginOut, err := f.run(args...)
	if err != nil {
		return nil, err
	}

	syncOut, err := f.run("sync")
	if err != nil {
		return nil, err
	}

	return append(loginOut, syncOut...), nil
}

func (f command) loginOAuth(
	url string,
	teamName string,
	username string,
	password string,
	insecure bool,
) ([]byte, error) {
	cmd := exec.Command("/opt/resource/oauth_login", url, teamName, username, password, strconv.FormatBool(insecure), f.target)
	cmd.Env = append(os.Environ(), "PATH=" + os.Getenv("PATH") + ":/opt/resource")
	f.logger.Debugf("\nCalling oauth logging: %s\n", strings.Join(cmd.Args, " "))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	loginError := cmd.Run()

	if loginError != nil {
		return nil, errors.New(loginError.Error() + "\n" + stdout.String() + "\n" + stderr.String())
	}

	return stdout.Bytes(), nil
}

func (f command) Pipelines() ([]string, error) {
	psOut, err := f.run("pipelines", "--json")
	if err != nil {
		return nil, err
	}

	var ps []struct {
		Name string `json:"name"`
	}

	err = json.Unmarshal(psOut, &ps)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(ps))
	for i, p := range ps {
		names[i] = p.Name
	}

	return names, nil
}

func (f command) GetPipeline(pipelineName string) ([]byte, error) {
	return f.run(
		"get-pipeline",
		"-p", pipelineName,
	)
}

func (f command) SetPipeline(
	pipelineName string,
	configFilepath string,
	varsFilepaths []string,
) ([]byte, error) {
	allArgs := []string{
		"set-pipeline",
		"-n",
		"-p", pipelineName,
		"-c", configFilepath,
	}

	for _, vf := range varsFilepaths {
		allArgs = append(allArgs, "-l", vf)
	}

	return f.run(allArgs...)
}

func (f command) UnpausePipeline(pipelineName string) ([]byte, error) {
	return f.run(
		"unpause-pipeline",
		"-p", pipelineName,
	)
}

func (f command) DestroyPipeline(pipelineName string) ([]byte, error) {
	return f.run(
		"destroy-pipeline",
		"-n",
		"-p", pipelineName,
	)
}

func (f command) run(args ...string) ([]byte, error) {
	if f.target == "" {
		return nil, fmt.Errorf("target cannot be empty in command.run")
	}

	defaultArgs := []string{
		"-t", f.target,
	}
	allArgs := append(defaultArgs, args...)
	cmd := exec.Command(f.flyBinaryPath, allArgs...)

	outbuf := bytes.NewBuffer(nil)
	errbuf := bytes.NewBuffer(nil)

	cmd.Stdout = outbuf
	cmd.Stderr = errbuf

	f.logger.Debugf("Starting fly command: %v\n", allArgs)
	err := cmd.Start()
	if err != nil {
		// If the command was never started, there will be nothing in the buffers
		return nil, err
	}

	f.logger.Debugf("Waiting for fly command: %v\n", allArgs)
	err = cmd.Wait()
	if err != nil {
		if len(errbuf.Bytes()) > 0 {
			err = fmt.Errorf("%v - %s", err, string(errbuf.Bytes()))
		}
		return outbuf.Bytes(), err
	}

	return outbuf.Bytes(), nil
}
