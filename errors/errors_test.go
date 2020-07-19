package errors_test

import (
	"io/ioutil"
	"os"
	"time"

	kiterrors "shakilakhtar/go-microservices-platform/errors"
	logger "github.com/sirupsen/logrus"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	validErrMsg = "Example Error"
)

var _ = Describe("Errors", func() {

	Describe("When sending an error string", func() {
		Context("If error string is not empty", func() {
			It("error should be logged", func() {

				pwd, err := os.Getwd()
				if err != nil {
					logger.Error("Could not get directory path", err)
				}

				logfile := pwd + "/mytest.log"
				//logger.SetupFileOutput("Error Test", logfile, true)

				err = kiterrors.HandleError(validErrMsg)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal(validErrMsg))

				time.Sleep(2)
				logContent, err := ioutil.ReadFile(logfile)
				if err != nil {
					logger.Error("Failed to read from log file", err)
				}

				Expect(string(logContent)).Should(ContainSubstring(validErrMsg))
				os.Remove(logfile)
			})
		})
	})

	Describe("When sending an error string", func() {
		Context("If error string is not empty", func() {
			It("error should be returned", func() {

				err := kiterrors.HandleError(validErrMsg)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal(validErrMsg))

			})
		})
	})
})
