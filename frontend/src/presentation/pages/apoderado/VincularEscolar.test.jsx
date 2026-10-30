// @vitest-environment jsdom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import VincularEscolar from './VincularEscolar.jsx';
import { solicitarVinculacion } from '../../../adapters/http/apoderadoApi.js';

vi.mock('../../../adapters/http/apoderadoApi.js', () => ({
  solicitarVinculacion: vi.fn(),
}));

const correoEscolar = 'escolar@example.com';

describe('I06: vincular escolar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it('envía el correo del escolar', async () => {
    solicitarVinculacion.mockResolvedValue({ vinculacion: { id: 'v-1' } });

    renderizarVincularEscolar();

    expect(
      await screen.findByRole('heading', { name: 'Vincular a un escolar' })
    ).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Ir al panel de escolares' })).toBeTruthy();

    const campoCorreoEscolar = screen.getByLabelText('Correo del escolar');
    fireEvent.change(campoCorreoEscolar, {
      target: { value: correoEscolar },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Enviar solicitud' }));

    expect(await screen.findByText(/Solicitud enviada/)).toBeTruthy();
    expect(solicitarVinculacion).toHaveBeenCalledWith(correoEscolar);
  });

  it('muestra el error de validación del backend', async () => {
    solicitarVinculacion.mockRejectedValue({
      message: 'Revisa los datos ingresados.',
      campos: { email_escolar: 'el correo es obligatorio' },
    });

    renderizarVincularEscolar();

    const campoCorreoEscolar = await screen.findByLabelText('Correo del escolar');
    fireEvent.change(campoCorreoEscolar, {
      target: { value: correoEscolar },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Enviar solicitud' }));

    expect(await screen.findByText('el correo es obligatorio')).toBeTruthy();
    expect(screen.getAllByRole('alert').length).toBeGreaterThan(0);
  });
});

function renderizarVincularEscolar() {
  return render(
    <MemoryRouter>
      <VincularEscolar />
    </MemoryRouter>
  );
}
