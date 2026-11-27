// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

vi.mock('../../../application/acceso/sesion.js', () => ({ ingresar: vi.fn() }));

import { ingresar } from '../../../application/acceso/sesion.js';
import Ingreso from './Ingreso.jsx';

beforeEach(() => {
  vi.clearAllMocks();
  ingresar.mockResolvedValue({});
});

afterEach(cleanup);

describe('Ingreso', () => {
  it('muestra el mensaje de redirección', () => {
    render(
      <MemoryRouter>
        <Ingreso />
      </MemoryRouter>
    );
    expect(screen.getByRole('heading', { name: 'Ingresar' })).toBeTruthy();
  });

  it('valida antes de enviar', () => {
    render(
      <MemoryRouter>
        <Ingreso />
      </MemoryRouter>
    );
    fireEvent.click(screen.getByRole('button', { name: 'Ingresar' }));
    expect(ingresar).not.toHaveBeenCalled();
    expect(screen.getAllByRole('alert').length).toBeGreaterThan(0);
  });

  it('ingresa y navega al destino', async () => {
    render(
      <MemoryRouter initialEntries={['/ingreso?next=/agenda']}>
        <Routes>
          <Route path="/ingreso" element={<Ingreso />} />
          <Route path="/agenda" element={<p>agenda</p>} />
        </Routes>
      </MemoryRouter>
    );
    fireEvent.change(screen.getByLabelText('Correo'), { target: { value: 'escolar@example.test' } });
    fireEvent.change(screen.getByLabelText('Contraseña'), { target: { value: 'Clave-Segura-2026!' } });
    fireEvent.click(screen.getByRole('button', { name: 'Ingresar' }));
    await waitFor(() => expect(ingresar).toHaveBeenCalledWith({
      email: 'escolar@example.test', clave: 'Clave-Segura-2026!',
    }));
    expect(await screen.findByText('agenda')).toBeTruthy();
  });
});
