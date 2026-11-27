package acceso

import "context"

var facultadesMemoria = []Facultad{
	{ID: "1", Nombre: "Facultad de Ciencias"},
	{ID: "2", Nombre: "Facultad de Ingeniería Industrial y de Sistemas"},
	{ID: "3", Nombre: "Facultad de Ingeniería Mecánica"},
	{ID: "4", Nombre: "Facultad de Arquitectura, Urbanismo y Artes"},
	{ID: "5", Nombre: "Facultad de Ingeniería Ambiental"},
	{ID: "6", Nombre: "Facultad de Ingeniería Civil"},
	{ID: "7", Nombre: "Facultad de Ingeniería Económica, Estadística y Ciencias Sociales"},
	{ID: "8", Nombre: "Facultad de Ingeniería Eléctrica y Electrónica"},
	{ID: "9", Nombre: "Facultad de Ingeniería Geológica, Minera y Metalúrgica"},
	{ID: "10", Nombre: "Facultad de Ingeniería de Petróleo, Gas Natural y Petroquímica"},
	{ID: "11", Nombre: "Facultad de Ingeniería Química y Textil"},
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
