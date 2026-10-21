// @vitest-environment jsdom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import VincularEscolar from './VincularEscolar.jsx';
import {
  cambiarAutorizacion,
  listarVinculaciones,
  solicitarVinculacion,
} from '../../controllers/apoderado/apiApoderado.js';

vi.mock('../../controllers/apoderado/apiApoderado.js', () => ({
  cambiarAutorizacion: vi.fn(),
  listarVinculaciones: vi.fn(),
  solicitarVinculacion: vi.fn(),
}));

function renderizar() {
  return render(
    <MemoryRouter>
      <VincularEscolar />
    </MemoryRouter>
  );
}

describe('I06: vincular escolar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it('envía el correo y muestra la solicitud pendiente', async () => {
    listarVinculaciones
      .mockResolvedValueOnce({ vinculaciones: [] })
      .mockResolvedValueOnce({
        vinculaciones: [
          {
            id: 'v-1',
            estado: 'pendiente',
            escolar: { nombre_uso: 'Escolar Prueba', email: 'escolar@example.com' },
          },
        ],
      });
    solicitarVinculacion.mockResolvedValue({ vinculacion: { id: 'v-1' } });

    renderizar();

    const campana = await screen.findByRole('heading', { name: 'Vincular a un escolar' });
    expect(campana).toBeTruthy();
    expect(await screen.findByText('Todavía no has vinculado a ningún escolar.')).toBeTruthy();

    fireEvent.change(screen.getByLabelText('Correo del escolar'), {
      target: { value: 'escolar@example.com' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Enviar solicitud' }));

    expect(await screen.findByText(/Solicitud enviada/)).toBeTruthy();
    expect(solicitarVinculacion).toHaveBeenCalledWith('escolar@example.com');
    expect(await screen.findByText('Pendiente de aceptación')).toBeTruthy();
  });

  it('permite revocar y volver a autorizar a un escolar aceptado', async () => {
    const aceptada = {
      id: 'v-1',
      estado: 'aceptada',
      escolar: { id: 'e-1', nombre_uso: 'Escolar Prueba', email: 'escolar@example.com' },
      autorizacion: { id: 'a-1', estado: 'autorizado', autorizado: true },
    };
    listarVinculaciones
      .mockResolvedValueOnce({ vinculaciones: [aceptada] })
      .mockResolvedValueOnce({
        vinculaciones: [
          {
            ...aceptada,
            autorizacion: { id: 'a-2', estado: 'no_autorizado', autorizado: false },
          },
        ],
      });
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

  it('muestra el error de validación del backend', async () => {
    listarVinculaciones.mockResolvedValue({ vinculaciones: [] });
    solicitarVinculacion.mockRejectedValue({
      message: 'Revisa los datos ingresados.',
      campos: { email_escolar: 'el correo es obligatorio' },
    });

    renderizar();

    fireEvent.change(await screen.findByLabelText('Correo del escolar'), {
      target: { value: 'escolar@example.com' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Enviar solicitud' }));

    expect(await screen.findByText('el correo es obligatorio')).toBeTruthy();
    expect(screen.getAllByRole('alert').length).toBeGreaterThan(0);
  });
});
