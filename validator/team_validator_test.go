package validator_test

import (
	"github.com/eugenmayer/concourse-pipeline-resource/concourse"
	"github.com/eugenmayer/concourse-pipeline-resource/validator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ValidateTeams", func() {
	var (
		teams []concourse.Team
	)

	BeforeEach(func() {
		teams = []concourse.Team{
			{
				Name:     "some team",
				Username: "some username",
				Password: "some password",
			},
		}
	})

	Context("when all the necessary info is provided", func() {
		It("does not throw an error", func() {
			err := validator.ValidateTeams(teams)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when no auth info is provided", func() {
		BeforeEach(func() {
			teams[0].Username = ""
			teams[0].Password = ""
		})

		It("does not throw an error", func() {
			err := validator.ValidateTeams(teams)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when no team name is provided", func() {
		BeforeEach(func() {
			teams[0].Name = ""
		})

		It("returns an error", func() {
			err := validator.ValidateTeams(teams)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*name.*provided.*team.*0"))
		})
	})

	Context("when no team username is provided", func() {
		BeforeEach(func() {
			teams[0].Username = ""
		})

		It("returns an error", func() {
			err := validator.ValidateTeams(teams)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*username.*provided.*team.*%s", "some team"))
		})
	})

	Context("when no team password is provided", func() {
		BeforeEach(func() {
			teams[0].Password = ""
		})

		It("returns an error", func() {
			err := validator.ValidateTeams(teams)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*password.*provided.*team.*%s", "some team"))
		})
	})

	Context("when there are no teams", func() {
		It("returns an error", func() {
			err := validator.ValidateTeams([]concourse.Team{})
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(Equal("teams must be provided in source"))

			err = validator.ValidateTeams(nil)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(Equal("teams must be provided in source"))
		})
	})
})
