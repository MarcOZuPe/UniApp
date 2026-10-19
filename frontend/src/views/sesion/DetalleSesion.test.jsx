// @vitest-environment jsdom

import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  fireEvent,
  render,
  screen,
} from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

import DetalleSesion from './DetalleSesion.jsx';

import {
  obtenerSesion,
  declararAsistencia,
} from '../../controllers/sesion/apiSesion.js';

// Simulamos las respuestas de la API de I05.
vi.mock('../../controllers/sesion/apiSesion.js', () => ({
  obtenerSesion: vi.fn(),
  declararAsistencia: vi.fn(),
}));

// Simulamos un tutor autenticado.
vi.mock('../../controllers/acceso/useSesion.js', () => ({
  useSesion: () => ({
    usuario: {
      id: 'tutor-001',
      rol: 'tutor',
      nombre_uso: 'Tutor Prueba',
    },
  }),
}));

const sesionInicial = {
  id: 'sesion-001',
  reserva_id: 'reserva-001',
  tutor_id: 'tutor-001',
  escolar_id: 'escolar-001',
  inicio: '2026-10-20T15:00:00Z',
  fin: '2026-10-20T16:00:00Z',
  asistencias: [],
};

function renderizarDetalle() {
  return render(
    <MemoryRouter initialEntries={['/sesiones/sesion-001']}>
      <Routes>
        <Route
          path="/sesiones/:id"
          element={<DetalleSesion />}
        />
      </Routes>
    </MemoryRouter>
  );
}

describe('I05: Detalle de sesion', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('muestra los datos de una sesion y el boton de asistencia', async () => {
    obtenerSesion.mockResolvedValue(sesionInicial);

    renderizarDetalle();

    expect(
      await screen.findByText('reserva-001')
    ).toBeTruthy();

    expect(
      screen.getByRole('button', {
        name: 'Declarar mi asistencia',
      })
    ).toBeTruthy();

    expect(obtenerSesion).toHaveBeenCalledWith('sesion-001');
  });

  it('registra la asistencia y deshabilita el boton', async () => {
    const sesionActualizada = {
      ...sesionInicial,
      asistencias: [
        {
          usuario_id: 'tutor-001',
          rol: 'tutor',
          declarada_en: '2026-10-20T15:00:00Z',
        },
      ],
    };

    obtenerSesion
      .mockResolvedValueOnce(sesionInicial)
      .mockResolvedValueOnce(sesionActualizada);

    declararAsistencia.mockResolvedValue(
      sesionActualizada.asistencias[0]
    );

    renderizarDetalle();

    const boton = await screen.findByRole('button', {
      name: 'Declarar mi asistencia',
    });

    fireEvent.click(boton);

    const botonActualizado = await screen.findByRole('button', {
      name: 'Asistencia ya declarada',
    });

    expect(botonActualizado.disabled).toBe(true);
    expect(declararAsistencia).toHaveBeenCalledWith('sesion-001');
  });
});
