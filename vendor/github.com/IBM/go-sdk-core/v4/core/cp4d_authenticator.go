package core

// (C) Copyright IBM Corp. 2019, 2021.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//
// CloudPakForDataAuthenticator uses either a username/password pair or a
// username/apikey pair to obtain a suitable bearer token from the CP4D authentication service,
// and adds the bearer token to requests via an Authorization header of the form:
//
// 		Authorization: Bearer <bearer-token>
//
type CloudPakForDataAuthenticator struct {
	// The URL representing the Cloud Pak for Data token service endpoint [required].
	URL string

	// The username used to obtain a bearer token [required].
	Username string

	// The password used to obtain a bearer token [required if APIKey not specified].
	// One of Password or APIKey must be specified.
	Password string

	// The apikey used to obtain a bearer token [required if Password not specified].
	// One of Password or APIKey must be specified.
	APIKey string

	// A flag that indicates whether verification of the server's SSL certificate
	// should be disabled; defaults to false [optional].
	DisableSSLVerification bool

	// Default headers to be sent with every CP4D token request [optional].
	Headers map[string]string

	// The http.Client object used to invoke token server requests [optional]. If
	// not specified, a suitable default Client will be constructed.
	Client *http.Client

	// The cached token and expiration time.
	tokenData *cp4dTokenData
}

var cp4dRequestTokenMutex sync.Mutex
var cp4dNeedsRefreshMutex sync.Mutex

// NewCloudPakForDataAuthenticator constructs a new CloudPakForDataAuthenticator
// instance from a username/password pair.
// This is the default way to create an authenticator and is a wrapper around
// the NewCloudPakForDataAuthenticatorUsingPassword() function
func NewCloudPakForDataAuthenticator(url string, username string, password string,
	disableSSLVerification bool, headers map[string]string) (*CloudPakForDataAuthenticator, error) {
	return NewCloudPakForDataAuthenticatorUsingPassword(url, username, password, disableSSLVerification, headers)
}

// NewCloudPakForDataAuthenticatorUsingPassword constructs a new CloudPakForDataAuthenticator
// instance from a username/password pair.
func NewCloudPakForDataAuthenticatorUsingPassword(url string, username string, password string,
	disableSSLVerification bool, headers map[string]string) (*CloudPakForDataAuthenticator, error) {
	return newAuthenticator(url, username, password, "", disableSSLVerification, headers)
}

// NewCloudPakForDataAuthenticatorUsingAPIKey constructs a new CloudPakForDataAuthenticator
// instance from a username/apikey pair.
func NewCloudPakForDataAuthenticatorUsingAPIKey(url string, username string, apikey string,
	disableSSLVerification bool, headers map[string]string) (*CloudPakForDataAuthenticator, error) {
	return newAuthenticator(url, username, "", apikey, disableSSLVerification, headers)
}

func newAuthenticator(url string, username string, password string, apikey string,
	disableSSLVerification bool, headers map[string]string) (authenticator *CloudPakForDataAuthenticator, err error) {

	authenticator = &CloudPakForDataAuthenticator{
		Username:               username,
		Password:               password,
		APIKey:                 apikey,
		URL:                    url,
		DisableSSLVerification: disableSSLVerification,
		Headers:                headers,
	}

	// Make sure the config is valid.
	err = authenticator.Validate()
	if err != nil {
		return nil, err
	}

	return
}

// newCloudPakForDataAuthenticatorFromMap : Constructs a new CloudPakForDataAuthenticator instance from a map.
func newCloudPakForDataAuthenticatorFromMap(properties map[string]string) (*CloudPakForDataAuthenticator, error) {
	if properties == nil {
		return nil, fmt.Errorf(ERRORMSG_PROPS_MAP_NIL)
	}

	disableSSL, err := strconv.ParseBool(properties[PROPNAME_AUTH_DISABLE_SSL])
	if err != nil {
		disableSSL = false
	}

	return newAuthenticator(properties[PROPNAME_AUTH_URL],
		properties[PROPNAME_USERNAME], properties[PROPNAME_PASSWORD],
		properties[PROPNAME_APIKEY], disableSSL, nil)
}

// AuthenticationType returns the authentication type for this authenticator.
func (CloudPakForDataAuthenticator) AuthenticationType() string {
	return AUTHTYPE_CP4D
}

// Validate the authenticator's configuration.
//
// Ensures the username, password, and url are not Nil. Additionally, ensures
// they do not contain invalid characters.
func (authenticator CloudPakForDataAuthenticator) Validate() error {

	if authenticator.Username == "" {
		return fmt.Errorf(ERRORMSG_PROP_MISSING, "Username")
	}

	// The user should specify exactly one of APIKey or Password.
	if (authenticator.APIKey == "" && authenticator.Password == "") ||
		(authenticator.APIKey != "" && authenticator.Password != "") {
		return fmt.Errorf(ERRORMSG_EXCLUSIVE_PROPS_ERROR, "APIKey", "Password")
	}

	if authenticator.URL == "" {
		return fmt.Errorf(ERRORMSG_PROP_MISSING, "URL")
	}

	return nil
}

// Authenticate adds the bearer token (obtained from the token server) to the
// specified request.
//
// The CP4D bearer token will be added to the request's headers in the form:
//
// 		Authorization: Bearer <bearer-token>
//
func (authenticator *CloudPakForDataAuthenticator) Authenticate(request *http.Request) error {
	token, err := authenticator.getToken()
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", fmt.Sprintf(`Bearer %s`, token))
	return nil
}

// getToken: returns an access token to be used in an Authorization header.
// Whenever a new token is needed (when a token doesn't yet exist, needs to be refreshed,
// or the existing token has expired), a new access token is fetched from the token server.
func (authenticator *CloudPakForDataAuthenticator) getToken() (string, error) {
	if authenticator.tokenData == nil || !authenticator.tokenData.isTokenValid() {
		// synchronously request the token
		err := authenticator.synchronizedRequestToken()
		if err != nil {
			return "", err
		}
	} else if authenticator.tokenData.needsRefresh() {
		// If refresh needed, kick off a go routine in the background to get a new token
		ch := make(chan error)
		go func() {
			ch <- authenticator.getTokenData()
		}()
		select {
		case err := <-ch:
			if err != nil {
				return "", err
			}
		default:
		}
	}

	// return an error if the access token is not valid or was not fetched
	if authenticator.tokenData == nil || authenticator.tokenData.AccessToken == "" {
		return "", fmt.Errorf("Error while trying to get access token")
	}

	return authenticator.tokenData.AccessToken, nil
}

// synchronizedRequestToken: synchronously checks if the current token in cache
// is valid. If token is not valid or does not exist, it will fetch a new token
// and set the tokenRefreshTime
func (authenticator *CloudPakForDataAuthenticator) synchronizedRequestToken() error {
	cp4dRequestTokenMutex.Lock()
	defer cp4dRequestTokenMutex.Unlock()
	// if cached token is still valid, then just continue to use it
	if authenticator.tokenData != nil && authenticator.tokenData.isTokenValid() {
		return nil
	}

	return authenticator.getTokenData()
}

// getTokenData: requests a new token from the token server and
// unmarshals the token information to the tokenData cache. Returns
// an error if the token was unable to be fetched, otherwise returns nil
func (authenticator *CloudPakForDataAuthenticator) getTokenData() error {
	tokenResponse, err := authenticator.requestToken()
	if err != nil {
		authenticator.tokenData = nil
		return err
	}

	authenticator.tokenData, err = newCp4dTokenData(tokenResponse)
	if err != nil {
		authenticator.tokenData = nil
		return err
	}

	return nil
}

// cp4dRequestBody is a struct used to model the request body for the "POST /v1/authorize" operation.
// Note: we list both Password and APIKey fields, although exactly one of those will be used for
// a specific invocation of the POST /v1/authorize operation.
type cp4dRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
}

// requestToken: fetches a new access token from the token server.
func (authenticator *CloudPakForDataAuthenticator) requestToken() (tokenResponse *cp4dTokenServerResponse, err error) {

	// Create the request body (only one of APIKey or Password should be set
	// on the authenticator so only one of them should end up in the serialized JSON).
	body := &cp4dRequestBody{
		Username: authenticator.Username,
		Password: authenticator.Password,
		APIKey:   authenticator.APIKey,
	}

	builder := NewRequestBuilder(POST)
	_, err = builder.ResolveRequestURL(authenticator.URL, "/v1/authorize", nil)
	if err != nil {
		return
	}

	// Add user-defined headers to request.
	for headerName, headerValue := range authenticator.Headers {
		builder.AddHeader(headerName, headerValue)
	}

	// Add the Content-Type header.
	builder.AddHeader("Content-Type", "application/json")

	// Add the request body to request.
	_, err = builder.SetBodyContentJSON(body)
	if err != nil {
		return
	}

	// Build the request object.
	req, err := builder.Build()
	if err != nil {
		return
	}

	// If the authenticator does not have a Client, create one now.
	if authenticator.Client == nil {
		authenticator.Client = &http.Client{
			Timeout: time.Second * 30,
		}

		// If the user told us to disable SSL verification, then do it now.
		if authenticator.DisableSSLVerification {
			transport := &http.Transport{
				/* #nosec G402 */
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			authenticator.Client.Transport = transport
		}
	}

	resp, err := authenticator.Client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buff := new(bytes.Buffer)
		_, _ = buff.ReadFrom(resp.Body)

		// Create a DetailedResponse to be included in the error below.
		detailedResponse := &DetailedResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			RawResult:  buff.Bytes(),
		}

		err = NewAuthenticationError(detailedResponse, fmt.Errorf(buff.String()))
		return
	}

	tokenResponse = &cp4dTokenServerResponse{}
	err = json.NewDecoder(resp.Body).Decode(tokenResponse)
	defer resp.Body.Close()
	if err != nil {
		err = fmt.Errorf(ERRORMSG_UNMARSHAL_AUTH_RESPONSE, err.Error())
		tokenResponse = nil
		return
	}

	return
}

// cp4dTokenServerResponse is a struct that models a response received from the token server.
type cp4dTokenServerResponse struct {
	Token       string `json:"token,omitempty"`
	MessageCode string `json:"_messageCode_,omitempty"`
	Message     string `json:"message,omitempty"`
}

// cp4dTokenData is a struct that represents the cached information related to a fetched access token.
type cp4dTokenData struct {
	AccessToken string
	RefreshTime int64
	Expiration  int64
}

// newCp4dTokenData: constructs a new Cp4dTokenData instance from the specified Cp4dTokenServerResponse instance.
func newCp4dTokenData(tokenResponse *cp4dTokenServerResponse) (*cp4dTokenData, error) {
	// Need to crack open the access token (a JWT) to get the expiration and issued-at times.
	claims, err := parseJWT(tokenResponse.Token)
	if err != nil {
		return nil, err
	}

	// Compute the adjusted refresh time (expiration time - 20% of timeToLive)
	timeToLive := claims.ExpiresAt - claims.IssuedAt
	expireTime := claims.ExpiresAt
	refreshTime := expireTime - int64(float64(timeToLive)*0.2)

	tokenData := &cp4dTokenData{
		AccessToken: tokenResponse.Token,
		Expiration:  expireTime,
		RefreshTime: refreshTime,
	}

	return tokenData, nil
}

// isTokenValid: returns true iff the Cp4dTokenData instance represents a valid (non-expired) access token.
func (tokenData *cp4dTokenData) isTokenValid() bool {
	if tokenData.AccessToken != "" && GetCurrentTime() < tokenData.Expiration {
		return true
	}
	return false
}

// needsRefresh: synchronously returns true iff the currently stored access token should be refreshed. This method also
// updates the refresh time if it determines the token needs refreshed to prevent other threads from
// making multiple refresh calls.
func (tokenData *cp4dTokenData) needsRefresh() bool {
	cp4dNeedsRefreshMutex.Lock()
	defer cp4dNeedsRefreshMutex.Unlock()

	// Advance refresh by one minute
	if tokenData.RefreshTime >= 0 && GetCurrentTime() > tokenData.RefreshTime {
		tokenData.RefreshTime = GetCurrentTime() + 60
		return true
	}
	return false
}
