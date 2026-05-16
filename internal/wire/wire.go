//go:build wireinject

package wire

import (
	"github.com/Medzoner/medzoner-go/internal/application/command"
	event2 "github.com/Medzoner/medzoner-go/internal/application/event"
	handler2 "github.com/Medzoner/medzoner-go/internal/ui/http/index"
	mockBase "github.com/Medzoner/medzoner-go/test"

	"github.com/Medzoner/gomedz/pkg/http"
	srv "github.com/Medzoner/gomedz/pkg/http/server"

	"context"
	"github.com/Medzoner/gomedz/pkg/auth"
	"github.com/Medzoner/gomedz/pkg/captcha"
	"github.com/Medzoner/gomedz/pkg/connector"
	"github.com/Medzoner/gomedz/pkg/http/adapter/fiber"
	"github.com/Medzoner/gomedz/pkg/http/probes"
	"github.com/Medzoner/gomedz/pkg/logger"
	"github.com/Medzoner/gomedz/pkg/notifier"
	"github.com/Medzoner/gomedz/pkg/observability"
	"github.com/Medzoner/gomedz/pkg/validation"
	"github.com/Medzoner/medzoner-go/internal/adapters/contact"
	"github.com/Medzoner/medzoner-go/internal/config"
	repository2 "github.com/Medzoner/medzoner-go/internal/ports/contact"
	database2 "github.com/Medzoner/medzoner-go/pkg/database"
	"github.com/Medzoner/medzoner-go/test/mocks"
	"github.com/google/wire"
)

func controllers(p *probes.Handler, a handler2.IndexHandler) []http.Controller {
	return []http.Controller{
		p,
		a,
	}
}

func closers(tl observability.Telemetry) []srv.Closer {
	return []srv.Closer{
		tl,
	}
}

func pingers() probes.Pingers {
	return []probes.Probes{}
}

func middlewaresStructEmpty() []http.Middleware[struct{}] {
	return []http.Middleware[struct{}]{}
}

func middlewaresAny() []http.Middleware[any] {
	return []http.Middleware[any]{}
}

func newEngineWithRenderer(ctx context.Context, cfg srv.Config, l logger.Interface, authCfg auth.Config, mdrs []http.Middleware[struct{}], renderer *http.ReloadingHTMLRenderer) *fiber.Engine[any] {
	engine := fiber.New(ctx, cfg, l, authCfg, mdrs)
	engine.SetRenderer(renderer)
	return engine
}

func newServer(
	ctx context.Context,
	log logger.Interface,
	tel observability.Telemetry,
	cfg srv.Config,
	engine srv.Enginer,
	closers []srv.Closer,
	mdwrs []http.Middleware[any],
	controllers []http.Controller,
) srv.Server {
	s := srv.NewServer(ctx, log, tel, cfg, engine, closers, mdwrs, controllers...)

	engine.SetNotFoundHandler(http.DefaultNotFoundHandler("404"))

	return s
}

func newHTMLRenderer(rootPath config.RootPath) (*http.ReloadingHTMLRenderer, error) {
	base := string(rootPath) + "tmpl"
	renderer := http.NewReloadingHTMLRenderer(base) // récursif, .html + .tmpl
	//if err != nil {
	//	return nil, fmt.Errorf("error creating HTML renderer: %w", err)
	//}
	return renderer, nil
}

var (
	CommonWiring = wire.NewSet(
		config.NewConfig,
		wire.FieldsOf(
			new(config.Config),
			"Obs",
			"Engine",
			"Logger",
			"Auth",
			"Server",
			"Mailer",
			"Database",
			"Recaptcha",
			"RootPath",
		),

		pingers,
		probes.New,
	)
	ServerWiring = wire.NewSet(
		newHTMLRenderer,
		wire.Bind(new(http.Renderer), new(*http.ReloadingHTMLRenderer)),
		newEngineWithRenderer,
		wire.Bind(new(srv.Enginer), new(*fiber.Engine[any])),
		controllers,
		closers,
		middlewaresStructEmpty,
		middlewaresAny,

		newServer,
	)
	ObsWiring = wire.NewSet(
		logger.NewLogger,
		observability.NewTelemetry,
	)
	UsecaseWiring = wire.NewSet(
		event2.NewContactCreatedEventHandler,
		command.NewCreateContactCommandHandler,

		wire.Bind(new(event2.Handler), new(*event2.ContactCreatedEventHandler)),
	)
	HandlerWiring = wire.NewSet(
		handler2.NewIndexHandler,
	)

	InfraWiring = wire.NewSet(
		validation.New,
		captcha.NewRecaptchaAdapter,
		wire.Bind(new(validation.Validater), new(*validation.ValidatorAdapter)),
		wire.Bind(new(captcha.Captcher), new(*captcha.RecaptchaAdapter)),
	)
	DbWiring = wire.NewSet(
		connector.NewDbSQLInstance,

		wire.Bind(new(connector.DbInstantiator), new(*connector.DbSQLInstance)),
	)
	MailerWiring = wire.NewSet(
		notifier.NewMailerSMTP,
		wire.Bind(new(event2.Mailer), new(*notifier.MailerSMTP)),
	)
	MailerMockWiring = wire.NewSet(
		wire.FieldsOf(
			new(*mockBase.Mocks),
			"Mailer",
		),
		wire.Bind(new(event2.Mailer), new(*mocks.MockMailer)),
	)
	RepositoryWiring = wire.NewSet(
		contact.NewRepository,

		wire.Bind(new(repository2.Repository), new(*contact.Repository)),
	)
	RepositoryMockWiring = wire.NewSet(
		wire.FieldsOf(
			new(*mockBase.Mocks),
			"ContactRepository",
		),
		wire.Bind(new(repository2.Repository), new(*mocks.MockRepository)),
	)
	AppWiring = wire.NewSet(
		event2.NewContactCreatedEventHandler,
		command.NewCreateContactCommandHandler,

		wire.Bind(new(event2.Handler), new(*event2.ContactCreatedEventHandler)),
	)
	UiWiring = wire.NewSet(
		handler2.NewIndexHandler,
	)
)

func InitDbMigration() (database2.DbMigration, error) {
	panic(wire.Build(database2.NewDbMigration, CommonWiring, DbWiring))
}

func InitServerTest(ctx context.Context, m *mockBase.Mocks) (srv.Server, error) {
	panic(wire.Build(
		InfraWiring,
		MailerMockWiring,
		UsecaseWiring,
		CommonWiring,
		ObsWiring,
		RepositoryMockWiring,
		HandlerWiring,
		ServerWiring,
	))
}

func InitServer(ctx context.Context) (srv.Server, error) {
	panic(wire.Build(
		DbWiring,
		InfraWiring,
		MailerWiring,
		UsecaseWiring,
		CommonWiring,
		ObsWiring,
		RepositoryWiring,
		HandlerWiring,
		ServerWiring,
	))
}
