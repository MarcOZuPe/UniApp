package acceso

import (
	"context"

	"uniapp/internal/events"
)

func (s *Servicio) Registrar(ctx context.Context, datos DatosRegistro) (*Usuario, error) {
	usuario, err := s.fabrica.CrearPublico(datos)
	if err != nil {
		return nil, err
	}
	registrado, err := s.repositorio.CrearUsuario(ctx, usuario)
	if err != nil {
		return nil, err
	}
	s.bus.Notificar(eventoDeUsuario(events.UsuarioRegistrado, registrado))
	return registrado, nil
}

func (s *Servicio) CrearAdministrado(ctx context.Context, datos DatosRegistro) (*Usuario, error) {
	usuario, err := s.fabrica.CrearInterno(datos)
	if err != nil {
		return nil, err
	}
	return s.repositorio.CrearUsuario(ctx, usuario)
}
