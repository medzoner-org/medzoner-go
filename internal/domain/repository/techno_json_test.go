package repository_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Medzoner/medzoner-go/internal/config"
	"github.com/Medzoner/medzoner-go/internal/domain/repository"
	"gotest.tools/assert"
)

func TestTechnoJSONRepository_FetchStack(t *testing.T) {
	currentPath, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	rootPath := filepath.Dir(filepath.Dir(filepath.Dir(currentPath))) + "/"

	t.Run("Unit: test FetchStack success", func(t *testing.T) {
		repo := repository.NewTechnoJSONRepository(config.Config{
			RootPath: config.RootPath(rootPath),
		})

		result, err := repo.FetchStack(context.Background())

		assert.NilError(t, err)
		assert.Assert(t, result != nil)
		assert.Assert(t, len(result) > 0)
	})

	t.Run("Unit: test FetchStack file not found", func(t *testing.T) {
		repo := repository.NewTechnoJSONRepository(config.Config{
			RootPath: config.RootPath("/nonexistent/path/"),
		})

		_, err := repo.FetchStack(context.Background())

		assert.ErrorContains(t, err, "error during open json file")
	})

	t.Run("Unit: test FetchStack invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		jsonDir := filepath.Join(tmpDir, "internal", "resources", "data", "jobs")
		err := os.MkdirAll(jsonDir, 0o755)
		assert.NilError(t, err)

		err = os.WriteFile(filepath.Join(jsonDir, "stacks.json"), []byte("{invalid json}"), 0o644)
		assert.NilError(t, err)

		repo := repository.NewTechnoJSONRepository(config.Config{
			RootPath: config.RootPath(tmpDir + "/"),
		})

		_, err = repo.FetchStack(context.Background())

		assert.ErrorContains(t, err, "error during unmarshal json")
	})

	t.Run("Unit: test FetchStack with empty JSON file", func(t *testing.T) {
		tmpDir := t.TempDir()
		jsonDir := filepath.Join(tmpDir, "internal", "resources", "data", "jobs")
		err := os.MkdirAll(jsonDir, 0o755)
		assert.NilError(t, err)

		err = os.WriteFile(filepath.Join(jsonDir, "stacks.json"), []byte(""), 0o644)
		assert.NilError(t, err)

		repo := repository.NewTechnoJSONRepository(config.Config{
			RootPath: config.RootPath(tmpDir + "/"),
		})

		_, err = repo.FetchStack(context.Background())

		assert.ErrorContains(t, err, "error during unmarshal json")
	})
}

