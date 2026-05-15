package bootstrap

import (
	"context"

	"github.com/Medzoner/gomedz/pkg/http/bddtesting"
	"github.com/Medzoner/gomedz/pkg/http/server"
	mocks "github.com/Medzoner/medzoner-go/test"
	"github.com/cucumber/godog"
)

// APIFeature embeds the generic bddtesting.APIFeature and adds project-specific mocks.
type APIFeature struct {
	*bddtesting.APIFeature
	Mocks  mocks.Mocks
	Server server.Server
}

// New initializes a new APIFeature with mocks and server shutdown wired in.
func New(srv server.Server, mocked mocks.Mocks, baseURL string) *APIFeature {
	feat := bddtesting.New(baseURL)
	feat.SetShutdownFunc(func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	})

	return &APIFeature{
		APIFeature: feat,
		Mocks:      mocked,
		Server:     srv,
	}
}

// InitializeTestSuite delegates to the embedded APIFeature.
func (a *APIFeature) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	a.APIFeature.InitializeTestSuite(ctx)
}

// InitializeScenario registers all built-in steps from the embedded APIFeature.
// Add project-specific steps here if needed.
func (a *APIFeature) InitializeScenario(ctx *godog.ScenarioContext) {
	a.APIFeature.InitializeScenario(ctx)
}
