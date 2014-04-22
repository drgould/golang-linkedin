package linkedin

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"bytes"
	"io/ioutil"
)

func TestSetCredentials(t *testing.T) {
	var li API

	li.SetCredentials("a", "b")

	if li.oauth_key != "a" {
		t.FailNow()
	}
	if li.oauth_secret != "b" {
		t.FailNow()
	}
}

func TestAuthUrl(t *testing.T) {
	var li API
	
	li.SetCredentials("**a**","**b**")

	if li.AuthUrl("**c**","**d**") != "https://www.linkedin.com/uas/oauth2/authorization?response_type=code&client_id=**a**&state=**c**&redirect_uri=**d**" {
		t.Fatal("url is not correct")
	}
}

func TestAuthRedirect(t *testing.T) {
	var li API

	li.SetCredentials("a","b")

	r, err := http.NewRequest("GET", "http://a.com/blah", nil)
	if err != nil {
		t.Fatal("couldn't create request")
	}

	w := httptest.NewRecorder()

	li.Auth(w, r, "**c**", "**d**")

	if w.Code != http.StatusFound {
		t.Fatal("not redirected")
	}
}

type reqCheck func(*http.Request)

func createResponder(rc reqCheck, response string) (Responder) {
	return Responder(func (req *http.Request) (*http.Response, error) {
		if rc != nil {
			rc(req)
		}

		resp := &http.Response{
			StatusCode: 200,
			ProtoMajor: 1,
			ProtoMinor: 0,
			Body: ioutil.NopCloser(bytes.NewBufferString(response)),
			ContentLength: int64(len(response)),
			Request: req,
		}

		resp.Header = make(map[string][]string)
		resp.Header.Add("Content-Type","application/json")

		return resp, nil
	})
}

func TestRetrieveAccessToken(t *testing.T) {
	var li API


	key := "key"
	secret := "secret"
	authcode := "authcode"
	redirect := "redirect"

	RegisterResponder("GET",
		"https://www.linkedin.com/uas/oauth2/accessToken?client_id="+key+"&client_secret="+secret+"&code="+authcode+
			"&grant_type=authorization_code&redirect_uri="+redirect,
		createResponder(
			reqCheck(func(req *http.Request) {
				u := req.URL
				qs := u.Query()

				if code, ok := qs["code"]; ok {
					if code[0] != authcode {
						t.Fatal("code is not '"+authcode+"'")
					}
				} else {
					t.Fatal("code not given")
				}
			}),
			`{"expires_in":5184000,"access_token":"TOKEN"}`))
	Activate(false)

	li.SetCredentials(key, secret)

	token, err := li.RetrieveAccessToken(http.DefaultClient, authcode, redirect)

	if err != nil {
		t.Fatal(err)
	}

	if token != "TOKEN" {
		t.Fatalf("invalid token returned: %v", token)
	}

	if li.access_token != "TOKEN" {
		t.Fatalf("token set incorrectly: %v", li.access_token)
	}
}

func TestUserId(t *testing.T) {
	if getUserIdString("~") != "~" {
		t.Fatal("'~' invalid")
	}
	if getUserIdString("") != "~" {
		t.Fatal("'' invalid")
	}
	if getUserIdString("id") != "id=id" {
		t.Fatal("'id' invalid")
	}
	if getUserIdString("http://blah.com/blah") != "url=http%3A%2F%2Fblah.com%2Fblah" {
		t.Fatal("'http://blah.com/blah' invalid")
	}
}

func TestGroupId(t *testing.T) {
	id, err := getGroupIdString(uint64(1234))
	if err != nil {
		t.Fatal(err)
	}
	if id != "1234" {
		t.Fatal("'1234' invalid")
	}
	id, err = getGroupIdString("http://blah.com/blah")
	if err != nil {
		t.Fatal(err)
	}
	if id != "url=http%3A%2F%2Fblah.com%2Fblah" {
		t.Fatal("'http://blah.com/blah' invalid")
	}
}

func TestUserProfile(t *testing.T) {
	var li API

	token := "abcde"

	RegisterResponder(
		"GET",
		"https://api.linkedin.com/v1/people/~:(id)?oauth2_access_token="+token,
		createResponder(nil,`{"id":"USERID"}`))
	Activate(false)

	fields := Fields{}
	fields.Add("id")

	li.SetToken(token)

	data, err := li.Profile(http.DefaultClient, "~", fields)

	if err != nil {
		t.Fatal(err)
	}

	uid, ok := data["id"]
	if !ok {
		t.Fatalf("id not returned: %#v", data)
	}

	if uid != "USERID" {
		t.Fatalf("invalid id: %v", uid)
	}

}
