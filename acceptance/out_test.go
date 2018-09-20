package acceptance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/eugenmayer/concourse-pipeline-resource/concourse"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const (
	outTimeout = 60 * time.Second

	defaultPipelinesFileFilename = "pipelines.yml"
)

var _ = Describe("Out", func() {
	var (
		command       *exec.Cmd
		outRequest    concourse.OutRequest
		stdinContents []byte
		sourcesDir    string

		pipelineName           string
		pipelineConfig         string
		pipelineConfigFilename string
		pipelineConfigFilepath string

		varsFileContents string
		varsFileFilename string
		varsFileFilepath string

		pipelinesFileContentsBytes []byte
		pipelinesFileFilename      string
		pipelinesFileFilepath      string

		pipelines []concourse.Pipeline
	)

	BeforeEach(func() {
		var err error
		By("Creating temp directory")
		sourcesDir, err = ioutil.TempDir("", "concourse-pipeline-resource")
		Expect(err).NotTo(HaveOccurred())

		By("Creating random pipeline name")
		pipelineName = fmt.Sprintf("cp-resource-test-%d", time.Now().UnixNano())

		By("Writing pipeline config file")
		pipelineConfig = `---
resources:
- name: concourse-pipeline-resource-repo
  type: git
  source:
    uri: https://github.com/eugenmayer/concourse-pipeline-resource.git
    branch: {{foo}}
jobs:
- name: get-concourse-pipeline-resource-repo
  plan:
  - get: concourse-pipeline-resource-repo
`

		pipelineConfigFilename = fmt.Sprintf("%s.yml", pipelineName)
		pipelineConfigFilepath = filepath.Join(sourcesDir, pipelineConfigFilename)
		err = ioutil.WriteFile(pipelineConfigFilepath, []byte(pipelineConfig), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		By("Writing vars file")
		varsFileContents = "foo: bar"

		varsFileFilename = fmt.Sprintf("%s_vars.yml", pipelineName)
		varsFileFilepath = filepath.Join(sourcesDir, varsFileFilename)
		err = ioutil.WriteFile(varsFileFilepath, []byte(varsFileContents), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		By("Creating command object")
		command = exec.Command(outPath, sourcesDir)

		By("Creating pipeline input")
		pipelines = []concourse.Pipeline{
			{
				Name:       pipelineName,
				TeamName:   teamName,
				ConfigFile: pipelineConfigFilename,
				VarsFiles: []string{
					varsFileFilename,
				},
			},
		}

		pipelinesFileContents := concourse.OutParams{
			Pipelines: pipelines,
		}

		pipelinesFileContentsBytes, err = yaml.Marshal(pipelinesFileContents)
		Expect(err).NotTo(HaveOccurred())

		By("Writing pipelines file")
		pipelinesFileFilename = defaultPipelinesFileFilename
		pipelinesFileFilepath = filepath.Join(sourcesDir, pipelinesFileFilename)
		err = ioutil.WriteFile(pipelinesFileFilepath, pipelinesFileContentsBytes, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		// Default test case uses static config so set the file name to empty
		By("Setting pipelinesFileFilename to empty")
		pipelinesFileFilename = ""
	})

	JustBeforeEach(func() {
		By("Creating default request")
		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				Target:   target,
				Insecure: fmt.Sprintf("%t", insecure),
				Teams: []concourse.Team{
					{
						Name:     teamName,
						Username: username,
						Password: password,
					},
				},
			},
			Params: concourse.OutParams{
				Pipelines:     pipelines,
				PipelinesFile: pipelinesFileFilename,
			},
		}

		var err error
		stdinContents, err = json.Marshal(outRequest)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("Creating pipelines successfully", func() {
		AfterEach(func() {
			_, err := flyCommand.DestroyPipeline(pipelineName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("creates pipeline and returns valid json", func() {
			By("Running the command")
			session := run(command, stdinContents)
			Eventually(session, outTimeout).Should(gexec.Exit(0))

			By("Outputting a valid json response")
			response := concourse.OutResponse{}
			err := json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).ShouldNot(HaveOccurred())

			By("Validating output contains pipeline version")
			Expect(response.Version[pipelineName]).NotTo(BeEmpty())
		})

		Context("when pipelines_file is provided instead", func() {
			BeforeEach(func() {
				pipelines = []concourse.Pipeline{}
				pipelinesFileFilename = defaultPipelinesFileFilename
			})

			It("creates pipeline and returns valid json", func() {
				By("Running the command")
				session := run(command, stdinContents)
				Eventually(session, outTimeout).Should(gexec.Exit(0))

				By("Outputting a valid json response")
				response := concourse.OutResponse{}
				err := json.Unmarshal(session.Out.Contents(), &response)
				Expect(err).ShouldNot(HaveOccurred())

				By("Validating output contains pipeline version")
				Expect(response.Version[pipelineName]).NotTo(BeEmpty())
			})
		})

		Context("when the target is not provided", func() {
			BeforeEach(func() {
				var err error
				err = os.Setenv("ATC_EXTERNAL_URL", outRequest.Source.Target)
				Expect(err).ShouldNot(HaveOccurred())

				outRequest.Source.Target = ""

				stdinContents, err = json.Marshal(outRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("creates pipeline and returns valid json", func() {
				By("Running the command")
				session := run(command, stdinContents)
				Eventually(session, outTimeout).Should(gexec.Exit(0))

				By("Outputting a valid json response")
				response := concourse.OutResponse{}
				err := json.Unmarshal(session.Out.Contents(), &response)
				Expect(err).ShouldNot(HaveOccurred())

				By("Validating output contains pipeline version")
				Expect(response.Version[pipelineName]).NotTo(BeEmpty())
			})
		})
	})

	Context("when validation fails", func() {
		BeforeEach(func() {
			pipelines = []concourse.Pipeline{}
			pipelinesFileFilename = ""
		})

		It("exits with error", func() {
			By("Running the command")
			session := run(command, stdinContents)

			By("Validating command exited with error")
			Eventually(session, outTimeout).Should(gexec.Exit(1))
			Expect(session.Err).Should(gbytes.Say(".*pipelines.*provided"))
		})
	})
})
