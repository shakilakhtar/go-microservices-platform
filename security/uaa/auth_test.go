package uaa_test

import (
	"net/http"
	"time"

	"crypto/rsa"
	"fmt"

	. "shakilakhtar/go-microservices-platform/security/uaa"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/inconshreveable/log15"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var testLogger = log15.New(log15.Ctx{"module": "auth_test"})

const (
	// to get the key value from an ssh public key, use openssl rsa -in <<file name>> -pubout
	sampleUAAKeysJSON = `{
    "keys": [
        {
            "kty": "RSA",
            "e": "AQAB",
            "use": "sig",
            "kid": "legacy-token-key",
            "alg": "RS256",
            "value": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA68HqdNKAHI7efg8GR/zf\n71sNkrwIOZMsgSrXgCVnefX9i86vuKMkL5If/wS9lofq7uHkKHSoe3E7dLeU1K4E\nDJYo53LALxSnSjpx+f3Q2wecpYGFGbQuAo5EEX9PBfB2pFKDQUMQiK2Pa6C5lgs3\nA2rQiUhWoUgtIb8rWIJ9DESjrBfNZxFuESGmjntoESLNYAqwoEOFGMVoXmkGYZEd\nSIS5ltWgA7nz3B549Y6FyXe1HO1BOR3ycsLg0YQb4UuSVyqOjgRO1AyFfhvZ0rCV\nsmfhXqQ685iPDyj0XX9/H+cI7iQr0hpuV18D0x4KjOClr0dZNohljMxhnKQRmp/r\njQIDAQAB\n-----END PUBLIC KEY-----",
            "n": "AMYxBBzSwaNlJwixok3_6ayAB5sIcOaAvklhfm4dsRowquL49jK0SrxyMz1_aV2c5p_obWoSFtvEaDWsx4rYK8K1q_uq324DQ4ENW9TeO6gYmc60cc4HGRUcdjnz_DAV7bwU-7V5QCYXRFpSVlHxrQpbjxuPUc-EneJ2qBhJXaQCkNCNhIo1QhkIdb7RDngOaXzwkn4TvQPw5U-5JeHbKLXU6HBvcw4RpsdPpGkolfavrKPUZ5kiVoC5TPuSmpSshRaCtz0YHC7IzH_gxCw-xd10zZ4BU418MvghV21grAIey_Dng3YvSuOdm2_CQyof2faQbT_NDwZUhT9DQKwFMU0="
        }
    ]
}`
	// private key only needed to create valid tokens for testing
	// to get it from an ssh private rsa key, use openssl pkcs8 -topk8 -inform PEM -outform PEM -nocrypt -in <<file name>>
	testPrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDrwep00oAcjt5+
DwZH/N/vWw2SvAg5kyyBKteAJWd59f2Lzq+4oyQvkh//BL2Wh+ru4eQodKh7cTt0
t5TUrgQMlijncsAvFKdKOnH5/dDbB5ylgYUZtC4CjkQRf08F8HakUoNBQxCIrY9r
oLmWCzcDatCJSFahSC0hvytYgn0MRKOsF81nEW4RIaaOe2gRIs1gCrCgQ4UYxWhe
aQZhkR1IhLmW1aADufPcHnj1joXJd7Uc7UE5HfJywuDRhBvhS5JXKo6OBE7UDIV+
G9nSsJWyZ+FepDrzmI8PKPRdf38f5wjuJCvSGm5XXwPTHgqM4KWvR1k2iGWMzGGc
pBGan+uNAgMBAAECggEBAIWDtX7hc5I4ywJDCgCc0klgnIg8GsBYe/zOwWqeRELK
sAOMUvHS2nxiWeJ30dK9OKx+m1LZ9kyqbMyF5zCnOD3UkGe7EeHX5YHhJYk5WB2i
6vDEMBfFdcUWwq/SFHO9ocMfw5ujGmr9N9rxFAlIYqh5xo3ovL7r/Ds/Y5HlnGHk
Kwi7Ch9nPzc7FEowstKZq+B6kRt80KqhZkbrmmEXuxfGYOOZjt5CcQEhaa1n7g/Q
OO8ohTikp2BC1/00INet1PsOOZ4R55HFOIqLbA763G8wZXMaAZodYrK+eEinht29
z1v4wN5r9cShF22ntBV13pUAGNfkvAuK8JPnTgFoE+ECgYEA/p31QWHal1ECAGpi
QfOiLEGymMh9MkJvf7yfEG2yQikc0G5jHgyzywbboHWa4waY4j/lZpBQP13/5xbv
1O9IX0S4yDNbfsQ3c2E8vW3uvgUpfDIC/f4wAmi9jO77qQVejTrbwkU1YyjXZTXr
EgqNnu34bjf+CklAIngEjZk2o4UCgYEA7Qm73XvD/BHH5w0kvI+f8M3Y8qT2JIuR
Acju1SZ6AbCdhTMps5f0Nd9LSFuFnGA8BjEbmxyc8rC0i1PR2cJzIJSnVgUNGscQ
uZQNqGjgI+hHysVSuCBIgUg59K8xp/p0hzFqk53s/sMhS4kgZAagDqIHyQm+sCf4
yNozOc+xkmkCgYEA83CLQYwRt4NYapVMhMowUCgwXiuyqA8lE/iADPEU8nTke9RP
KDf03zUbX/uRr2ZrXkbBSqLIVw3E0mn3vJtbktrd4WxZGob4jXR24pbtIPlGhNw8
SCR0OplyQgFs1Fmx4U5ZNxF8zeYKq1Y1/vXgGghk8tzOI3+NtmcR02CeARECgYA1
2Qg8gGk9Uiy/aFT4IQiMg7bNKHxiQPJoHWVkNqFw0NZ38+99RP/NXTSU83We2J3K
Kk3DJvTgjRP2ssvxVCMjO6HoAK3Bb4d1IRUZNPn2LkZg4gKwoWTXObkwxLvbFSJz
s94qOq4kEd/2cOhS0M57hIOQQA55phr2RdttPqlwQQKBgGfzvgrZQ/SRop8uq9Cg
byP09wL/pDVUO3NvScV+mzt1X7qVhRcG9RkT5CbdBdLyy1HJrrEM4TWeTyn6oYbE
ExKry/UrXWcTtEscRzB301BcRHneeOkPEJNVAwhyjVsJ7DVDKMm9baXloWPBbR9n
rUS4vZufOUcsZfkxgbIz6eI+
-----END PRIVATE KEY-----`
)

func getTestPrivateKey() (*rsa.PrivateKey, error) {
	privateKey, keyParsingErr := jwt.ParseRSAPrivateKeyFromPEM([]byte(testPrivateKey))
	if keyParsingErr != nil {
		return nil, fmt.Errorf("Error while parsing test private key - %s", keyParsingErr.Error())
	}
	return privateKey, nil
}

func createTokenWithClaims(scopes []string) string {
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = "Jon Snow"
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	claims["scope"] = scopes
	signingKey, keyErr := getTestPrivateKey()
	if keyErr != nil {
		testLogger.Error("Error loading private signing key", log15.Ctx{ErrorCtxName: keyErr.Error()})
		return ""
	}
	ss, signingErr := token.SignedString(signingKey)
	if signingErr != nil {
		testLogger.Error("Error signing the token ", log15.Ctx{ErrorCtxName: signingErr.Error()})
		return ""
	}
	return ss
}

var _ = Describe("Auth", func() {
	var authCtx Auth
	didExecute := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		didExecute = true
	}

	BeforeEach(func() {
		authCtx = New(func(url string) (string, error) {
			return sampleUAAKeysJSON, nil
		})
		authCtx.LoadUaaKeys("URL not actually used in this test")
		didExecute = false
	})

	Describe("Parsing UAA response", func() {
		Context("using input data sampleUAAKeysJSON", func() {
			It("should have a single valid RSA key", func() {
				algs := authCtx.GetSupportedAlgorythms()
				Expect(len(algs)).To(Equal(1))
				Expect(algs[0]).To(Equal("RS256"))
			})
		})
	})

	Describe("Protecting handler function", func() {
		Context("having request without an Authorization header", func() {
			It("should fail", func() {
				requiredScopes := RequiredScopes{"admin", "user"}
				protectedHandler := authCtx.Protected(requiredScopes, handler)
				status := 0
				w := writerMock{statusHeader: &status}
				protectedHandler(w, &http.Request{Header: http.Header{}})
				Expect(status).To(Equal(http.StatusForbidden))
				Expect(didExecute).To(Equal(false))
			})
		})

		Context("having request with an invalid token", func() {
			It("should fail", func() {
				requiredScopes := RequiredScopes{"admin", "user"}
				protectedHandler := authCtx.Protected(requiredScopes, handler)
				status := 0
				w := writerMock{statusHeader: &status}
				protectedHandler(w, &http.Request{Header: http.Header{"Authorization": []string{"Bearer abc"}}})
				Expect(status).To(Equal(http.StatusForbidden))
				Expect(didExecute).To(Equal(false))
			})
		})

		Context("having a valid request without any scopes", func() {
			It("should execute the protected handler", func() {
				requiredScopes := RequiredScopes{}
				protectedHandler := authCtx.Protected(requiredScopes, handler)
				status := 0
				w := writerMock{statusHeader: &status}
				protectedHandler(w, &http.Request{Header: http.Header{"Authorization": []string{"Bearer " + createTokenWithClaims(requiredScopes)}}})
				Expect(status).To(Equal(http.StatusOK))
				Expect(didExecute).To(Equal(true))
			})
		})

		Context("having a valid request with all required scopes", func() {
			It("should execute the protected handler", func() {
				requiredScopes := RequiredScopes{"admin", "user"}
				protectedHandler := authCtx.Protected(requiredScopes, handler)
				status := 0
				w := writerMock{statusHeader: &status}
				protectedHandler(w, &http.Request{Header: http.Header{"Authorization": []string{"Bearer " + createTokenWithClaims(requiredScopes)}}})
				Expect(status).To(Equal(http.StatusOK))
				Expect(didExecute).To(Equal(true))
			})
		})

		Context("having a valid request with one missing required scope", func() {
			It("should fail", func() {
				requiredScopes := RequiredScopes{"admin", "user"}
				protectedHandler := authCtx.Protected(requiredScopes, handler)
				status := 0
				w := writerMock{statusHeader: &status}
				protectedHandler(w, &http.Request{Header: http.Header{"Authorization": []string{"Bearer " + createTokenWithClaims(RequiredScopes{"user"})}}})
				Expect(status).To(Equal(http.StatusForbidden))
				Expect(didExecute).To(Equal(false))
			})
		})
	})
})

type writerMock struct {
	statusHeader *int
	headers      http.Header
}

func (w writerMock) Header() http.Header {
	return w.headers
}

func (w writerMock) Write(data []byte) (int, error) {
	return len(data), nil
}

func (w writerMock) WriteHeader(header int) {
	*w.statusHeader = header
}
