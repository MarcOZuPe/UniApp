// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import FormularioReserva from './FormularioReserva.jsx';
import {
  crearReserva,
  listarSlots,
  obtenerAutorizacionReserva,
  subirAdjunto,
  urlAdjunto,
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

const reservaCreada = {
  id: 'reserva-1',
  curso: 'Álgebra',
  tutor_nombre: 'Tutor Uno',
  inicio_lima: '2026-11-24 10:00',
  tema: 'Repaso de factorización',
  grado: '5to de secundaria',
};

function renderizar() {
  return render(
    <MemoryRouter>
      <FormularioReserva />
    </MemoryRouter>
  );
}

function completarFormulario() {
  fireEvent.change(screen.getByLabelText('Horario'), { target: { value: 'slot-1' } });
  fireEvent.change(screen.getByLabelText('Tema'), {
    target: { value: 'Repaso de factorización' },
  });
  fireEvent.change(screen.getByLabelText('Objetivo'), {
    target: { value: 'Resolver ejercicios de factorización de trinomios sin ayuda del tutor.' },
  });
  fireEvent.change(screen.getByLabelText('Grado'), { target: { value: '5to de secundaria' } });
}

beforeEach(() => {
  vi.resetAllMocks();
  obtenerAutorizacionReserva.mockResolvedValue({ autorizado: true });
  listarSlots.mockResolvedValue({ slots: [slot] });
  crearReserva.mockResolvedValue(reservaCreada);
  subirAdjunto.mockResolvedValue({ id: 'adjunto-1', nombre: 'guia.pdf', estado: 'validado' });
  urlAdjunto.mockImplementation((id, adjuntoId) => `/api/reserva/${id}/adjuntos/${adjuntoId}`);
});

afterEach(cleanup);

describe('F-02: idempotencia de la reserva', () => {
  it('reutiliza la clave de idempotencia al reintentar la misma reserva', async () => {
    crearReserva
      .mockRejectedValueOnce(new Error('No hay conexión con el servidor.'))
      .mockResolvedValueOnce(reservaCreada);

    renderizar();
    await screen.findByRole('option', { name: /Álgebra/ });
    completarFormulario();

    fireEvent.click(screen.getByRole('button', { name: 'Confirmar reserva' }));
    expect((await screen.findByRole('alert')).textContent).toContain(
      'No hay conexión con el servidor.'
    );

    fireEvent.click(screen.getByRole('button', { name: 'Confirmar reserva' }));
    await waitFor(() => expect(crearReserva).toHaveBeenCalledTimes(2));

    const primeraClave = crearReserva.mock.calls[0][1];
    const segundaClave = crearReserva.mock.calls[1][1];
    expect(primeraClave).toBeTruthy();
    expect(segundaClave).toBe(primeraClave);
  });
});

describe('F-24: tipos de adjuntos en reserva', () => {
  it('limita los tipos de archivo aceptados en el input de adjuntos', async () => {
    renderizar();
    await screen.findByLabelText('Horario');

    const input = screen.getByLabelText('Adjuntos (opcional)');
    expect(input.getAttribute('accept')).toBeTruthy();
    expect(input.getAttribute('accept')).toContain('application/pdf');
  });
});
