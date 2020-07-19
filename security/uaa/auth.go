package uaa

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/inconshreveable/log15"
	"io/ioutil"
)

// Auth is the main inteface exposed by the auth module providing encapsulation
type Auth interface {
	LoadUaaKeys(uaaURL string) error
	Protected(RequiredScopes, func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request)
	GetSupportedAlgorythms() []string
}

type BodyLoader func(url string) (string, error)

// RequiredScopes is a type containing the required scopes for a given URL
type RequiredScopes []string

type uaaKey struct {
	Alg   string `json:"alg"`
	Value string `json:"value"`
}

type uaaKeysResponse struct {
	Keys []uaaKey `json:"keys"`
}

var (
	logger = log15.New(log15.Ctx{"module": "auth"})
)

type keyMapping map[string]*rsa.PublicKey

// hidden datastructure being the actual real implementation of the auth module
type authInternal struct {
	uaaKeys    keyMapping
	bodyLoader BodyLoader
}

const (
	// ErrorCtxName is the name of the error string in log messages
	ErrorCtxName = "error"
)

// New creates a new uninitialized authentication module
func New(optionalBodyLoader ...BodyLoader) Auth {
	var bodyLoader BodyLoader
	switch {
	case len(optionalBodyLoader) == 0:
		// no value was provided, falling back to using HTTP
		bodyLoader = HTTPGetBodyAsString
	case len(optionalBodyLoader) > 0:
		// taking into consideration the first function only
		bodyLoader = optionalBodyLoader[0]
	}
	return &authInternal{uaaKeys: nil, bodyLoader: bodyLoader}
}

// LoadUaaKeys initializes the module by loading the public keys of the trusted UAA instance
func (auth *authInternal) LoadUaaKeys(uaaURL string) error {
	url := uaaURL + "/token_keys"
	data, loadErr := auth.bodyLoader(url)
	if loadErr != nil {
		return loadErr
	} // need to put the key into a local variable, because the value of the module level variable can't be altered in case of an error
	uaaKeysLocal, parseErr := parseConfigString(data)
	if parseErr != nil {
		return parseErr
	}
	auth.uaaKeys = uaaKeysLocal
	logger.Info("Successfuly loaded UAA keys, the service is operational")
	return nil
}

func parseConfigString(data string) (keyMapping, error) {
	var keyData uaaKeysResponse
	jsonErr := json.NewDecoder(strings.NewReader(data)).Decode(&keyData)
	if jsonErr != nil {
		logger.Error("Parsing UAA response JSON failed", log15.Ctx{ErrorCtxName: jsonErr.Error})
		return nil, jsonErr
	}
	uaaKeys := keyMapping{}
	for _, keyItem := range keyData.Keys {
		uaaKey, keyParsingErr := jwt.ParseRSAPublicKeyFromPEM([]byte(keyItem.Value))
		if keyParsingErr != nil {
			logger.Error("Parsing UAA key failed", log15.Ctx{ErrorCtxName: keyParsingErr.Error})
			return nil, keyParsingErr
		}
		uaaKeys[keyItem.Alg] = uaaKey
	}

	return uaaKeys, nil
}

// Protected returns a new, protected version of the originally provided HTTP handler function
func (auth *authInternal) Protected(requiredScopes RequiredScopes, protectedFunc func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if auth.uaaKeys == nil || len(auth.uaaKeys) == 0 {
			logger.Error("Trying to use an uninitialized authentication context")
		} else {
			authorization := r.Header.Get("authorization")
			if authorization == "" {
				logger.Debug("No authorization header was present")
			} else {
				splittedToken := strings.Split(authorization, " ")
				if len(splittedToken) != 2 || strings.ToLower(splittedToken[0]) != "bearer" {
					logger.Debug("The provided Authorization header value does not match the expectations")
				} else {
					keyFunc := func(token *jwt.Token) (interface{}, error) {
						if alg, ok := token.Header["alg"].(string); ok {
							if key, ok := auth.uaaKeys[alg]; ok {
								return key, nil
							}
							return nil, fmt.Errorf("No key found for alg %s", alg)
						}
						return nil, fmt.Errorf("No algorythm found on the token")
					}

					token, err := jwt.Parse(splittedToken[1], keyFunc)

					if err != nil || !token.Valid {
						logger.Debug("The token was invalid", log15.Ctx{ErrorCtxName: err.Error})
					} else {
						if hasScopes, scopesError := checkRequiredScopes(requiredScopes, token); scopesError != nil || !hasScopes {
							logger.Debug("The token did not have all the required scopes")
						} else {
							logger.Debug("The token was valid having all required scopes")
							// everything is OK -> calling the protected function
							protectedFunc(w, r)
							return
						}
					}

				}
			}
		}
		// in any failure case, sending back a 403
		w.WriteHeader(http.StatusForbidden)
	}
}

type claimSet map[string]bool

func extractClaimSet(claims jwt.MapClaims) claimSet {
	claimSet := claimSet{}
	if scopes, found := claims["scope"]; found {
		switch castedScopes := scopes.(type) {
		case []interface{}:
			for _, scope := range castedScopes {
				switch castedScope := scope.(type) {
				case string:
					claimSet[castedScope] = true
				default:
					// string should always be the case, adding this only to be sure
					claimSet[fmt.Sprintf("%s", castedScope)] = true
				}

			}
		}

	}
	return claimSet
}

func checkRequiredScopes(requiredScopes RequiredScopes, token *jwt.Token) (bool, error) {
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		claimSet := extractClaimSet(claims)
		for _, requiredScope := range requiredScopes {
			if _, present := claimSet[requiredScope]; !present {
				logger.Debug("Required scope not present on token", log15.Ctx{"missingScope": requiredScope})
				return false, nil
			}
		}
		return true, nil
	}
	return false, fmt.Errorf("The token had no valid scopes data embeded")
}

func (auth *authInternal) GetSupportedAlgorythms() []string {
	result := []string{}
	for alg := range auth.uaaKeys {
		result = append(result, alg)
	}
	return result
}

// GotBodyAsString is a small helper to load http request bodies as strings. This is suitable for moderate sized
// request bodies only.
func HTTPGetBodyAsString(url string) (body string, err error) {
	if resp, httpErr := http.Get(url); httpErr == nil {
		// not doing this would cause a memory leak
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if bytes, readError := ioutil.ReadAll(resp.Body); readError == nil {
				body = string(bytes)
			} else {
				err = readError
			}
		} else {
			err = fmt.Errorf("Received HTTP error status code %d", resp.StatusCode)
		}
	} else {
		err = httpErr
	}
	return
}
