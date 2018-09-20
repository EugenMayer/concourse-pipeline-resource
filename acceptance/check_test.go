package acceptance

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/eugenmayer/concourse-pipeline-resource/concourse"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const (
	checkTimeout = 40 * time.Second
)

var _ = Describe("Check", func() {
	var (
		command       *exec.Cmd
		checkRequest  concourse.CheckRequest
		stdinContents []byte
	)

	BeforeEach(func() {
		By("Creating command object")
		command = exec.Command(checkPath)

		By("Creating default request")
		checkRequest = concourse.CheckRequest{
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
			Version: concourse.Version{},
		}

		var err error
		stdinContents, err = json.Marshal(checkRequest)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("successful behavior", func() {
		It("returns pipeline versions without error", func() {
			By("Running the command")
			session := run(command, stdinContents)

			By("Validating command exited without error")
			Eventually(session, checkTimeout).Should(gexec.Exit(0))

			var resp concourse.CheckResponse
			err := json.Unmarshal(session.Out.Contents(), &resp)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(resp)).To(BeNumerically(">", 0))
			for _, v := range resp {
				Expect(v).NotTo(BeEmpty())
			}
		})

		Context("target not provided", func() {
			BeforeEach(func() {
				var err error
				err = os.Setenv("ATC_EXTERNAL_URL", checkRequest.Source.Target)
				Expect(err).ShouldNot(HaveOccurred())

				checkRequest.Source.Target = ""

				stdinContents, err = json.Marshal(checkRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		It("returns pipeline version without error", func() {
			By("Running the command")
			session := run(command, stdinContents)

			By("Validating command exited without error")
			Eventually(session, checkTimeout).Should(gexec.Exit(0))

			var resp concourse.CheckResponse
			err := json.Unmarshal(session.Out.Contents(), &resp)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(resp)).To(BeNumerically(">", 0))
		})
	})

	Context("when validation fails", func() {
		BeforeEach(func() {
			checkRequest.Source.Teams = nil

			var err error
			stdinContents, err = json.Marshal(checkRequest)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("exits with error", func() {
			By("Running the command")
			session := run(command, stdinContents)

			By("Validating command exited with error")
			Eventually(session, checkTimeout).Should(gexec.Exit(1))
			Expect(session.Err).Should(gbytes.Say(".*teams.*provided"))
		})
	})
})
