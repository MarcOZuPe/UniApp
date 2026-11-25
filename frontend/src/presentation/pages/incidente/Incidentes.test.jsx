// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import Incidentes from './Incidentes.jsx';
import { listarIncidentes } from '../../../adapters/http/incidenteApi.js';

const estadoSesion = vi.hoisted(() => ({
  usuario: { id: 'escolar-1', rol: 'escolar', nombre_uso: 'Escolar Uno' },
}));

vi.mock('../../hooks/useSesion.js', () => ({
  useSesion: () => ({ usuario: estadoSesion.usuario }),
}));

vi.mock('../../../adapters/http/incidenteApi.js', () => ({
  listarIncidentes: vi.fn(),
  revisarIncidente: vi.fn(),
  urlEvidencia: vi.fn(),
}));

function crearIncidente(indice, extra = {}) {
  return {
    id: `incidente-${indice}`,
    tipo: 'conducta',
    estado: 'pendiente',
    escolar_id: 'escolar-1',
    escolar_nombre: 'Escolar Uno',
    reportado_por: 'escolar-1',
    reportado_por_rol: 'escolar',
    tutor_id: 'tutor-1',
    tutor_nombre: 'Tutor Uno',
    curso: 'Álgebra',
    inicio_lima: '2026-11-24 10:00',
    descripcion: `Descripción del incidente ${indice}`,
    resultado: undefined,
    evidencias: [],
    ...extra,
  };
}

function renderizar(entradas = ['/incidentes']) {
  return render(
    <MemoryRouter initialEntries={entradas}>
      <Incidentes />
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
  estadoSesion.usuario = { id: 'escolar-1', rol: 'escolar', nombre_uso: 'Escolar Uno' };
});

afterEach(cleanup);

describe('F-16: listados sin paginación', () => {
  it('pagina el histórico de incidentes en lugar de renderizarlo completo', async () => {
    const incidentes = Array.from({ length: 30 }, (_, indice) => crearIncidente(indice));
    listarIncidentes.mockResolvedValue({ incidentes });

    renderizar();

    await screen.findAllByRole('button', { name: /Reportado por/ });

    expect(screen.getByRole('navigation', { name: /paginación/i })).toBeTruthy();
  });
});

describe('B-01: mensaje de éxito persistente', () => {
  it('no muestra un mensaje de éxito viejo al volver a abrir la pantalla', async () => {
    listarIncidentes.mockResolvedValue({ incidentes: [] });

    renderizar([{ pathname: '/incidentes', state: { mensaje: 'Revisión registrada.' } }]);

    await screen.findByText('No hay incidentes registrados.');

    expect(screen.queryByText('Revisión registrada.')).toBeNull();
  });
});

describe('B-04: estados crudos o desconocidos', () => {
  it('no presenta un estado desconocido como "Pendiente de revisión"', async () => {
    listarIncidentes.mockResolvedValue({
      incidentes: [crearIncidente(1, { estado: 'archivado' })],
    });

    renderizar();

    await screen.findAllByRole('button', { name: /Reportado por/ });

    expect(screen.queryByText('Pendiente de revisión')).toBeNull();
    expect(screen.getByText(/archivado|desconocido/i)).toBeTruthy();
  });
});

describe('errores no recuperables en incidentes', () => {
  it('ofrece reintentar cuando falla la carga de incidentes', async () => {
    listarIncidentes
      .mockRejectedValueOnce(new Error('Sin conexión.'))
      .mockResolvedValueOnce({ incidentes: [] });

    renderizar();

    expect((await screen.findByRole('alert')).textContent).toContain('Sin conexión.');

    expect(screen.getByRole('button', { name: /reintentar/i })).toBeTruthy();
  });
});
