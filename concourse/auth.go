package concourse

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
)

//go:generate counterfeiter . TokenProvider
type TokenProvider interface {
	GetAuthorizationHeader() (string, error)
}

type tokenProvider struct {
	host     string
	team     string
	username string
	password string

	clock clock.Clock

	cachedAuthHeader string
	lastTokenRequest time.Time
}

func NewTokenProvider(host, team, username, password string, clock clock.Clock) TokenProvider {
	return &tokenProvider{
		host:     strings.TrimSuffix(host, "/"),
		team:     team,
		username: username,
		password: password,
		clock:    clock,
	}
}

func (t *tokenProvider) GetAuthorizationHeader() (string, error) {
	if t.cachedAuthHeader == "" || t.clock.Since(t.lastTokenRequest) > 23*time.Hour {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/teams/%s/auth/token", t.host, t.team), nil)
		if err != nil {
			return "", err
		}
		req.SetBasicAuth(t.username, t.password)

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client := &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		}
		res, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()

		var authResponse struct {
			Type  string
			Value string
		}
		err = json.NewDecoder(res.Body).Decode(&authResponse)
		if err != nil {
			return "", err
		}

		t.cachedAuthHeader = fmt.Sprintf("%s %s", authResponse.Type, authResponse.Value)
		t.lastTokenRequest = t.clock.Now()
	}

	return t.cachedAuthHeader, nil
}
