// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import PanelEscolares from './PanelEscolares.jsx';
import {
  cambiarAutorizacion,
  historialAutorizaciones,
  listarVinculaciones,
} from '../../../adapters/http/apoderadoApi.js';

vi.mock('../../../adapters/http/apoderadoApi.js', () => ({
  cambiarAutorizacion: vi.fn(),
  historialAutorizaciones: vi.fn(),
  listarVinculaciones: vi.fn(),
}));

const vinculacion = {
  id: 'vinculacion-1',
  estado: 'aceptada',
  escolar: { id: 'escolar-1', nombre_uso: 'Escolar Uno', email: 'escolar@example.test' },
  autorizacion: { autorizado: true },
};

beforeEach(() => {
  vi.resetAllMocks();
  listarVinculaciones.mockResolvedValue({ vinculaciones: [vinculacion] });
  historialAutorizaciones.mockResolvedValue({ autorizaciones: [] });
  cambiarAutorizacion.mockResolvedValue({
    mensaje: 'Se retiró el permiso para pedir tutorías.',
  });
});

afterEach(cleanup);

describe('B-02: éxito anunciado con role="alert"', () => {
  it('anuncia el cambio de autorización como estado, no como alerta', async () => {
    render(
      <MemoryRouter>
        <PanelEscolares />
      </MemoryRouter>
    );

    fireEvent.change(await screen.findByLabelText('Escolar'), {
      target: { value: 'escolar-1' },
    });
    fireEvent.click(await screen.findByRole('button', { name: 'Revocar autorización' }));

    const confirmacion = await screen.findByText('Se retiró el permiso para pedir tutorías.');
    expect(confirmacion.getAttribute('role')).toBe('status');
  });
});
