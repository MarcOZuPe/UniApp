// @vitest-environment jsdom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react';

import App from '../../../app/App.jsx';
import { obtenerAgenda } from '../../../adapters/http/agendaApi.js';

// Simulamos únicamente la sesión y la respuesta de la API.
const sesion = vi.hoisted(() => ({
  usuario: null,
  cargando: false,
}));

vi.mock('../../hooks/useSesion.js', () => ({
  useSesion: () => sesion,
}));

vi.mock('../../../application/acceso/sesion.js', () => ({
  cargarSesion: vi.fn(),
  salir: vi.fn(),
}));

vi.mock('../../../adapters/http/agendaApi.js', () => ({
  obtenerAgenda: vi.fn(),
}));

const proxima = {
  id: 'reserva-futura',
  curso: 'Matemáticas',
  tema: 'Ecuaciones de segundo grado',
  tutor_nombre: 'Tutor de Matemáticas',
  estado: 'confirmada',
  inicio: '2099-11-20T15:00:00Z',
  fin: '2099-11-20T16:00:00Z',
  inicio_lima: '2099-11-20 10:00',
  fin_lima: '2099-11-20 11:00',
};

const pasada = {
  id: 'reserva-pasada',
  curso: 'Física',
  tema: 'Leyes de Newton',
  tutor_nombre: 'Tutor de Física',
  estado: 'confirmada',
  inicio: '2020-10-15T15:00:00Z',
  fin: '2020-10-15T16:00:00Z',
  inicio_lima: '2020-10-15 10:00',
  fin_lima: '2020-10-15 11:00',
};

function abrirAgenda() {
  window.history.replaceState({}, '', '/agenda');
  return render(<App />);
}

beforeEach(() => {
  vi.resetAllMocks();

  sesion.usuario = {
    id: 'escolar-prueba',
    rol: 'escolar',
    nombre_uso: 'Escolar Prueba',
  };

  sesion.cargando = false;

  obtenerAgenda.mockResolvedValue([proxima, pasada]);
});

afterEach(() => {
  cleanup();
});

describe('I10 - Pantalla de agenda', () => {
  it('consulta la API cuando el escolar abre la agenda', async () => {
    abrirAgenda();

    expect(
      screen.getByText('Cargando agenda…')
    ).toBeTruthy();

    expect(
      await screen.findByRole('heading', { name: 'Mi agenda' })
    ).toBeTruthy();

    await waitFor(() => {
      expect(obtenerAgenda).toHaveBeenCalledTimes(1);
    });

    expect(
      await screen.findByRole('heading', { name: 'Matemáticas' })
    ).toBeTruthy();
  });

  it('muestra inicialmente las tutorías próximas', async () => {
    abrirAgenda();

    expect(
      await screen.findByRole('heading', { name: 'Matemáticas' })
    ).toBeTruthy();

    expect(
      screen.getByText('Ecuaciones de segundo grado')
    ).toBeTruthy();

    expect(
      screen.getByText('2099-11-20 10:00 (hora de Lima)')
    ).toBeTruthy();

    expect(
      screen.queryByRole('heading', { name: 'Física' })
    ).toBeNull();
  });

  it('permite consultar las tutorías pasadas', async () => {
    abrirAgenda();

    await screen.findByRole('heading', {
      name: 'Matemáticas',
    });

    fireEvent.click(
      screen.getByRole('button', { name: 'Pasadas (1)' })
    );

    expect(
      screen.getByRole('heading', { name: 'Física' })
    ).toBeTruthy();

    expect(
      screen.getByText('Leyes de Newton')
    ).toBeTruthy();

    expect(
      screen.queryByRole('heading', { name: 'Matemáticas' })
    ).toBeNull();
  });

  it('muestra un mensaje cuando no existen reservas', async () => {
    obtenerAgenda.mockResolvedValue([]);

    abrirAgenda();

    expect(
      await screen.findByText(
        'No tienes tutorías en esta sección.'
      )
    ).toBeTruthy();

    expect(
      screen.getByRole('button', { name: 'Próximas (0)' })
    ).toBeTruthy();
  });

  it('muestra los errores de la API', async () => {
    obtenerAgenda.mockRejectedValue(
      new Error('No hay conexión con el servidor.')
    );

    abrirAgenda();

    const alerta = await screen.findByRole('alert');

    expect(alerta.textContent).toContain(
      'No hay conexión con el servidor.'
    );

    expect(
      screen.queryByRole('heading', { name: 'Matemáticas' })
    ).toBeNull();
  });

  it('permite al escolar reservar otra tutoría', async () => {
    abrirAgenda();

    await screen.findByRole('heading', {
      name: 'Matemáticas',
    });

    const enlace = screen.getByRole('link', {
      name: 'Reservar otra tutoría',
    });

    expect(enlace.getAttribute('href')).toBe('/reservar');
  });

  it('permite al tutor consultar su agenda', async () => {
    sesion.usuario = {
      id: 'tutor-prueba',
      rol: 'tutor',
      nombre_uso: 'Tutor Prueba',
    };

    abrirAgenda();

    expect(
      await screen.findByText(
        'Consulta las tutorías que tienes asignadas.'
      )
    ).toBeTruthy();

    await screen.findByRole('heading', {
      name: 'Matemáticas',
    });

    expect(
      screen.queryByRole('link', {
        name: 'Reservar otra tutoría',
      })
    ).toBeNull();
  });

  it.each(['coordinador', 'apoderado'])(
    'impide el acceso al rol %s',
    async (rol) => {
      sesion.usuario = {
        id: 'usuario-prueba',
        rol,
        nombre_uso: 'Usuario Prueba',
      };

      abrirAgenda();

      await waitFor(() => {
        expect(window.location.pathname).toBe(
          '/no-autorizado'
        );
      });

      expect(obtenerAgenda).not.toHaveBeenCalled();
    }
  );

  it('redirige al ingreso si no existe sesión', async () => {
    sesion.usuario = null;

    abrirAgenda();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/ingreso');
    });

    expect(obtenerAgenda).not.toHaveBeenCalled();
  });
});
