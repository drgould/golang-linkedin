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

var apiRoot = "https://api.linkedin.com"
var apiUser = "/v1/people/{id}"
var apiGroup = "/v1/groups/{id}"

var apiUrls = map[string]interface{}{
	"profile": apiUser+"{fields}",
	"connections": apiUser+"/connections{fields}{",
	"group": apiGroup+"{fields}",
}

type User struct {
	user_id string
}

type API struct {
	oauth_key string			// your oauth key
	oauth_secret string			// your oauth secret 
	access_token string			// the user's access token
}

func (a *API) SetCredentials(key string, secret string) {
	a.oauth_key = key
	a.oauth_secret = secret
}

func (a *API) SetToken(token string) {
	a.access_token = token
}

func (a API) GetToken() (t string) {
	return a.access_token
}

func (a API) AuthUrl(state string, redirect_url string) (url string) {
	return "https://www.linkedin.com/uas/oauth2/authorization?response_type=code&client_id="+a.oauth_key+
								 "&state="+state+"&redirect_uri="+redirect_url
}

func (a API) Auth(w http.ResponseWriter, r *http.Request, state string, redirect_url string) {
	http.Redirect(w, r, a.AuthUrl(state, redirect_url), http.StatusFound)
}

func (a *API) RetrieveAccessToken(client *http.Client, code string, redirect_url string) (t string, e error) {
	
	resp, err := client.Get("https://www.linkedin.com/uas/oauth2/accessToken?grant_type=authorization_code&code="+code+"&redirect_uri="+
						redirect_url+"&client_id="+a.oauth_key+"&client_secret="+a.oauth_secret)
	
	if err != nil {
		return t, err
	}
	
	token, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	
	var response map[string]interface{}
	err = json.Unmarshal(token, &response)
	
	if err != nil {
		return t, err
	}
	
	if _, err := response["error"]; err {
		return t, errors.New(response["error"].(string)+" - "+response["error_description"].(string))
	}
	
	t = response["access_token"].(string)
	
	a.SetToken(t)
	
	return t, nil
}

func getUserIdString(id string) (uid string) {
	if id == "~" {
		return "~"
	} else if strings.Contains(id, "http") {
		return "url="+url.QueryEscape(id)
	} else {
		return "id="+id
	}
}

func getGroupIdString(id interface{}) (gid string, err error) {
	switch t := id.(type) {
		case string:
			if strings.Contains(id.(string), "http") {
				return "url="+url.QueryEscape(id.(string)), nil
			}
			return id.(string), nil
		case uint64:
			return strconv.FormatUint(id.(uint64),10), nil
		default:
			return gid, errors.New(fmt.Sprintf("Group ID type exception: Expecting string or uint64 got %#v", t))
	}
}

func getFieldsString(fields map[string]interface{}) (f string) {
	if len(fields) > 0 {
		str := ":("
		comma := ""
		for field, subfields := range fields {
			str += comma + field
			switch t := subfields.(type) {
				case []string:
					str += ":("
					subcomma := ""
					for _, subfield := range subfields.([]string) {
						str += subcomma + subfield
						subcomma = ","
					}
					str += ")"
			}
			comma = ","
		}
		return str + ")"
	}
	return ""
}

func (a API) Profile(client *http.Client, user_id string, fields map[string]interface{}) (j map[string]interface{}, err error) {
	return a.request(client, "profile", map[string]string{
		"id": getUserIdString(user_id),
		"fields": getFieldsString(fields),
	}, url.Values{})
}

func (a API) Connections(client *http.Client, user_id string, fields map[string]interface{}, params map[string]interface{}) (j map[string]interface{}, err error) {
	return a.request(client, "connections", map[string]string{
		"id": getUserIdString(user_id),
		"fields": getFieldsString(fields),
	}, map[string]string{})
}

func (a API) Group(client *http.Client, group_id interface{}, fields map[string]interface{}) (j map[string]interface{}, err error) {
	gid, err := getGroupIdString(group_id)
	if err != nil {
		return j, err
	}
	
	return a.request(client, "group", map[string]string{
		"id": gid,
		"fields": getFieldsString(fields),
	}, map[string]string{})
}

func (a API) Raw(client *http.Client, u interface{}) (j map[string]interface{}, e error) {
	var endpoint url.URL
	
	switch t := u.(type) {
		default:
			return nil, errors.New(fmt.Sprintf("Expecting a string or url.URL, got %v instead", t))
		case string:
			endpoint, err := url.Parse(u.(string))
			if err != nil {
				return nil, err
			}
		case url.URL:
			endpoint = u.(url.URL)
	}
	
	qs := endpoint.Query()
	qs.Add("oauth2_access_token",a.access_token)
	
	req, _ := http.NewRequest("GET", apiRoot+endpoint.Path+"?"+qs.Encode(), nil)
	req.URL.Opaque = endpoint.Path
	req.Header.Add("x-li-format","json")
	
	r, err := client.Do(req)
	
	if err != nil {
		return nil, err
	}
	
	response, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	
	var d map[string]interface{}
	err = json.Unmarshal(response.([]byte), &d)
	
	if err != nil {
		return d, err
	}
	
	if _, error := d["errorCode"]; error {
		err = errors.New(string(response))
		return d, err
	}
	
	return d, nil
}

func (a API) request(client *http.Client, endpoint string, options map[string]string, params url.Values) (j map[string]interface{}, e error)  {
	
	if endpoint, err := apiUrls[endpoint]; err {
		return j, errors.New("Endpoint \""+endpoint+"\" not defined")
	}
	
	for field, value := range options {
		endpoint = strings.Replace(endpoint, "{"+field+"}", value, 1)
	}
	
	u := url.Parse(endpoint)
	
	u.Query = params.Encode()
	
	return a.Raw(client, u)
}
