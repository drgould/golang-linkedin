package linkedin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	apiVer     = "/v2/"
	apiRoot    = "https://api.linkedin.com" + apiVer // api domain
	apiProfile = apiRoot + "me"                      // user root
	//PeopleURL https://docs.microsoft.com/en-us/linkedin/shared/integrations/people/profile-api?context=linkedin/marketing/context
	PeopleURL = apiRoot + "people/(id:{people-id})"
	apiGroup  = "groups/:id" // group root
	//OrgURL https://docs.microsoft.com/en-us/linkedin/marketing/integrations/community-management/organizations/organization-access-control#find-access-control-information
	OrgURL = apiRoot + "organizationAcls"
	//PageURL https://docs.microsoft.com/en-us/linkedin/marketing/integrations/community-management/organizations/organization-lookup-api h
	PageURL = apiRoot + "organizations"
	//ShareURL  https://docs.microsoft.com/en-us/linkedin/marketing/integrations/community-management/shares/share-api
	ShareURL = apiRoot + "shares"
	//CommentURL https://docs.microsoft.com/en-us/linkedin/marketing/integrations/community-management/shares/network-update-social-actions#retrieve-social-actions
	CommentURL = apiRoot + "socialActions/{activity-id}/comments"
	//AssetURL https://docs.microsoft.com/en-us/linkedin/marketing/integrations/community-management/shares/vector-asset-api
	AssetURL       = apiRoot + "/assets"
	authURL        = "https://www.linkedin.com/oauth/v2/authorization"
	accessTokenURL = "https://www.linkedin.com/oauth/v2/accessToken"
	scopes         = []string{"r_organization_social", "w_organization_social", "rw_organization_admin", "rw_ads", "r_ads_reporting", "r_liteprofile"}
	//ProfileURL ....
	ProfileURL = apiProfile
)

// api endpoint path
var apiUrls = map[string]string{
	"profile":     apiProfile,                         // user profile request
	"connections": apiProfile + "/connections:fields", // user connections request
	"group":       apiGroup + ":fields",               // group info request
}

// API base
type API struct {
	OauthKey        string // your oauth key
	OauthSecret     string // your oauth secret
	AccessToken     string // the user's access token
	RefreshToken    string
	ProtocolVersion uint
}

// SetCredentials your api key and secret
func (a *API) SetCredentials(key string, secret string) {
	a.OauthKey = key
	a.OauthSecret = secret
}

// SetToken the access token for this user
func (a *API) SetToken(token string) {
	a.AccessToken = token
}

// GetToken the user's access token
func (a API) GetToken() (t string) {
	return a.AccessToken
}

//AuthURL Compile the authentication URL
func (a API) AuthURL(state string, redirectURL string) (URL string) {
	scp := strings.Join(scopes, "%20")
	return authURL + "?response_type=code&client_id=" + a.OauthKey +
		"&state=" + state + "&redirect_uri=" + redirectURL + "&scope=" + scp
}

//Auth Convenience method to redirect the user to the authentication url
func (a API) Auth(w http.ResponseWriter, r *http.Request, state string, redirectURL string) {
	http.Redirect(w, r, a.AuthURL(state, redirectURL), http.StatusFound)
}

//RetrieveAccessToken Convert an authorization code to an access token
func (a *API) RetrieveAccessToken(client *http.Client, code string, redirectURL string) ([]byte, error) {
	var response []byte
	// send the request
	resp, err := client.Get(accessTokenURL + "?grant_type=authorization_code&code=" + code + "&redirect_uri=" +
		redirectURL + "&client_id=" + a.OauthKey + "&client_secret=" + a.OauthSecret)

	if err != nil {
		return response, err
	}
	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	return response, nil
}

// format the given user id for api calls
func getUserIdString(id string) (uid string) {
	if id == "~" || id == "" {
		return "~" // me
	} else if strings.Contains(id, "http") {
		return "url=" + url.QueryEscape(id) // someone else
	} else {
		return "id=" + id // someone else's url
	}
}

//getGroupIdString format the given group id for api calls
func getGroupIdString(id interface{}) (gid string, err error) {
	switch t := id.(type) {
	case string:
		if strings.Contains(id.(string), "http") {
			return "url=" + url.QueryEscape(id.(string)), nil // group url
		}
		return id.(string), nil // group id as a string
	case uint64:
		return strconv.FormatUint(id.(uint64), 10), nil // group id as an int
	default:
		return gid, fmt.Errorf("Group ID type exception: Expecting string or uint64 got %T", t)
	}
}

//Raw Make an http request to get results
func (a API) Raw(client *http.Client, u string, params url.Values) ([]byte, error) {
	req, err := http.NewRequest("GET", u, nil) // make a new request
	if err != nil {
		return nil, err
	}
	if params != nil {
		req.URL.RawQuery = params.Encode()
	}
	token := a.GetToken()
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("X-Restli-Protocol-Version", "2.0.0")

	r, err := client.Do(req) // send the request

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(r.Body) // read the response data
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return data, nil
}

// Convenience method for normal api calls
func (a API) request(client *http.Client, endpoint string, options map[string]string, params url.Values) ([]byte, error) {
	ep, ok := apiUrls[endpoint]
	if !ok {
		return nil, errors.New("Endpoint \"" + endpoint + "\" not defined")
	}

	for field, value := range options {
		ep = strings.Replace(ep, ":"+field, value, -1)
	}

	if len(params) > 0 {
		ep += "?" + params.Encode()
	}

	return a.Raw(client, ep, nil)
}

//RawNonHeader Conveient method for raw api calls which returns JSON in bytes format
//
//This is an open format so anyone if wanted to unmarshal to struct or any map[string]interface{}
func (a API) RawNonHeader(client *http.Client, URL string, params url.Values) (j []byte, e error) {
	req, err := http.NewRequest("GET", URL, nil) // make a new request
	if err != nil {
		return nil, err
	}
	if params != nil {
		req.URL.RawQuery = params.Encode()
	}
	token := a.GetToken()
	req.Header.Add("Authorization", "Bearer "+token)
	r, err := client.Do(req) // send the request
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(r.Body) // read the response data
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return data, nil
}

//SendRequest Convenient method for normal POST/PUT call
//
//This api call will allow you to submit a comment and shares to linkedin
//
//https://docs.microsoft.com/en-us/linkedin/marketing/integrations/community-management/shares/share-api
func (a *API) SendRequest(client *http.Client, URL string, params interface{}) ([]byte, error) {
	if params == nil {
		return nil, errors.New("empty params can not send")
	}
	jsonByte, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	log.Printf("jsonByte=%v", string(jsonByte))
	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewBuffer(jsonByte))
	if err != nil {
		return nil, err
	}
	token := a.GetToken()
	req.Header.Add("Authorization", "Bearer "+token)
	if a.ProtocolVersion == 2 {
		req.Header.Add("X-Restli-Protocol-Version", "2.0.0")
		req.Header.Add("Content-Type", "application/json")
	}

	r, err := client.Do(req) // send the request

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(r.Body) // read the response data
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return data, nil
}
