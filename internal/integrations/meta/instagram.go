package meta

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type urlParams = map[string][]string

type ContainerResp struct {
	Id string `json:"id"`
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type TokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type TokenInspect struct {
	Data struct {
		AppId       string `json:"app_id"`
		Type        string `json:"type"`
		Application string `json:"application"`
		ExpiresAt   int    `json:"expires_at"`
		IsValid     bool   `json:"is_valid"`
		IssuedAt    int    `json:"issued_at"`
		Metadata    struct {
			Sso string `json:"sso"`
		} `json:"metadata"`
		Scopes []string `json:"scopes"`
		UserId string   `json:"user_id"`
	} `json:"data"`
}

type BusAccount struct {
	InstagramBusinessAccount struct {
		Id string `json:"id"`
	} `json:"instagram_business_account"`
	Id string `json:"id"`
}

type Details struct {
	Data []struct {
		Name         string   `json:"name"`
		Category     string   `json:"category"`
		Id           string   `json:"id"`
		AccessToken  string   `json:"access_token"`
		Tasks        []string `json:"tasks"`
		CategoryList []struct {
			Name string `json:"name"`
			Id   string `json:"id"`
		} `json:"category_list"`
	} `json:"data"`
}

type LongToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

const (
	privacyPolicy = `Privacy Policy 

xbd.au is committed to providing quality services to you and this policy outlines our ongoing obligations to you in respect of how we manage your Personal Information.

We have adopted the Australian Privacy Principles (APPs) contained in the Privacy Act 1988 (Cth) (the Privacy Act). The NPPs govern the way in which we collect, use, disclose, store, secure and dispose of your Personal Information.

A copy of the Australian Privacy Principles may be obtained from the website of The Office of the Australian Information Commissioner at https://www.oaic.gov.au/.

What is Personal Information and why do we collect it?

Personal Information is information or an opinion that identifies an individual. Examples of Personal Information we collect includes names, addresses, email addresses, phone and facsimile numbers.

This Personal Information is obtained in many ways including from website. We don’t guarantee website links or policy of authorised third parties.

We collect your Personal Information for the primary purpose of providing our services to you, providing information to our clients and marketing. We may also use your Personal Information for secondary purposes closely related to the primary purpose, in circumstances where you would reasonably expect such use or disclosure. You may unsubscribe from our mailing/marketing lists at any time by contacting us in writing.

When we collect Personal Information we will, where appropriate and where possible, explain to you why we are collecting the information and how we plan to use it.

Sensitive Information

Sensitive information is defined in the Privacy Act to include information or opinion about such things as an individual's racial or ethnic origin, political opinions, membership of a political association, religious or philosophical beliefs, membership of a trade union or other professional body, criminal record or health information.

Sensitive information will be used by us only:

•	For the primary purpose for which it was obtained

•	For a secondary purpose that is directly related to the primary purpose

•	With your consent; or where required or authorised by law.

Third Parties

Where reasonable and practicable to do so, we will collect your Personal Information only from you. However, in some circumstances we may be provided with information by third parties. In such a case we will take reasonable steps to ensure that you are made aware of the information provided to us by the third party.

Disclosure of Personal Information

Your Personal Information may be disclosed in a number of circumstances including the following:

•	Third parties where you consent to the use or disclosure; and

•	Where required or authorised by law.

Security of Personal Information

Your Personal Information is stored in a manner that reasonably protects it from misuse and loss and from unauthorized access, modification or disclosure.

When your Personal Information is no longer needed for the purpose for which it was obtained, we will take reasonable steps to destroy or permanently de-identify your Personal Information. However, most of the Personal Information is or will be stored in client files which will be kept by us for a minimum of 7 years.

Access to your Personal Information

You may access the Personal Information we hold about you and to update and/or correct it, subject to certain exceptions. If you wish to access your Personal Information, please contact us in writing.

xbd.au will not charge any fee for your access request, but may charge an administrative fee for providing a copy of your Personal Information.

In order to protect your Personal Information we may require identification from you before releasing the requested information.

Maintaining the Quality of your Personal Information

It is an important to us that your Personal Information is up to date. We  will  take reasonable steps to make sure that your Personal Information is accurate, complete and up-to-date. If you find that the information we have is not up to date or is inaccurate, please advise us as soon as practicable so we can update our records and ensure we can continue to provide quality services to you.

Policy Updates

This Policy may change from time to time and is available on our website.

Privacy Policy Complaints and Enquiries

If you have any queries or complaints about our Privacy Policy please contact us at:

mail@xbd.au
`
)

var (
	appId       = os.Getenv("IG_APP_ID")
	appSecret   = os.Getenv("IG_SECRET")
	callbackUri = "https://weight.xbd.au/new-token"
)

func DoReq[T any](method, baseUri string, params urlParams) (T, error) {
	client := &http.Client{}
	var t T

	req, err := http.NewRequest(method, baseUri, nil)
	q := req.URL.Query()
	for key, valList := range params {
		for _, val := range valList {
			q.Add(key, val)
		}
	}

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return t, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return t, err
	}

	if resp.StatusCode != http.StatusOK {
		return t, fmt.Errorf("error from %s, %d: %s", baseUri, resp.StatusCode, body)
	}

	err = json.Unmarshal(body, &t)
	if err != nil {
		return t, err
	}

	return t, nil
}

func GetReq[T any](baseUri string, params map[string][]string) (T, error) {
	return DoReq[T](http.MethodGet, baseUri, params)
}

func PostReq[T any](baseUri string, params map[string][]string) (T, error) {
	return DoReq[T](http.MethodPost, baseUri, params)
}

func GetAccessToken() (string, error) {
	baseUri := "https://graph.facebook.com/oauth/access_token"
	params := urlParams{
		"client_id":     []string{appId},
		"client_secret": []string{appSecret},
		"grant_type":    []string{"client_credentials"},
		"redirect_uri":  []string{callbackUri},
	}

	access, err := GetReq[AccessToken](baseUri, params)
	if err != nil {
		return "", err
	}

	return access.AccessToken, nil
}

func GetUserId(accessToken string) (string, error) {
	access, err := GetAccessToken()
	if err != nil {
		return "", err
	}

	baseUri := "https://graph.facebook.com/debug_token"
	params := urlParams{
		"input_token":  []string{accessToken},
		"access_token": []string{access},
	}

	inspect, err := GetReq[TokenInspect](baseUri, params)
	if err != nil {
		return "", err
	}

	return inspect.Data.UserId, nil
}

func GetDetails(accessToken string) (string, string, error) {
	userId, err := GetUserId(accessToken)
	if err != nil {
		return "", "", err
	}

	baseUri := fmt.Sprintf("https://graph.facebook.com/v17.0/%s/accounts", userId)
	params := urlParams{
		"access_token": []string{accessToken},
	}

	details, err := GetReq[Details](baseUri, params)
	if err != nil {
		return "", "", err
	}

	for _, page := range details.Data {
		if page.Name == "Blw" {
			return page.Id, page.AccessToken, nil
		}
	}

	return "", "", fmt.Errorf("no page found")
}

func CreateContainer(igId, imgAddr, caption, accessToken string) (string, error) {
	baseUri := fmt.Sprintf("https://graph.facebook.com/v17.0/%s/media", igId)
	params := urlParams{
		"image_url":    []string{imgAddr},
		"caption":      []string{caption},
		"access_token": []string{accessToken},
	}

	cr, err := PostReq[ContainerResp](baseUri, params)
	if err != nil {
		return "", err
	}

	return cr.Id, nil
}

func PublishContent(igId, containerId, accessToken string) error {
	baseUri := fmt.Sprintf("https://graph.facebook.com/v17.0/%s/media_publish", igId)
	params := urlParams{
		"creation_id":  []string{containerId},
		"access_token": []string{accessToken},
	}

	_, err := PostReq[interface{}](baseUri, params)
	if err != nil {
		return err
	}

	return nil
}

func AuthUrl() string {
	return fmt.Sprintf("https://www.facebook.com/v17.0/dialog/oauth?client_id=%s&redirect_uri=%s&state={false=true}&scope=instagram_basic,pages_show_list,business_management,instagram_basic,instagram_content_publish&request_", appId, callbackUri)
}

func GetToken(code string) (string, error) {
	baseUri := "https://graph.facebook.com/v17.0/oauth/access_token"
	params := urlParams{
		"client_id":     []string{appId},
		"redirect_uri":  []string{callbackUri},
		"client_secret": []string{appSecret},
		"code":          []string{code},
	}

	tokenResp, err := GetReq[TokenResp](baseUri, params)
	if err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func BusinessAccount(pageId, accessToken string) (string, error) {
	baseUri := fmt.Sprintf("https://graph.facebook.com/v17.0/%s", pageId)
	params := urlParams{
		"fields":       []string{"instagram_business_account"},
		"access_token": []string{accessToken},
	}

	account, err := GetReq[BusAccount](baseUri, params)
	if err != nil {
		return "", err
	}

	return account.InstagramBusinessAccount.Id, nil
}

func GetLongToken(token string) (string, error) {
	baseUri := "https://graph.facebook.com/v17.0/oauth/access_token"
	params := urlParams{
		"grant_type":        []string{"fb_exchange_token"},
		"client_id":         []string{appId},
		"client_secret":     []string{appSecret},
		"fb_exchange_token": []string{token},
	}

	longToken, err := GetReq[LongToken](baseUri, params)
	if err != nil {
		return "", err
	}

	return longToken.AccessToken, nil
}

func Policy() string {
	return privacyPolicy
}
