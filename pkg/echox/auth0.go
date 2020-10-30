package echox

import (
	"encoding/json"
	"errors"
	"fmt"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

type UserInfo struct {
	Subject string `json:"sub"`
	Name    string `json:"name,omitempty"`
	Raw     string `json:"-"`
}

type Auth0Api struct {
	Tenant   string
	Audience string
}

func (api Auth0Api) Middleware() func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
	auth0Middleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {

			// Verify 'aud' claim
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(api.Audience, false)
			if !checkAud {
				return token, errors.New("Invalid audience.")
			}
			// Verify 'iss' claim
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(api.Tenant, false)
			if !checkIss {
				return token, errors.New("Invalid issuer.")
			}

			cert, err := api.PemCert(token)
			if err != nil {
				panic(err.Error())
			}

			result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			return result, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
	})

	return func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := auth0Middleware.CheckJWT(c.Response().Writer, c.Request())
			if err != nil {
				c.Logger().Error("JWT check failed")
				return err
			}

			return handlerFunc(c)
		}
	}
}

func (api Auth0Api) ContextToken(c echo.Context) *jwt.Token {
	return c.Request().Context().Value("user").(*jwt.Token)
}

func (api Auth0Api) ContextUserInfo(c echo.Context) (UserInfo, error) {
	return api.UserInfo(api.ContextToken(c))
}

func (api Auth0Api) UserInfo(token *jwt.Token) (UserInfo, error) {
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

func (api Auth0Api) PemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get(fmt.Sprintf("%s.well-known/jwks.json", api.Tenant))

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k, _ := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}
