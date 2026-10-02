//go:build whyor

package wire

import (
	"github.com/Medzoner/medzoner-go/internal/application/command"
	event2 "github.com/Medzoner/medzoner-go/internal/application/event"
	handler2 "github.com/Medzoner/medzoner-go/internal/ui/http/index"
	mockBase "github.com/Medzoner/medzoner-go/test"

	"github.com/Medzoner/gomedz/pkg/http"
	srv "github.com/Medzoner/gomedz/pkg/http/server"

	"context"
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
	"github.com/Medzoner/whyor"
)

var (
	CommonWiring = whyor.Set(
		config.NewConfig,
		whyor.FieldsOf[config.Config](
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
	ServerWiring = whyor.Set(
		newHTMLRenderer,
		whyor.Bind[http.Renderer, *http.ReloadingHTMLRenderer](),
		newEngineWithRenderer,
		whyor.Bind[srv.Enginer, *fiber.Engine[any]](),
		controllers,
		closers,
		middlewaresStructEmpty,
		middlewaresAny,

		newServer,
	)
	ObsWiring = whyor.Set(
		logger.NewLogger,
		observability.NewTelemetry,
	)
	UsecaseWiring = whyor.Set(
		event2.NewContactCreatedEventHandler,
		command.NewCreateContactHandler,

		whyor.Bind[event2.Handler, *event2.ContactCreatedEventHandler](),
	)
	HandlerWiring = whyor.Set(
		handler2.NewIndexHandler,
	)

	InfraWiring = whyor.Set(
		validation.New,
		captcha.NewRecaptchaAdapter,
		whyor.Bind[validation.Validater, *validation.ValidatorAdapter](),
		whyor.Bind[captcha.Captcher, *captcha.RecaptchaAdapter](),
	)
	DbWiring = whyor.Set(
		connector.NewDbSQLInstance,

		whyor.Bind[connector.DbInstantiator, *connector.DbSQLInstance](),
	)
	MailerWiring = whyor.Set(
		notifier.NewMailerSMTP,
		whyor.Bind[event2.Mailer, *notifier.MailerSMTP](),
	)
	MailerMockWiring = whyor.Set(
		whyor.FieldsOf[*mockBase.Mocks](
			"Mailer",
		),
		whyor.Bind[event2.Mailer, *mocks.MockMailer](),
	)
	RepositoryWiring = whyor.Set(
		contact.NewRepository,

		whyor.Bind[repository2.Repository, *contact.Repository](),
	)
	RepositoryMockWiring = whyor.Set(
		whyor.FieldsOf[*mockBase.Mocks](
			"ContactRepository",
		),
		whyor.Bind[repository2.Repository, *mocks.MockRepository](),
	)
	AppWiring = whyor.Set(
		event2.NewContactCreatedEventHandler,
		command.NewCreateContactHandler,

		whyor.Bind[event2.Handler, *event2.ContactCreatedEventHandler](),
	)
	UiWiring = whyor.Set(
		handler2.NewIndexHandler,
	)
)

func InitDbMigration() (database2.DbMigration, error) {
	panic(whyor.Build(database2.NewDbMigration, CommonWiring, DbWiring))
}

func InitServerTest(ctx context.Context, m *mockBase.Mocks) (srv.Server, error) {
	panic(whyor.Build(
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
	panic(whyor.Build(
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
