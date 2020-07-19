package uaa


import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	dtlogger "github.com/sirupsen/logrus"
	kiterrors "go-microservices-platform/errors"
)

// UaaHelper represents a UAA utility for retrieving tokens and updating users.
type UaaHelper struct {
	address      string
	clientID     string
	clientSecret string
}

// NewUaaHelper returns a new struct of type UaaHelper with given credentials.
func NewUaaHelper(address, clientID, clientSecret string) *UaaHelper {
	return &UaaHelper{address: address, clientID: clientID, clientSecret: clientSecret}
}

// UAAClient represents a UAA utility for retrieving tokens and updating users.
type UAAClient struct {
	address      string
	clientID     string
	clientSecret string
}


// UaaTokenResponse represents a decoded json response from request for a UAA token.
type UaaTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string
	Jti          string
}




// GetBearerToken makes a HTTP request to UAA for a token. This function constructs
// a uaaTokenResponse and returns the access token as a string.
func (u *UaaHelper) GetBearerToken() (string, error) {
	tokenResponse, err := u.GetTokenResponse()
	if err != nil {
		fmt.Printf("Error getting bearer token: %s", err)
	}
	return tokenResponse.AccessToken, nil
}

// GetTokenResponse makes a call to UAA for a token response
func (u *UaaHelper) GetTokenResponse() (UaaTokenResponse, error) {
	tokenResponse := UaaTokenResponse{}
	tokenErrResponse := kiterrors.Error{}
	payload := fmt.Sprintf("grant_type=client_credentials&client_id=%s", u.clientID)
	uri := fmt.Sprintf("%s/oauth/token", u.address)
	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(payload))
	if err != nil {
		dtlogger.Error(fmt.Sprintf("creating request: %s", err.Error()))
		return tokenResponse, err
	}

	req.SetBasicAuth(u.clientID, u.clientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		dtlogger.Error(fmt.Sprintf("making request: %s", err.Error()))
		return tokenResponse, err
	}

	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		dtlogger.Error(fmt.Sprintf("reading body: %s", err.Error()))
		return tokenResponse, err
	}

	err = json.Unmarshal(responseBody, &tokenErrResponse)
	if err != nil {
		dtlogger.Error(fmt.Sprintf("unmarshalling body: %s", err.Error()))
		return tokenResponse, err

	}
	if tokenErrResponse.Type != "" {
		err := kiterrors.NewError(fmt.Sprintf("token is invalid due to %s", tokenErrResponse.Description))
		return tokenResponse, err
	}

	err = json.Unmarshal(responseBody, &tokenResponse)
	if err != nil {
		dtlogger.Error(fmt.Sprintf("unmarshalling body: %s", err.Error()))
		return tokenResponse, err

	}

	return tokenResponse, nil
}

func (u *UaaHelper) IsValidToken(token string) (bool, *kiterrors.Error) {

	isValid :=false

	if token == "" {
		return  isValid,kiterrors.ErrBadRequest
	}

	req, err := http.NewRequest("POST", u.address, bytes.NewBuffer([]byte("token="+token)))
	if err != nil {
		return isValid,kiterrors.ErrBadRequest
	}

	req.Header.Add("Authorization", "Basic "+u.clientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return isValid,kiterrors.ErrUnauthorized
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return isValid,kiterrors.ErrUnauthorized
		}
		return isValid,kiterrors.ErrUnauthorized
	}
	isValid=true
	return isValid,&kiterrors.Error{}
}


// UpdateUAAClientAuthorities takes in credentials and updates the clientID
// to have the given authorities. Currently used mainly for Multitenancy.
func UpdateUAAClientAuthorities(authURL string, authclientID string, authClientSecret string, updateclientID string, authorities []string) error {
	authorityUpdateTokenUaaHelper := NewUaaHelper(authURL, authclientID, authClientSecret)
	authorityUpdateToken, err := authorityUpdateTokenUaaHelper.GetBearerToken()
	authoritiesStr := ""
	for _, authority := range authorities {
		authoritiesStr = authoritiesStr + `"` + authority + `",`
	}
	authoritiesStr = strings.TrimRight(authoritiesStr, ",")
	authorityUpdateTokenBody := []byte(fmt.Sprintf(`{"client_id":"%s",
	  "authorities":[`+authoritiesStr+`]}`, updateclientID))

	authorityUpdateTokenBuf := bytes.NewBuffer(authorityUpdateTokenBody)
	client := http.DefaultClient
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/oauth/clients/%s", authURL, updateclientID),
		authorityUpdateTokenBuf)
	if err != nil {
		dtlogger.Error("unauthorized call to uaa for authorities update:", "error", err)
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", authorityUpdateToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		dtlogger.Error("UAA Client Authorities Response Error: ", "error", err)
		return err
	}
	defer resp.Body.Close()
	return nil
}

// ResetUAAClientAuthorities takes in credentials and updates the clientID
// to have minimal Authorities (uaa.resource).
func ResetUAAClientAuthorities(authURL string, authclientID string, authClientSecret string, updateclientID string) error {
	authorityUpdateTokenUaaHelper := NewUaaHelper(authURL, authclientID, authClientSecret)
	authorityUpdateToken, err := authorityUpdateTokenUaaHelper.GetBearerToken()
	authorityUpdateTokenBody := []byte(fmt.Sprintf(`{"client_id":"%s",
	  "authorities":["uaa.resource"]}`, updateclientID))

	authorityUpdateTokenBuf := bytes.NewBuffer(authorityUpdateTokenBody)

	client := http.DefaultClient
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/oauth/clients/%s", authURL, updateclientID),
		authorityUpdateTokenBuf)
	if err != nil {
		dtlogger.Error("unauthorized call to uaa for authorities update:", err)
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", authorityUpdateToken))
	req.Header.Set("Content-Type", "application/json")

	dtlogger.Debug("UAA Client Authorities Update url: ", req.URL.String(),
		"Headers: ", req.Header,
		"Request body: ", fmt.Sprintf(string(authorityUpdateTokenBody)), "")

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		dtlogger.Error("UAA Client Authorities Response Error: ", "error", err)
		return err
	}
	return nil
}
