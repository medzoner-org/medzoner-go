package templater_test

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Medzoner/medzoner-go/internal/ui/http/templater"
	"gotest.tools/assert"
	"github.com/Medzoner/medzoner-go/internal/config"
)

func TestRender(t *testing.T) {
	currentPath, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	rootPath := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(currentPath)))) + "/"
	t.Run("Unit: test Render success", func(t *testing.T) {
		cfg := config.Config{
			RootPath: config.RootPath(rootPath),
		}
		tpl := templater.NewTemplateHTML(cfg)
		err := tpl.Render(
			"index",
			nil,
			httptest.NewRecorder(),
		)
		if err != nil {
			t.Errorf("Error: %v", err)
			return
		}
		assert.Equal(t, true, true)
	})
	t.Run("Unit: test Render failed when parse tpl", func(t *testing.T) {
		tpl := templater.TemplateHTML{
			RootPath: rootPath + ".var/test/",
		}
		err := os.Chmod(tpl.RootPath+"tmpl/failed.html", 0o000)
		if err != nil {
			t.Error(err)
		}

		err = tpl.Render(
			"failed.html",
			nil,
			httptest.NewRecorder(),
		)

		assert.ErrorContains(t, err, "error parsing templates: error getting template "+rootPath+".var/test/tmpl/: error parsing files tpl: open "+rootPath+".var/test/tmpl/failed.html: permission denied - info: ")

		err = os.Chmod(tpl.RootPath+"/tmpl/failed.html", 0o600)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Unit: test Render failed with bad tpl name", func(t *testing.T) {
		tpl := templater.TemplateHTML{
			RootPath: "../../../..",
		}
		err := tpl.Render(
			"fail",
			nil,
			httptest.NewRecorder(),
		)

		assert.Error(t, err, "error parsing templates: error getting template ../../../..tmpl/: error walking the path ../../../..tmpl/: lstat ../../../..tmpl/: no such file or directory")
	})
	t.Run("Unit: test Render failed with bad rootPath", func(t *testing.T) {
		tpl := templater.TemplateHTML{
			RootPath: "../../..",
		}
		err := tpl.Render(
			"index",
			nil,
			httptest.NewRecorder(),
		)

		assert.Error(t, err, "error parsing templates: error getting template ../../..tmpl/: error walking the path ../../..tmpl/: lstat ../../..tmpl/: no such file or directory")
	})
	t.Run("Unit: test Render failed execute template with unknown name", func(t *testing.T) {
		cfg := config.Config{
			RootPath: config.RootPath(rootPath),
		}
		tpl := templater.NewTemplateHTML(cfg)
		err := tpl.Render(
			"nonexistent_template",
			nil,
			httptest.NewRecorder(),
		)

		assert.ErrorContains(t, err, "error executing template")
	})
}
