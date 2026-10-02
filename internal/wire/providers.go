package wire

import (
	"context"

	"github.com/Medzoner/gomedz/pkg/auth"
	"github.com/Medzoner/gomedz/pkg/http"
	"github.com/Medzoner/gomedz/pkg/http/adapter/fiber"
	"github.com/Medzoner/gomedz/pkg/http/probes"
	srv "github.com/Medzoner/gomedz/pkg/http/server"
	"github.com/Medzoner/gomedz/pkg/logger"
	"github.com/Medzoner/gomedz/pkg/observability"
	"github.com/Medzoner/medzoner-go/internal/config"
	handler2 "github.com/Medzoner/medzoner-go/internal/ui/http/index"
)

func controllers(p *probes.Handler, a handler2.Handler) []http.Controller {
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
