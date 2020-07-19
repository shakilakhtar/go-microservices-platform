package handler_test

import (

	."github.com/onsi/ginkgo"
	"net/http/httptest"

)
var (
	recorder *httptest.ResponseRecorder
)

var _=Describe("tranport",func(){

	BeforeEach(func(){
		recorder=httptest.NewRecorder()
	})

	Context("Handler", func(){
		It("",func(){

		})

	})
})