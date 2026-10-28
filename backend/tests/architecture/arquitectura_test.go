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

var flujos = []string{"acceso", "catalogo", "oferta", "reserva", "sesion", "tutor"}

// capaYFlujo identifica la capa Clean y la feature de una ruta relativa a internal/.
func capaYFlujo(rel string) (capa, flujo string, ok bool) {
	p := strings.Split(filepath.ToSlash(rel), "/")
	if len(p) < 2 {
		return "", "", false
	}
	switch p[0] {
	case "domain", "application":
		return p[0], p[1], true
	case "adapters":
		if len(p) < 3 {
			return "", "", false
		}
		switch p[1] {
		case "http", "presentation":
			return "adapters", p[2], true
		case "repositories":
			if len(p) < 4 {
				return "", "", false
			}
			return "adapters", p[3], true
		}
	case "infrastructure":
		return "infrastructure", "", true
	}
	return "", "", false
}

func permitido(flujo, otro string) bool { return otro == flujo || otro == "acceso" }

// TestReglasDeDependenciasClean verifica la regla de dependencias:
//
//	domain <- application <- adapters <- infrastructure
//
// Las capas internas no pueden importar capas externas y una feature no puede
// importar el dominio/aplicación de otra, salvo acceso (identidad compartida).
func TestReglasDeDependenciasClean(t *testing.T) {
	raiz := filepath.Join("..", "..", "internal")
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
			c2, f2, ok2 := capaYFlujo(strings.TrimPrefix(imp, prefijo))
			if !ok2 {
				continue
			}
			switch capa {
			case "domain":
				if c2 != "domain" {
					t.Errorf("%s: domain no puede importar la capa %s (%s)", rel, c2, imp)
				}
			case "application":
				if c2 == "adapters" || c2 == "infrastructure" {
					t.Errorf("%s: application no puede importar la capa %s (%s)", rel, c2, imp)
				}
			case "adapters":
				if c2 == "infrastructure" {
					t.Errorf("%s: adapters no puede importar infrastructure (%s)", rel, imp)
				}
			}
			if flujo != "" && f2 != "" && !permitido(flujo, f2) {
				t.Errorf("%s: la feature %q no puede importar %q de la feature %q", rel, flujo, c2, f2)
			}
		}
		return nil
	})
}

// TestCadaFlujoTieneCapas verifica la estructura mínima por feature.
func TestCadaFlujoTieneCapas(t *testing.T) {
	for _, flujo := range flujos {
		for _, capa := range []string{
			filepath.Join("..", "..", "internal", "domain", flujo),
			filepath.Join("..", "..", "internal", "application", flujo),
			filepath.Join("..", "..", "internal", "adapters", "http", flujo),
		} {
			m, _ := filepath.Glob(filepath.Join(capa, "*.go"))
			if len(m) == 0 {
				t.Errorf("falta %s", capa)
			}
		}
	}
}

// TestCompositionRootUnico verifica que las implementaciones concretas solo se
// importen desde cmd/api. Así el resto del código depende de interfaces.
func TestCompositionRootUnico(t *testing.T) {
	raiz := filepath.Join("..", "..", "internal")
	_ = filepath.WalkDir(raiz, func(ruta string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(ruta, ".go") || strings.HasSuffix(ruta, "_test.go") {
			return nil
		}
		rel, _ := filepath.Rel(raiz, ruta)
		if strings.HasPrefix(filepath.ToSlash(rel), "adapters/repositories/") {
			return nil
		}
		f, perr := parser.ParseFile(token.NewFileSet(), ruta, nil, parser.ImportsOnly)
		if perr != nil {
			return nil
		}
		for _, im := range f.Imports {
			imp := strings.Trim(im.Path.Value, `"`)
			if strings.HasPrefix(imp, prefijo+"adapters/repositories/") {
				t.Errorf("%s: importa un repositorio concreto (%s); debe inyectarse desde cmd/api", rel, imp)
			}
		}
		return nil
	})
}
