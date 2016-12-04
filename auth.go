package sechat

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

var (
	// TODO: ideally these would be replaced with something like goquery to
	// avoid hardcoding the exact format of the HTML elements containing
	// important data
	fkeyRegexp = regexp.MustCompile(`<input type="hidden" id="fkey" name="fkey" value="([0-9a-f-]+)" />`)
	authRegexp = regexp.MustCompile(`<a href="([^"]+)" target="_top">`)

	errFkey = errors.New("unable to find fkey")
	errAuth = errors.New("unable to find auth URL")
	errUrl  = errors.New("unexpected page redirection (bad credentials?)")
)

// Auth contains the information necessary to login to the Stack Exchange chat
// network. Currently, only Stack Exchange credentials are accepted (OpenID is
// not supported yet).
type Auth struct {
	client   http.Client
	email    string
	password string
}

// fetchLoginPage retrieves the URL of the page that contains the login form.
func (a *Auth) fetchLoginURL() (string, error) {
	request, err := http.NewRequest(
		http.MethodGet, "https://stackexchange.com/users/signin", nil,
	)
	if err != nil {
		return "", err
	}
	response, err := a.client.Do(request)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// fetchFkey retrieves the fkey from the login form so that it can be submitted
// during the login process.
func (a *Auth) fetchFkey(url string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	response, err := a.client.Do(request)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	matches := fkeyRegexp.FindSubmatch(b)
	if len(matches) != 2 {
		return "", errFkey
	}
	return string(matches[1]), nil
}

// submitLoginForm submits the authentication information along with the fkey.
// A URL is returned which is necessary to complete the login process.
func (a *Auth) submitLoginForm(fkey string) (string, error) {
	form := &url.Values{}
	form.Set("email", a.email)
	form.Set("password", a.password)
	form.Set("affId", "11")
	form.Set("fkey", fkey)
	request, err := http.NewRequest(
		http.MethodPost,
		"https://openid.stackexchange.com/affiliate/form/login/submit",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := a.client.Do(request)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	matches := authRegexp.FindSubmatch(b)
	if len(matches) != 2 {
		return "", errAuth
	}
	return string(matches[1]), nil
}

// completeLogin finishes the login process.
func (a *Auth) completeLogin(url string) error {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := a.client.Do(request)
	if err != nil {
		return err
	}
	if response.Request.URL.Path != "/" {
		return errUrl
	}
	return nil
}

// NewAuth creates an object that can be used to issue authenticated requests
// against the Stack Exchange servers.
func NewAuth(email, password string) (*Auth, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &Auth{
		client: http.Client{
			Jar: jar,
		},
		email:    email,
		password: password,
	}, nil
}

// Login performs the sequence of steps necessary to authenticate with the
// Stack Exchange servers.
func (a *Auth) Login() error {
	loginURL, err := a.fetchLoginURL()
	if err != nil {
		return err
	}
	fkey, err := a.fetchFkey(loginURL)
	if err != nil {
		return err
	}
	authURL, err := a.submitLoginForm(fkey)
	if err != nil {
		return err
	}
	return a.completeLogin(authURL)
}
