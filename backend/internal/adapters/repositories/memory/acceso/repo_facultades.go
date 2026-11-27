package acceso

import "context"

var facultadesMemoria = []Facultad{
	{ID: "1", Nombre: "Facultad de Ciencias"},
	{ID: "2", Nombre: "Facultad de Ingeniería Industrial y Sistemas"},
	{ID: "3", Nombre: "Facultad de Ingeniería Mecánica"},
	{ID: "4", Nombre: "Facultad de Arquitectura, Urbanismo y Artes"},
}

func (repositorio *RepositorioMemoria) Facultades(context.Context) ([]Facultad, error) {
	copia := append([]Facultad(nil), facultadesMemoria...)
	return copia, nil
}

func (repositorio *RepositorioMemoria) FacultadPorID(_ context.Context, facultadID string) (*Facultad, error) {
	for _, facultad := range facultadesMemoria {
		if facultad.ID == facultadID {
			copia := facultad
			return &copia, nil
		}
	}
	return nil, ErrNoEncontrado
}
