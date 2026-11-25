// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import VincularEscolar from './VincularEscolar.jsx';
import { solicitarVinculacion } from '../../../adapters/http/apoderadoApi.js';

vi.mock('../../../adapters/http/apoderadoApi.js', () => ({
  solicitarVinculacion: vi.fn(),
}));

beforeEach(() => {
  vi.resetAllMocks();
  solicitarVinculacion.mockResolvedValue({});
});

afterEach(cleanup);

describe('B-02: éxito anunciado con role="alert"', () => {
  it('anuncia la solicitud enviada como estado, no como alerta', async () => {
    render(
      <MemoryRouter>
        <VincularEscolar />
      </MemoryRouter>
    );

    fireEvent.change(screen.getByLabelText('Correo del escolar'), {
      target: { value: 'escolar@example.test' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Enviar solicitud' }));

    const confirmacion = await screen.findByText(/Solicitud enviada/);
    expect(confirmacion.getAttribute('role')).toBe('status');
  });
});
