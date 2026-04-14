package config_test

import (
	"testing"

	"github.com/Medzoner/medzoner-go/internal/config"
	"gotest.tools/assert"
)

func TestNewConfig(t *testing.T) {
	t.Run("Unit: test NewConfig with default env", func(t *testing.T) {
		t.Setenv("ROOT_PATH", "/tmp/test/")

		cfg, err := config.NewConfig()

		assert.NilError(t, err)
		assert.Equal(t, string(cfg.RootPath), "/tmp/test/")
	})

	t.Run("Unit: test NewConfig with multiple env vars", func(t *testing.T) {
		t.Setenv("ROOT_PATH", "/app/")
		t.Setenv("DATABASE_DSN", "user:pass@tcp(localhost:3306)")
		t.Setenv("DATABASE_NAME", "testdb")

		cfg, err := config.NewConfig()

		assert.NilError(t, err)
		assert.Equal(t, string(cfg.RootPath), "/app/")
		assert.Equal(t, cfg.Database.Dsn, "user:pass@tcp(localhost:3306)")
		assert.Equal(t, cfg.Database.Name, "testdb")
	})
}
