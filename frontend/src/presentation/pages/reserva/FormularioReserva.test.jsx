// @vitest-environment jsdom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import FormularioReserva from './FormularioReserva.jsx';
import {
  crearReserva,
  listarSlots,
  obtenerAutorizacionReserva,
} from '../../../adapters/http/reservaApi.js';

vi.mock('../../../adapters/http/reservaApi.js', () => ({
  obtenerAutorizacionReserva: vi.fn(),
  listarSlots: vi.fn(),
  crearReserva: vi.fn(),
  subirAdjunto: vi.fn(),
  urlAdjunto: vi.fn(),
}));

const slot = {
  id: 'slot-1',
  curso: 'Álgebra',
  inicio_lima: '2026-11-24 10:00',
  disponible: true,
};

function renderizarFormulario() {
  return render(
    <MemoryRouter>
      <FormularioReserva />
    </MemoryRouter>
  );
}

describe('I04: autorización para reservar', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    listarSlots.mockResolvedValue({ slots: [slot] });
  });

  afterEach(() => {
    cleanup();
  });

  it('oculta el formulario y avisa cuando no hay autorización vigente', async () => {
    obtenerAutorizacionReserva.mockResolvedValue({
      autorizado: false,
      mensaje: 'No tienes una autorización vigente para reservar tutorías. Pide a tu apoderado que la active.',
    });

    renderizarFormulario();

    expect((await screen.findByRole('alert')).textContent).toContain(
      'No tienes una autorización vigente para reservar tutorías',
    );
    expect(screen.queryByLabelText('Horario')).toBeNull();
    expect(screen.queryByRole('button', { name: 'Confirmar reserva' })).toBeNull();
    expect(listarSlots).not.toHaveBeenCalled();
  });

  it('muestra el formulario cuando la autorización está vigente', async () => {
    obtenerAutorizacionReserva.mockResolvedValue({ autorizado: true });

    renderizarFormulario();

    expect(await screen.findByLabelText('Horario')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Confirmar reserva' })).toBeTruthy();
    await waitFor(() => expect(listarSlots).toHaveBeenCalledTimes(1));
  });

  it('oculta el formulario si la autorización se revoca al intentar reservar', async () => {
    obtenerAutorizacionReserva.mockResolvedValue({ autorizado: true });
    crearReserva.mockRejectedValue(
      Object.assign(new Error('No tienes una autorización vigente para reservar tutorías.'), {
        estado: 403,
        codigo: 'sin_autorizacion',
        campos: {},
      }),
    );

    renderizarFormulario();

    await screen.findByRole('option', { name: /Álgebra/ });
    fireEvent.change(screen.getByLabelText('Horario'), { target: { value: 'slot-1' } });
    fireEvent.change(screen.getByLabelText('Tema'), {
      target: { value: 'Repaso de factorización' },
    });
    fireEvent.change(screen.getByLabelText('Objetivo'), {
      target: { value: 'Resolver ejercicios de factorización de trinomios sin ayuda del tutor.' },
    });
    fireEvent.change(screen.getByLabelText('Grado'), { target: { value: '5to de secundaria' } });
    fireEvent.click(screen.getByRole('button', { name: 'Confirmar reserva' }));

    await waitFor(() => expect(screen.queryByLabelText('Horario')).toBeNull());
    expect(screen.getByRole('alert').textContent).toContain(
      'No tienes una autorización vigente para reservar tutorías',
    );
  });
});
