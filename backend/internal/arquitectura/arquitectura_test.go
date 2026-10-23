package arquitectura

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

const prefijo = "uniapp/internal/"

func capaYFlujo(rel string) (string, string, bool) {
	p := strings.Split(filepath.ToSlash(rel), "/")
	if len(p) < 3 {
		return "", "", false
	}
	switch p[0] {
	case "models", "views", "controllers":
		return p[0], p[1], true
	}
	return "", "", false
}

func permitido(flujo, otro string) bool { return otro == flujo || otro == "acceso" }

func TestReglasMVCPorFlujo(t *testing.T) {
	raiz := ".."
	_ = filepath.WalkDir(raiz, func(ruta string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(ruta, ".go") || strings.HasSuffix(ruta, "_test.go") {
			return nil
		}
		rel, _ := filepath.Rel(raiz, ruta)
		capa, flujo, ok := capaYFlujo(rel)
		if !ok {
			return nil
		}
		f, perr := parser.ParseFile(token.NewFileSet(), ruta, nil, parser.ImportsOnly)
		if perr != nil {
			t.Errorf("%s: %v", rel, perr)
			return nil
		}
		for _, im := range f.Imports {
			imp := strings.Trim(im.Path.Value, `"`)
			if !strings.HasPrefix(imp, prefijo) {
				continue
			}
			c2, f2, ok2 := capaYFlujo(strings.TrimPrefix(imp, prefijo) + "/x.go")
			if !ok2 {
				continue
			}
			switch {
			case capa == "models" && c2 != "models":
				t.Errorf("%s: el modelo no puede importar %s", rel, imp)
			case capa == "views" && c2 == "controllers":
				t.Errorf("%s: la vista no puede importar controladores (%s)", rel, imp)
			case c2 == "models" && !permitido(flujo, f2):
				t.Errorf("%s: %s no puede usar el modelo del flujo %q (usar su API o fixture)", rel, flujo, f2)
			case c2 == "controllers" && capa == "controllers" && !permitido(flujo, f2):
				t.Errorf("%s: no importar el controlador del flujo %q; consumirlo por su API", rel, f2)
			}
		}
		return nil
	})
}

func TestModeloNoImportaInfraestructura(t *testing.T) {
	raiz := filepath.Join("..", "models")
	_ = filepath.WalkDir(raiz, func(ruta string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(ruta, ".go") || strings.HasSuffix(ruta, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(token.NewFileSet(), ruta, nil, parser.ImportsOnly)
		if perr != nil {
			t.Errorf("%s: %v", ruta, perr)
			return nil
		}
		for _, im := range f.Imports {
			imp := strings.Trim(im.Path.Value, `"`)
			if strings.HasPrefix(imp, prefijo+"infra") {
				t.Errorf("%s: el modelo no puede importar infraestructura (%s)", ruta, imp)
			}
		}
		return nil
	})
}

func TestCadaFlujoTieneLasTresCapas(t *testing.T) {
	for _, flujo := range []string{"acceso", "tutor", "oferta", "reserva", "sesion", "catalogo"} {
		for _, capa := range []string{"models", "views", "controllers"} {
			m, _ := filepath.Glob(filepath.Join("..", capa, flujo, "*.go"))
			if len(m) == 0 {
				t.Errorf("falta %s/%s", capa, flujo)
			}
		}
	}
}
