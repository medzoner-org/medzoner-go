package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	mocks "github.com/Medzoner/medzoner-go/test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/golang/mock/gomock"
	"github.com/Medzoner/medzoner-go/internal/wire"
	"github.com/Medzoner/medzoner-go/test/features/bootstrap"
)

var opt = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "progress",
}

func init() {
	godog.BindCommandLineFlags("godog.", &opt)
}

func TestFeatures(t *testing.T) {
	if os.Getenv("GODOG_INTEGRATION") == "" {
		t.Skip("Skipping godog tests (set GODOG_INTEGRATION=1 to run)")
		return
	}

	mocked := mocks.New(t)
	mocked.ContactRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mocked.Mailer.EXPECT().Send(gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
	mocked.TechnoRepository.EXPECT().FetchStack(gomock.Any()).Return(map[string]interface{}{}, nil).AnyTimes()

	t.Setenv("APP_ENV", "test")
	t.Setenv("DEBUG", "true")
	t.Setenv("SERVER_PORT", "9876")
	t.Setenv("ROOT_PATH", "./")

	srv, err := wire.InitServerTest(context.Background(), mocked)
	if err != nil {
		t.Fatalf("failed to init server: %v", err)
	}

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Serve(ctx); err != nil {
			fmt.Println("server stopped:", err)
		}
	}()

	// Wait for server to be ready
	baseURL := "http://localhost:9876"
	ready := false
	for i := 0; i < 50; i++ {
		resp, err := http.Get(baseURL + "/healthz/live")
		if err == nil {
			resp.Body.Close()
			ready = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !ready {
		cancel()
		t.Fatal("server did not start in time")
	}

	opts := godog.Options{
		Output:      colors.Colored(os.Stdout),
		Format:      "pretty",
		Paths:       []string{"./test/features/test"},
		Concurrency: 1,
	}
	featureCtx := bootstrap.New(srv, *mocked, baseURL)
	suite := godog.TestSuite{
		Name: "medzoner",
		TestSuiteInitializer: func(suiteContext *godog.TestSuiteContext) {
			featureCtx.InitializeTestSuite(suiteContext)
		},
		ScenarioInitializer: func(scenarioContext *godog.ScenarioContext) {
			featureCtx.InitializeScenario(scenarioContext)
		},
		Options: &opts,
	}

	status := suite.Run()

	cancel()

	if status != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
