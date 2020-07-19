package uaa_test

import (
	"net/http"
	. "github.com/onsi/ginkgo"
)

type dummyHandler struct{}

func (d dummyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

var _ = Describe("Basic Authentication Handler", func() {

})
