package echox

import (
	"encoding/json"
	"errors"
	"fmt"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"net/http"
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

type Auth0Api struct {
	Tenant   string
	Audience string
	cache    *cache
}

func New(tenant, audience string) Auth0Api {
	return Auth0Api{
		Tenant:   tenant,
		Audience: audience,
		cache:    newCache(1000),
	}
}

func (api Auth0Api) SetCacheMaxAge(age int32) {
	api.cache.maxAgeMs = age
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
				return echo.ErrUnauthorized
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
