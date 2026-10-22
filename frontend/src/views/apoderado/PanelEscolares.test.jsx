// @vitest-environment jsdom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import PanelEscolares from './PanelEscolares.jsx';
import {
  cambiarAutorizacion,
  historialAutorizaciones,
  listarVinculaciones,
} from '../../controllers/apoderado/apiApoderado.js';

vi.mock('../../controllers/apoderado/apiApoderado.js', () => ({
  cambiarAutorizacion: vi.fn(),
  historialAutorizaciones: vi.fn(),
  listarVinculaciones: vi.fn(),
}));

function renderizar() {
  return render(
    <MemoryRouter>
      <PanelEscolares />
    </MemoryRouter>
  );
}

function vinculacionAceptada(autorizado = true) {
  return {
    id: 'v-1',
    estado: 'aceptada',
    escolar: { id: 'e-1', nombre_uso: 'Escolar Prueba', email: 'escolar@example.com' },
    autorizacion: {
      id: autorizado ? 'a-1' : 'a-2',
      estado: autorizado ? 'autorizado' : 'no_autorizado',
      autorizado,
    },
  };
}

describe('I06: panel de escolares', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    historialAutorizaciones.mockResolvedValue({ autorizaciones: [] });
  });

  afterEach(() => {
    cleanup();
  });

  it('muestra la autorización y el historial de cada escolar', async () => {
    listarVinculaciones.mockResolvedValue({
      vinculaciones: [vinculacionAceptada(false)],
    });
    historialAutorizaciones.mockResolvedValue({
      autorizaciones: [
        {
          id: 'a-2',
          estado: 'no_autorizado',
          autorizado: false,
          creada_en: '2026-10-20T15:30:00Z',
        },
        {
          id: 'a-1',
          estado: 'autorizado',
          autorizado: true,
          creada_en: '2026-10-18T12:00:00Z',
        },
      ],
    });

    renderizar();

    expect(
      await screen.findByRole('heading', { name: 'Panel de escolares' })
    ).toBeTruthy();
    expect(await screen.findByRole('table')).toBeTruthy();
    expect(screen.getByText('No puede solicitar nuevas tutorías.')).toBeTruthy();
    expect(historialAutorizaciones).toHaveBeenCalledWith('e-1');
    expect(screen.getByRole('columnheader', { name: 'Fecha y hora' })).toBeTruthy();
    expect(screen.getByRole('columnheader', { name: 'Cambio' })).toBeTruthy();
    expect(screen.getAllByRole('row')).toHaveLength(3);
    expect(
      screen.getAllByText('Se retiró el permiso para pedir tutorías').length
    ).toBeGreaterThan(0);
    expect(
      screen.getAllByText('Se otorgó el permiso para pedir tutorías').length
    ).toBeGreaterThan(0);
  });

  it('permite revocar y volver a autorizar desde el panel', async () => {
    listarVinculaciones
      .mockResolvedValueOnce({ vinculaciones: [vinculacionAceptada(true)] })
      .mockResolvedValueOnce({ vinculaciones: [vinculacionAceptada(false)] });
    cambiarAutorizacion.mockResolvedValue({
      mensaje: 'Autorización revocada. El escolar no podrá crear nuevas reservas.',
    });

    renderizar();

    expect(await screen.findByText('Puede solicitar nuevas tutorías.')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Revocar autorización' }));

    expect(await screen.findByText(/Autorización revocada/)).toBeTruthy();
    expect(cambiarAutorizacion).toHaveBeenCalledWith('e-1', false);
    expect(await screen.findByText('No puede solicitar nuevas tutorías.')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Autorizar nuevas tutorías' })).toBeTruthy();
  });

  it('muestra las vinculaciones pendientes sin acciones', async () => {
    listarVinculaciones.mockResolvedValue({
      vinculaciones: [
        {
          id: 'v-2',
          estado: 'pendiente',
          escolar: {
            id: 'e-2',
            nombre_uso: 'Escolar Pendiente',
            email: 'pendiente@example.com',
          },
        },
      ],
    });

    renderizar();

    expect(await screen.findByText('Pendiente de aceptación')).toBeTruthy();
    expect(screen.queryByRole('button', { name: /autorización/i })).toBeNull();
    expect(screen.queryByRole('table')).toBeNull();
  });
});
