// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import Recuperar from './Recuperar.jsx';
import { solicitarRecuperacion } from '../../../application/acceso/recuperacion.js';

vi.mock('../../../application/acceso/recuperacion.js', () => ({
  solicitarRecuperacion: vi.fn(),
}));

beforeEach(() => {
  vi.resetAllMocks();
  solicitarRecuperacion.mockReturnValue(new Promise(() => {}));
});

afterEach(cleanup);

describe('F-25: doble envío en recuperación', () => {
  it('no debe solicitar dos recuperaciones si se hace doble clic', () => {
    render(
      <MemoryRouter>
        <Recuperar />
      </MemoryRouter>
    );
    fireEvent.change(screen.getByLabelText('Correo'), {
      target: { value: 'escolar@example.test' },
    });

    const boton = screen.getByRole('button', { name: 'Enviar enlace' });
    fireEvent.click(boton);
    fireEvent.click(boton);

    expect(solicitarRecuperacion).toHaveBeenCalledTimes(1);
    expect(screen.getByRole('button', { name: 'Enviar enlace' }).disabled).toBe(true);
  });
});
