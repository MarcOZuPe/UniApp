// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

import DetalleSesion from './DetalleSesion.jsx';
import { obtenerSesion } from '../../../adapters/http/sesionApi.js';

vi.mock('../../hooks/useSesion.js', () => ({
  useSesion: () => ({
    usuario: { id: 'escolar-1', rol: 'escolar', nombre_uso: 'Escolar Uno' },
  }),
}));

vi.mock('../../../adapters/http/sesionApi.js', () => ({
  obtenerSesion: vi.fn(),
  declararAsistencia: vi.fn(),
  reportarAusencia: vi.fn(),
  registrarResultado: vi.fn(),
}));

vi.mock('./SeccionMinutos.jsx', () => ({
  default: () => null,
}));

const sesionBase = {
  id: 'sesion-1',
  reserva_id: 'reserva-1',
  tutor_id: 'tutor-1',
  escolar_id: 'escolar-1',
  inicio: '2026-11-24T15:00:00Z',
  fin: '2026-11-24T16:00:00Z',
  asistencias: [],
  ausencias: [],
  plazos: {},
  resultado: null,
};

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/sesiones/sesion-1']}>
      <Routes>
        <Route path="/sesiones/:id" element={<DetalleSesion />} />
      </Routes>
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
});

afterEach(cleanup);

describe('B-06: duración fija de 60 minutos', () => {
  it('deriva la duración de inicio y fin en lugar de asumir 60 minutos', async () => {
    obtenerSesion.mockResolvedValue({
      ...sesionBase,
      inicio: '2026-11-24T15:00:00Z',
      fin: '2026-11-24T16:30:00Z',
    });

    renderizar();

    expect(await screen.findByRole('heading', { name: 'Detalle de tutoría' })).toBeTruthy();
    expect(screen.getByText(/90 minutos/)).toBeTruthy();
  });
});

describe('F-21: fechas inválidas sin error boundary', () => {
  it('muestra un aviso en lugar de lanzar RangeError con una fecha inválida', async () => {
    obtenerSesion.mockResolvedValue({ ...sesionBase, inicio: 'fecha-invalida' });

    let errorRender;
    try {
      renderizar();
      await screen.findByRole('heading', { name: 'Detalle de tutoría' });
      expect(screen.getByText(/fecha inválida|no se pudo mostrar/i)).toBeTruthy();
    } catch (error) {
      errorRender = error;
    }

    expect(errorRender).toBeUndefined();
  });
});
