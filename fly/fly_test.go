package fly_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/eugenmayer/concourse-pipeline-resource/fly"
	"github.com/eugenmayer/concourse-pipeline-resource/logger/loggerfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	errScript = `#!/bin/sh
>&1 echo "some std output"
>&2 echo "some err output"
exit 1
`
)

var _ = Describe("Command", func() {
	var (
		flyCommand fly.Command

		target   string
		teamName string

		tempDir         string
		flyBinaryPath   string
		fakeFlyContents string

		fakeLogger *loggerfakes.FakeLogger
	)

	BeforeEach(func() {
		target = "some-target"
		teamName = "main"

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		flyBinaryPath = filepath.Join(tempDir, "fake_fly")

		fakeFlyContents = `#!/bin/sh
		echo $@`

		fakeLogger = &loggerfakes.FakeLogger{}
	})

	JustBeforeEach(func() {
		err := ioutil.WriteFile(flyBinaryPath, []byte(fakeFlyContents), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		flyCommand = fly.NewCommand(target, fakeLogger, flyBinaryPath)
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Login", func() {
		var (
			url      string
			username string
			password string
			insecure bool
		)

		BeforeEach(func() {
			url = "some-url"
			username = "some-username"
			password = "some-password"
			insecure = false
		})

		It("returns output without error", func() {
			output, err := flyCommand.Login(url, teamName, username, password, insecure, false)
			Expect(err).NotTo(HaveOccurred())

			expectedOutput := fmt.Sprintf(
				"%s %s %s %s %s %s %s %s %s %s %s\n%s %s %s\n",
				"-t", target,
				"login",
				"-c", url,
				"-n", teamName,
				"-u", username,
				"-p", password,
				"-t", target,
				"sync",
			)

			Expect(string(output)).To(Equal(expectedOutput))
		})

		Context("when insecure is true", func() {
			BeforeEach(func() {
				insecure = true
			})

			It("adds -k flag to command", func() {
				output, err := flyCommand.Login(url, teamName, username, password, insecure, false)
				Expect(err).NotTo(HaveOccurred())

				expectedOutput := fmt.Sprintf(
					"%s %s %s %s %s %s %s %s %s %s %s %s\n%s %s %s\n",
					"-t", target,
					"login",
					"-c", url,
					"-n", teamName,
					"-u", username,
					"-p", password,
					"-k",
					"-t", target,
					"sync",
				)

				Expect(string(output)).To(Equal(expectedOutput))
			})
		})

		Context("when there is an error starting the commmand", func() {
			BeforeEach(func() {
				fakeFlyContents = ""
			})

			It("returns an error", func() {
				_, err := flyCommand.Login(url, teamName, username, password, insecure, false)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when no username or password is specified", func() {
			BeforeEach(func() {
				username = ""
				password = ""
			})

			It("does not pass the `p` or `u` flags to fly", func() {
				output, err := flyCommand.Login(url, teamName, username, password, insecure, false)
				Expect(err).NotTo(HaveOccurred())

				expectedOutput := fmt.Sprintf(
					"%s %s %s %s %s %s %s\n%s %s %s\n",
					"-t", target,
					"login",
					"-c", url,
					"-n", teamName,
					"-t", target,
					"sync",
				)

				Expect(string(output)).To(Equal(expectedOutput))
			})
		})

		Context("when the command returns an error", func() {
			BeforeEach(func() {
				fakeFlyContents = errScript
			})

			It("appends stderr to the error", func() {
				_, err := flyCommand.Login(url, teamName, username, password, insecure, false)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*some err output.*"))
			})
		})
	})

	Describe("Pipelines", func() {
		BeforeEach(func() {
			fakeFlyContents = `#!/bin/sh
echo '[{"name":"abc"},{"name":"def"}]'
`
		})

		It("returns pipelines without error", func() {
			pipelines, err := flyCommand.Pipelines()
			Expect(err).NotTo(HaveOccurred())

			Expect(pipelines).To(Equal([]string{"abc", "def"}))
		})
	})

	Describe("GetPipeline", func() {
		var (
			pipelineName string
		)

		BeforeEach(func() {
			pipelineName = "some-pipeline"
		})

		It("returns output without error", func() {
			output, err := flyCommand.GetPipeline(pipelineName)
			Expect(err).NotTo(HaveOccurred())

			expectedOutput := fmt.Sprintf(
				"%s %s %s %s %s\n",
				"-t", target,
				"get-pipeline",
				"-p", pipelineName,
			)

			Expect(string(output)).To(Equal(expectedOutput))
		})
	})

	Describe("SetPipeline", func() {
		var (
			pipelineName   string
			configFilepath string
		)

		BeforeEach(func() {
			pipelineName = "some-pipeline"
			configFilepath = "some-config-file"
		})

		It("returns output without error", func() {
			output, err := flyCommand.SetPipeline(pipelineName, configFilepath, nil)
			Expect(err).NotTo(HaveOccurred())

			expectedOutput := fmt.Sprintf(
				"%s %s %s %s %s %s %s %s\n",
				"-t", target,
				"set-pipeline",
				"-n",
				"-p", pipelineName,
				"-c", configFilepath,
			)

			Expect(string(output)).To(Equal(expectedOutput))
		})

		Context("when optional vars files are provided", func() {

			var (
				varsFiles []string
			)

			BeforeEach(func() {
				varsFiles = []string{
					"vars-file-1",
					"vars-file-2",
				}
			})

			It("returns output without error", func() {
				output, err := flyCommand.SetPipeline(pipelineName, configFilepath, varsFiles)
				Expect(err).NotTo(HaveOccurred())

				expectedOutput := fmt.Sprintf(
					"%s %s %s %s %s %s %s %s %s %s %s %s\n",
					"-t", target,
					"set-pipeline",
					"-n",
					"-p", pipelineName,
					"-c", configFilepath,
					"-l", varsFiles[0],
					"-l", varsFiles[1],
				)

				Expect(string(output)).To(Equal(expectedOutput))
			})
		})
	})

	Describe("DestroyPipeline", func() {
		var (
			pipelineName string
		)

		BeforeEach(func() {
			pipelineName = "some-pipeline"
		})

		It("returns output without error", func() {
			output, err := flyCommand.DestroyPipeline(pipelineName)
			Expect(err).NotTo(HaveOccurred())

			expectedOutput := fmt.Sprintf(
				"%s %s %s %s %s %s\n",
				"-t", target,
				"destroy-pipeline",
				"-n",
				"-p", pipelineName,
			)

			Expect(string(output)).To(Equal(expectedOutput))
		})
	})

	Describe("UnpausePipeline", func() {
		var (
			pipelineName string
		)

		BeforeEach(func() {
			pipelineName = "some-pipeline"
		})

		It("returns output without error", func() {
			output, err := flyCommand.UnpausePipeline(pipelineName)
			Expect(err).NotTo(HaveOccurred())

			expectedOutput := fmt.Sprintf(
				"%s %s %s %s %s\n",
				"-t", target,
				"unpause-pipeline",
				"-p", pipelineName,
			)

			Expect(string(output)).To(Equal(expectedOutput))
		})
	})
})
