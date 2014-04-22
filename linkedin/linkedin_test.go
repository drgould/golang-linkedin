package linkedin

import (
	"testing"
)

func TestAuthUrl(t *testing.T) {
	var li API
	
	li.SetApiCredentials("aaa","bbb")
	
	assert.Equal(t, li.AuthUrl("ccc","localhost"), 
		"https://www.linkedin.com/uas/oauth2/authorization?response_type=code&client_id=aaa&state=ccc&redirect_uri=localhost")
}



