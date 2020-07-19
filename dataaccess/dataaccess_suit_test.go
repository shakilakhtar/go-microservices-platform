package dataaccess_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDataAccess(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dataaccess")
}

