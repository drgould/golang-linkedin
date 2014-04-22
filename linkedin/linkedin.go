package linkedin

import (
	"net/http"
	"net/url"
	"strings"
	"strconv"
	"io/ioutil"
	"errors"
	"encoding/json"
	"fmt"
)

var apiRoot = "https://api.linkedin.com"	// api domain
var apiUser = "/v1/people/:id"				// user root
var apiGroup = "/v1/groups/:id"				// group root

// api endpoint path
var apiUrls = map[string]string{
	"profile": apiUser+":fields",					// user profile request
	"connections": apiUser+"/connections:fields", 	// user connections request
	"group": apiGroup+":fields",					// group info request
}

// api base
type API struct {
	oauth_key string		// your oauth key
	oauth_secret string		// your oauth secret
	access_token string		// the user's access token
}

// Set your api key and secret
func (a *API) SetCredentials(key string, secret string) {
	a.oauth_key = key
	a.oauth_secret = secret
}

// Set the access token for this user
func (a *API) SetToken(token string) {
	a.access_token = token
}

// Get the user's access token
func (a API) GetToken() (t string) {
	return a.access_token
}

// Compile the authentication URL
func (a API) AuthUrl(state string, redirect_url string) (url string) {
	return "https://www.linkedin.com/uas/oauth2/authorization?response_type=code&client_id="+a.oauth_key+
								 "&state="+state+"&redirect_uri="+redirect_url
}

// Convenience method to redirect the user to the authentication url
func (a API) Auth(w http.ResponseWriter, r *http.Request, state string, redirect_url string) {
	http.Redirect(w, r, a.AuthUrl(state, redirect_url), http.StatusFound)
}

// Convert an authorization code to an access token
func (a *API) RetrieveAccessToken(client *http.Client, code string, redirect_url string) (t string, e error) {

	// send the request
	resp, err := client.Get("https://www.linkedin.com/uas/oauth2/accessToken?grant_type=authorization_code&code="+code+"&redirect_uri="+
						redirect_url+"&client_id="+a.oauth_key+"&client_secret="+a.oauth_secret)
	
	if err != nil {
		return t, err
	}

	// read the response data
	data, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	// decode the response data to json
	var response map[string]interface{}
	err = json.Unmarshal(data, &response)
	
	if err != nil {
		return t, err
	}

	// if there is an "error" index something went wrong
	if _, err := response["error"]; err {
		return t, errors.New(response["error"].(string)+" - "+response["error_description"].(string))
	}

	// pull out the token
	t = response["access_token"].(string)

	// set my access token
	a.SetToken(t)

	// return token
	return t, nil
}

// format the given user id for api calls
func getUserIdString(id string) (uid string) {
	if id == "~" || id == "" {
		return "~"								// me
	} else if strings.Contains(id, "http") {
		return "url="+url.QueryEscape(id)		// someone else
	} else {
		return "id="+id							// someone else's url
	}
}

// format the given group id for api calls
func getGroupIdString(id interface{}) (gid string, err error) {
	switch t := id.(type) {
		case string:
			if strings.Contains(id.(string), "http") {
				return "url="+url.QueryEscape(id.(string)), nil		// group url
			}
			return id.(string), nil									// group id as a string
		case uint64:
			return strconv.FormatUint(id.(uint64),10), nil			// group id as an int
		default:
			return gid, errors.New(fmt.Sprintf("Group ID type exception: Expecting string or uint64 got %T", t))
	}
}

// Make a call to get info about the given user's profile
func (a API) Profile(client *http.Client, user_id string, fields Fields) (j map[string]interface{}, err error) {
	f := ""
	if fields != nil {
		f = fields.Encode()
	}

	return a.request(client, "profile", map[string]string{
		"id": getUserIdString(user_id),
		"fields": f,
	}, nil)
}

// Make a call to get info about the given user's connections
func (a API) Connections(client *http.Client, user_id string, fields Fields, params url.Values) (j map[string]interface{}, err error) {
	f := ""
	if fields != nil {
		f = fields.Encode()
	}

	return a.request(client, "connections", map[string]string{
		"id": getUserIdString(user_id),
		"fields": f,
	}, params)
}

// Make a call to get info about the given group
func (a API) Group(client *http.Client, group_id interface{}, fields Fields) (j map[string]interface{}, err error) {
	f := ""
	if fields != nil {
		f = fields.Encode()
	}

	gid, err := getGroupIdString(group_id)
	if err != nil {
		return j, err
	}
	
	return a.request(client, "group", map[string]string{
		"id": gid,
		"fields": f,
	}, nil)
}

// Make a raw api call
func (a API) Raw(client *http.Client, u interface{}) (j map[string]interface{}, e error) {
	endpoint := url.URL{}	// initialize the url
	
	switch t := u.(type) {
		default:
			return nil, errors.New(fmt.Sprintf("Expecting string or *url.URL, got %v: %#v", t, u))
		case string:							// the url provided is a string so we need to parse it
			ep, err := url.Parse(u.(string))
			if err != nil {
				return nil, err
			}
			endpoint = *ep
		case url.URL:							// the url provided is already parsed
			endpoint = u.(url.URL)
	}
	
	qs := endpoint.Query()
	qs.Add("oauth2_access_token",a.access_token)	// add the access token to the query
	
	req, _ := http.NewRequest("GET", apiRoot+endpoint.Path+"?"+qs.Encode(), nil)	// make a new request
	req.URL.Opaque = endpoint.Path			// make sure it doesn't query string encode the path
	req.Header.Add("x-li-format","json")	// we want json
	
	r, err := client.Do(req)				// send the request
	
	if err != nil {
		return nil, err
	}
	
	data, _ := ioutil.ReadAll(r.Body)	// read the response data
	r.Body.Close()
	
	var d map[string]interface{}
	err = json.Unmarshal(data, &d)		// convert the response data to json
	
	if err != nil {
		return nil, err
	}
	
	if _, error := d["errorCode"]; error {	// if an error code is provided in the json something went wrong
		err = errors.New(string(response))
		return nil, err
	}
	
	return d, nil
}

// Convenience method for normal api calls
func (a API) request(client *http.Client, endpoint string, options map[string]string, params url.Values) (j map[string]interface{}, e error)  {
	ep, ok := apiUrls[endpoint]
	if !ok {
		return nil, errors.New("Endpoint \""+endpoint+"\" not defined")
	}
	
	for field, value := range options {
		ep = strings.Replace(ep, ":"+field, value, -1)
	}

	if len(params) > 0 {
		ep += "?"+params.Encode()
	}
	
	u, err := url.Parse(ep)
	if err != nil {
		return nil, err
	}
	
	return a.Raw(client, *u)
}
