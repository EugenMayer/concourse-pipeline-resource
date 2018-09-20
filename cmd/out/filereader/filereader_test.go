package filereader_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/eugenmayer/concourse-pipeline-resource/cmd/out/filereader"
	"github.com/eugenmayer/concourse-pipeline-resource/concourse"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var _ = Describe("Filereader", func() {
	var (
		sourcesDir        string
		pipelinesFilename string

		pipelines []concourse.Pipeline
	)

	BeforeEach(func() {
		var err error
		sourcesDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		pipelinesFilename = "pipelines.yml"

		pipelines = []concourse.Pipeline{
			{
				Name:       "name-1",
				ConfigFile: "pipeline_1.yml",
				VarsFiles: []string{
					"vars_1.yml",
					"vars_2.yml",
				},
			},
			{
				Name:       "name-2",
				ConfigFile: "pipeline_2.yml",
				VarsFiles:  []string{},
			},
		}

		pipelinesFileContents := concourse.OutParams{
			Pipelines: pipelines,
		}

		pipelinesFileContentsBytes, err := yaml.Marshal(pipelinesFileContents)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(
			filepath.Join(sourcesDir, pipelinesFilename),
			pipelinesFileContentsBytes,
			os.ModePerm,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(sourcesDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("parses pipelines from the file", func() {
		returnedPipelines, err := filereader.PipelinesFromFile(pipelinesFilename, sourcesDir)
		Expect(err).NotTo(HaveOccurred())

		Expect(returnedPipelines).To(Equal(pipelines))
	})

	Context("when sourcesDir is empty", func() {
		BeforeEach(func() {
			sourcesDir = ""
		})

		It("returns error", func() {
			_, err := filereader.PipelinesFromFile(pipelinesFilename, sourcesDir)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the pipelines file cannot be read", func() {
		BeforeEach(func() {
			pipelinesFilename = "pipelines-never-written.yml"
		})

		It("returns error", func() {
			_, err := filereader.PipelinesFromFile(pipelinesFilename, sourcesDir)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the pipelines file fails to parse", func() {
		BeforeEach(func() {
			pipelinesFilename = "pipelines.yml"

			pipelinesFileContentsBytes := []byte(`{{`)

			err := ioutil.WriteFile(
				filepath.Join(sourcesDir, pipelinesFilename),
				pipelinesFileContentsBytes,
				os.ModePerm,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns error", func() {
			_, err := filereader.PipelinesFromFile(pipelinesFilename, sourcesDir)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when pipelines filename is empty", func() {
		BeforeEach(func() {
			pipelinesFilename = ""
		})

		It("returns nil pipelines without error", func() {
			returnedPipelines, err := filereader.PipelinesFromFile(pipelinesFilename, sourcesDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedPipelines).NotTo(BeNil())
			Expect(returnedPipelines).To(BeEmpty())
		})
	})
})
