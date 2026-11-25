// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

import Agenda from './Agenda.jsx';
import { obtenerAgenda } from '../../../adapters/http/agendaApi.js';

vi.mock('../../hooks/useSesion.js', () => ({
  useSesion: () => ({
    usuario: { id: 'escolar-1', rol: 'escolar', nombre_uso: 'Escolar Uno' },
  }),
}));

vi.mock('../../../adapters/http/agendaApi.js', () => ({
  obtenerAgenda: vi.fn(),
}));

vi.mock('../../../adapters/http/sesionApi.js', () => ({
  obtenerSesionPorReserva: vi.fn(),
}));

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/agenda']}>
      <Routes>
        <Route path="/agenda" element={<Agenda />} />
        <Route path="/sesiones/:id" element={<p>Destino</p>} />
      </Routes>
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
});

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

describe('B-04: estados crudos en la agenda', () => {
  it('traduce un estado desconocido en lugar de mostrarlo tal cual', async () => {
    obtenerAgenda.mockResolvedValue([{
      id: 'reserva-estado-raro',
      curso: 'Matemáticas',
      tema: 'Ecuaciones',
      tutor_nombre: 'Tutor de Matemáticas',
      estado: 'en_revision_desconocida',
      inicio: '2099-11-20T15:00:00Z',
      fin: '2099-11-20T16:00:00Z',
      inicio_lima: '2099-11-20 10:00',
      fin_lima: '2099-11-20 11:00',
    }]);

    renderizar();

    await screen.findByRole('button', { name: 'Próximas (1)' });

    expect(screen.queryByText(/en_revision_desconocida/)).toBeNull();
    expect(screen.getByText(/estado desconocido|desconocido/i)).toBeTruthy();
  });
});

describe('B-21: actualización temporal de la agenda', () => {
  it('mueve una tutoría a pasadas cuando el tiempo avanza', async () => {
    vi.useFakeTimers();
    obtenerAgenda.mockResolvedValue([{
      id: 'reserva-que-vence',
      curso: 'Matemáticas',
      tema: 'Ecuaciones',
      tutor_nombre: 'Tutor de Matemáticas',
      estado: 'confirmada',
      inicio: new Date(Date.now() + 30_000).toISOString(),
      fin: new Date(Date.now() + 60_000).toISOString(),
      inicio_lima: '2099-11-20 10:00',
      fin_lima: '2099-11-20 11:00',
    }]);

    renderizar();
    await act(async () => { await vi.advanceTimersByTimeAsync(0); });

    expect(screen.getByRole('button', { name: 'Próximas (1)' })).toBeTruthy();

    await act(async () => { await vi.advanceTimersByTimeAsync(5 * 60_000); });

    expect(screen.getByRole('button', { name: 'Próximas (0)' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Pasadas (1)' })).toBeTruthy();
  });
});
