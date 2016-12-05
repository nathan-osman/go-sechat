package sechat

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	errNetworkFkey = errors.New("unable to find network fkey")
	errChatFkey    = errors.New("unable to find chat fkey")
	errAuth        = errors.New("unable to find auth URL")
	errUrl         = errors.New("incomplete login (invalid credentials)")
)

// Auth contains the information necessary to login to the Stack Exchange chat
// network. Currently, only Stack Exchange credentials are accepted (OpenID is
// not supported yet).
type Auth struct {
	client *http.Client
	state  *AuthState
}

// AuthState maintains the information both to establish an authenticated
// connection and restore an existing session that has been serialized.
type AuthState struct {
	Email    string
	Password string
	Fkey     string
	Cookies  []*http.Cookie
}

// fetchLoginPage retrieves the URL of the page that contains the login form.
func (a *Auth) fetchLoginURL() (string, error) {
	req, err := http.NewRequest(
		http.MethodGet, "https://stackexchange.com/users/signin", nil,
	)
	if err != nil {
		return "", err
	}
	res, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// fetchNetworkFkey retrieves the network fkey from the login form so that it
// can be submitted during the login process.
func (a *Auth) fetchNetworkFkey(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	res, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}
	fkey, ok := doc.Find("#fkey").Attr("value")
	if !ok {
		return "", errNetworkFkey
	}
	return fkey, nil
}

// fetchAuthURL exists due to a bug in golang.org/x/net/html that prevents
// noscript elements from being parsed correctly. This method works around it.
func (a *Auth) fetchAuthURL(res *http.Response) (string, error) {
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return "", err
	}
	noscriptDoc, err := goquery.NewDocumentFromReader(
		strings.NewReader(doc.Find("noscript").Text()),
	)
	if err != nil {
		return "", err
	}
	url, ok := noscriptDoc.Find("a").Attr("href")
	if !ok {
		return "", errAuth
	}
	return url, nil
}

// submitLoginForm submits the authentication information along with the fkey.
// A URL is returned which is necessary to complete the login process.
func (a *Auth) submitLoginForm(fkey string) (string, error) {
	form := &url.Values{}
	form.Set("email", a.state.Email)
	form.Set("password", a.state.Password)
	form.Set("affId", "11")
	form.Set("fkey", fkey)
	req, err := http.NewRequest(
		http.MethodPost,
		"https://openid.stackexchange.com/affiliate/form/login/submit",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	// fetchAuthURL ought to be part of this method but please see its
	// docstring for an explanation of why it isn't
	return a.fetchAuthURL(res)
}

// completeLogin finishes the login process.
func (a *Auth) completeLogin(authUrl string) error {
	req, err := http.NewRequest(http.MethodGet, authUrl, nil)
	if err != nil {
		return err
	}
	res, err := a.client.Do(req)
	if err != nil {
		return err
	}
	if res.Request.URL.Path != "/" {
		return errUrl
	}
	return nil
}

// fetchChatFkey loads the home page for chat in order to retrieve the fkey
// that is required to accompany every authenticated request.
func (a *Auth) fetchChatFkey() (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://chat.stackexchange.com",
		nil,
	)
	if err != nil {
		return "", err
	}
	res, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}
	fkey, ok := doc.Find("#fkey").Attr("value")
	if !ok {
		return "", errChatFkey
	}
	return fkey, nil
}

// NewAuth creates an object that can be used to issue authenticated requests
// against the Stack Exchange servers.
func NewAuth(state *AuthState) (*Auth, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse("https://chat.stackexchange.com")
	if err != nil {
		return nil, err
	}
	jar.SetCookies(u, state.Cookies)
	return &Auth{
		client: &http.Client{
			Jar: jar,
		},
		state: state,
	}, nil
}

// IsLoggedIn determines if the current state contains login information. Note
// that the information is not checked to be valid.
func (a *Auth) IsLoggedIn() bool {
	return len(a.state.Fkey) != 0 && len(a.state.Cookies) != 0
}

// Login performs the sequence of steps necessary to authenticate with the
// Stack Exchange servers.
func (a *Auth) Login() error {
	loginURL, err := a.fetchLoginURL()
	if err != nil {
		return err
	}
	networkFkey, err := a.fetchNetworkFkey(loginURL)
	if err != nil {
		return err
	}
	authURL, err := a.submitLoginForm(networkFkey)
	if err != nil {
		return err
	}
	if err := a.completeLogin(authURL); err != nil {
		return err
	}
	chatFkey, err := a.fetchChatFkey()
	if err != nil {
		return err
	}
	a.state.Fkey = chatFkey
	return nil
}

// State returns the current login state in preparation for serialization.
func (a *Auth) State() (*AuthState, error) {
	u, err := url.Parse("https://chat.stackexchange.com")
	if err != nil {
		return nil, err
	}
	a.state.Cookies = a.client.Jar.Cookies(u)
	return a.state, nil
}
