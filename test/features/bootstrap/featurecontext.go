package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Medzoner/gomedz/pkg/http/server"
	mocks "github.com/Medzoner/medzoner-go/test"
	"github.com/cucumber/godog"
)

// APIFeature APIFeature
type APIFeature struct {
	Mocks    mocks.Mocks
	Response *http.Response
	headers  map[string]string
	BaseURL  string
	Server   server.Server
}

// New initialize a new APIFeature
func New(srv server.Server, mocked mocks.Mocks, baseURL string) *APIFeature {
	return &APIFeature{
		Response: nil,
		Server:   srv,
		Mocks:    mocked,
		headers:  make(map[string]string),
		BaseURL:  baseURL,
	}
}

// InitializeTestSuite InitializeTestSuite
func (a *APIFeature) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {})
	ctx.AfterSuite(func() {
		if err := a.Server.Shutdown(context.Background()); err != nil {
			fmt.Println(err)
		}
	})
}

// InitializeScenario InitializeScenario
func (a *APIFeature) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, s *godog.Scenario) (context.Context, error) {
		_ = s
		a.resetResponse()
		return ctx, nil
	})
	ctx.Step(`^I add "([^"]*)" header equal to "([^"]*)"$`, a.iAddHeaderEqualTo)
	ctx.Step(`^I send a GET request to "([^"]*)"$`, a.iSendAGETRequestTo)
	ctx.Step(`^I send a POST request to "([^"]*)" with body:$`, a.iSendAPOSTRequestToWithBody)
	ctx.Step(`^the response status code should be (\d+)$`, a.theResponseStatusCodeShouldBe)
	ctx.Step(`^the response body should contain "([^"]*)"$`, a.theResponseBodyShouldContain)
	ctx.Step(`^the response header "([^"]*)" should contain "([^"]*)"$`, a.theResponseHeaderShouldContain)
}

func (a *APIFeature) resetResponse() {
	a.Response = nil
	a.headers = make(map[string]string)
}

func (a *APIFeature) iAddHeaderEqualTo(key, value string) error {
	a.headers[key] = value
	return nil
}

func (a *APIFeature) iSendAGETRequestTo(endpoint string) error {
	req, err := http.NewRequest(http.MethodGet, a.BaseURL+endpoint, nil)
	if err != nil {
		return fmt.Errorf("error creating GET request: %w", err)
	}
	for k, v := range a.headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending GET request: %w", err)
	}
	a.Response = resp
	return nil
}

func (a *APIFeature) iSendAPOSTRequestToWithBody(endpoint string, body *godog.DocString) error {
	var formData url.Values
	var reqBody io.Reader

	contentType := a.headers["Content-Type"]
	if contentType == "application/x-www-form-urlencoded" || contentType == "" {
		formData = url.Values{}
		if body != nil && body.Content != "" {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(body.Content), &data); err == nil {
				for key, value := range data {
					formData.Set(key, fmt.Sprintf("%v", value))
				}
			}
		}
		reqBody = strings.NewReader(formData.Encode())
		if contentType == "" {
			a.headers["Content-Type"] = "application/x-www-form-urlencoded"
		}
	} else {
		if body != nil {
			reqBody = strings.NewReader(body.Content)
		} else {
			reqBody = strings.NewReader("")
		}
	}

	req, err := http.NewRequest(http.MethodPost, a.BaseURL+endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("error creating POST request: %w", err)
	}
	for k, v := range a.headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending POST request: %w", err)
	}
	a.Response = resp
	return nil
}

func (a *APIFeature) theResponseStatusCodeShouldBe(code int) error {
	if a.Response == nil {
		return fmt.Errorf("no response received")
	}
	if a.Response.StatusCode != code {
		body, _ := io.ReadAll(a.Response.Body)
		return fmt.Errorf("expected response code %d, but got %d (body: %s)", code, a.Response.StatusCode, string(body))
	}
	return nil
}

func (a *APIFeature) theResponseBodyShouldContain(expected string) error {
	if a.Response == nil {
		return fmt.Errorf("no response received")
	}
	body, err := io.ReadAll(a.Response.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	a.Response.Body.Close()
	if !strings.Contains(string(body), expected) {
		return fmt.Errorf("expected response body to contain %q, but got: %s", expected, string(body))
	}
	return nil
}

func (a *APIFeature) theResponseHeaderShouldContain(header, expected string) error {
	if a.Response == nil {
		return fmt.Errorf("no response received")
	}
	actual := a.Response.Header.Get(header)
	if !strings.Contains(actual, expected) {
		return fmt.Errorf("expected header %q to contain %q, but got %q", header, expected, actual)
	}
	return nil
}

