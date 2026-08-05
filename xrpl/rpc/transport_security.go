package rpc

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

const authorizationHeader = "Authorization"

// validateAuthorizationTransport enforces the authorization transport policy
// for rawURL and headers, and returns the HTTP client that requests carrying
// authorization must use. It is the single entry point shared by config-time
// and request-time validation.
func validateAuthorizationTransport(rawURL string, headers map[string][]string, client HTTPClient) (HTTPClient, error) {
	hasAuthorization, err := validateAuthorizationEndpoint(rawURL, headers)
	if err != nil {
		return nil, err
	}
	return authorizationHTTPClient(client, hasAuthorization)
}

func validateAuthorizationEndpoint(rawURL string, headers map[string][]string) (bool, error) {
	hasHeader := hasAuthorizationHeader(headers)
	endpoint, err := url.Parse(rawURL)
	if err != nil {
		hasAuthorization := hasHeader || malformedURLHasUserinfo(rawURL)
		if hasAuthorization {
			return true, ErrInsecureAuthorization
		}
		return false, nil
	}

	hasAuthorization := hasHeader || endpoint.User != nil
	if hasAuthorization && !isHTTPSURL(endpoint) {
		return true, ErrInsecureAuthorization
	}
	return hasAuthorization, nil
}

func hasAuthorizationHeader(headers map[string][]string) bool {
	for name := range headers {
		if strings.EqualFold(name, authorizationHeader) {
			return true
		}
	}
	return false
}

func isHTTPSURL(endpoint *url.URL) bool {
	return endpoint != nil && endpoint.IsAbs() && endpoint.Host != "" && strings.EqualFold(endpoint.Scheme, "https")
}

func authorizationHTTPClient(client HTTPClient, hasAuthorization bool) (HTTPClient, error) {
	if !hasAuthorization {
		return client, nil
	}

	baseClient, ok := client.(*http.Client)
	if !ok || baseClient == nil {
		return nil, ErrInsecureAuthorization
	}

	return &http.Client{
		Transport:     baseClient.Transport,
		CheckRedirect: authorizationRedirectPolicy(baseClient.CheckRedirect),
		Jar:           baseClient.Jar,
		Timeout:       baseClient.Timeout,
	}, nil
}

func authorizationRedirectPolicy(previousCheckRedirect func(*http.Request, []*http.Request) error) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if !isHTTPSURL(req.URL) && redirectChainHasAuthorization(req, via) {
			return ErrInsecureAuthorization
		}

		if previousCheckRedirect != nil {
			if err := previousCheckRedirect(req, via); err != nil {
				return err
			}
			// Check again because a caller callback may mutate the redirect target.
			if !isHTTPSURL(req.URL) && redirectChainHasAuthorization(req, via) {
				return ErrInsecureAuthorization
			}
		} else if len(via) >= 10 {
			return errTooManyRedirects
		}

		return nil
	}
}

func redirectChainHasAuthorization(req *http.Request, via []*http.Request) bool {
	if requestHasAuthorization(req) {
		return true
	}
	return slices.ContainsFunc(via, requestHasAuthorization)
}

func requestHasAuthorization(req *http.Request) bool {
	return req != nil && (hasAuthorizationHeader(req.Header) || req.URL != nil && req.URL.User != nil)
}

func redactAuthorizationError(err error, rawURL string, headers map[string][]string) error {
	if err == nil {
		return nil
	}

	var secrets []string
	for name, values := range headers {
		if strings.EqualFold(name, authorizationHeader) {
			secrets = append(secrets, values...)
		}
	}
	endpoint, parseErr := url.Parse(rawURL)
	if parseErr != nil {
		if malformedURLHasUserinfo(rawURL) {
			return ErrAuthorizationRequestFailed
		}
	} else if endpoint.User != nil {
		username := endpoint.User.Username()
		password, _ := endpoint.User.Password()
		secrets = append(secrets,
			username,
			password,
			endpoint.User.String(),
			"Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)),
		)
	}

	errorMessage := err.Error()
	for _, secret := range secrets {
		if secret != "" && strings.Contains(errorMessage, secret) {
			return ErrAuthorizationRequestFailed
		}
	}
	return err
}

func malformedURLHasUserinfo(rawURL string) bool {
	authority := rawURL
	if strings.HasPrefix(authority, "//") {
		authority = authority[2:]
	} else {
		var found bool
		_, authority, found = strings.Cut(authority, "://")
		if !found {
			return false
		}
	}
	if end := strings.IndexAny(authority, "/?#"); end >= 0 {
		authority = authority[:end]
	}
	return strings.Contains(authority, "@")
}
