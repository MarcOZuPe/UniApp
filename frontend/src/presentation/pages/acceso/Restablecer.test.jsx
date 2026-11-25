// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

import Restablecer from './Restablecer.jsx';
import { restablecerClave } from '../../../application/acceso/recuperacion.js';

vi.mock('../../../application/acceso/recuperacion.js', () => ({
  restablecerClave: vi.fn(),
}));

const CLAVE_NUEVA = 'Clave-Nueva-2026!';

beforeEach(() => {
  vi.resetAllMocks();
  restablecerClave.mockReturnValue(new Promise(() => {}));
});

afterEach(cleanup);

describe('F-25: doble envío en restablecimiento', () => {
  it('no debe consumir dos veces el token si se hace doble clic', () => {
    render(
      <MemoryRouter initialEntries={['/restablecer?token=token-secreto']}>
        <Routes>
          <Route path="/restablecer" element={<Restablecer />} />
        </Routes>
      </MemoryRouter>
    );
    fireEvent.change(screen.getByLabelText('Nueva contraseña'), { target: { value: CLAVE_NUEVA } });
    fireEvent.change(screen.getByLabelText('Repite la contraseña'), { target: { value: CLAVE_NUEVA } });

    const boton = screen.getByRole('button', { name: 'Guardar' });
    fireEvent.click(boton);
    fireEvent.click(boton);

    expect(restablecerClave).toHaveBeenCalledTimes(1);
    expect(screen.getByRole('button', { name: 'Guardar' }).disabled).toBe(true);
  });
});
