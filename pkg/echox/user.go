package echox

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net/http"
	"net/url"
)

type UserInfo struct {
	Subject string `json:"sub"`
	Name    string `json:"name,omitempty"`
	Raw     string `json:"-"`
}

func (api Auth0Api) UserInfo(token *jwt.Token) (UserInfo, error) {
	if u, ok := api.cache.Get(token); ok {
		return u, nil
	}
	u, err := api.FetchUserInfo(token)
	if err != nil {
		return u, err
	}
	api.cache.Put(token, u)
	return u, nil
}

func (api Auth0Api) FetchUserInfo(token *jwt.Token) (UserInfo, error) {
	requestUrl, _ := url.Parse(fmt.Sprintf("%suserinfo", api.Tenant))
	req := &http.Request{
		Method: "GET",
		URL:    requestUrl,
		Header: map[string][]string{
			"Authorization": {"Bearer " + token.Raw},
		},
	}
	resp, err := http.DefaultClient.Do(req)
	var user UserInfo
	if err == nil {
		//err = json.NewDecoder(resp.Body).Decode(&user)
		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return user, err
		}
		err = json.Unmarshal(respBytes, &user)
		user.Raw = string(respBytes)
	}
	return user, err
}
