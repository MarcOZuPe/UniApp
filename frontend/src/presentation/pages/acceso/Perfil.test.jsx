// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';

import Perfil from './Perfil.jsx';
import { cambiarClavePerfil } from '../../../application/acceso/perfil.js';
import { salir } from '../../../application/acceso/sesion.js';

const estado = vi.hoisted(() => ({
  usuario: {
    id: 'escolar-1',
    email: 'escolar@example.test',
    nombre_uso: 'Escolar Prueba',
    rol: 'escolar',
    correo_confirmado: true,
  },
}));

vi.mock('../../hooks/useSesion.js', () => ({
  useSesion: () => ({ usuario: estado.usuario }),
}));

vi.mock('../../../application/acceso/perfil.js', () => ({
  cambiarClavePerfil: vi.fn(),
}));

vi.mock('../../../application/acceso/sesion.js', () => ({
  salir: vi.fn(),
}));

function DestinoIngreso() {
  const ubicacion = useLocation();
  return (
    <>
      <h1>Ingresar</h1>
      {ubicacion.state?.mensaje && <p role="status">{ubicacion.state.mensaje}</p>}
    </>
  );
}

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/perfil']}>
      <Routes>
        <Route path="/perfil" element={<Perfil />} />
        <Route path="/ingreso" element={<DestinoIngreso />} />
      </Routes>
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
  cambiarClavePerfil.mockResolvedValue(null);
  salir.mockResolvedValue(undefined);
});

afterEach(cleanup);

describe('perfil de cuenta', () => {
  it('muestra datos no editables y permite cambiar la contraseña', async () => {
    renderizar();

    expect(screen.getByText('Escolar Prueba')).toBeTruthy();
    expect(screen.getByText('escolar@example.test')).toBeTruthy();
    expect(screen.queryByLabelText('Correo')).toBeNull();
    expect(screen.queryByLabelText('Nombre de identificación')).toBeNull();

    fireEvent.change(screen.getByLabelText('Contraseña actual'), {
      target: { value: 'Clave-Vieja-2026!' },
    });
    fireEvent.change(screen.getByLabelText('Nueva contraseña'), {
      target: { value: 'Clave-Nueva-2026!' },
    });
    fireEvent.change(screen.getByLabelText('Repite la nueva contraseña'), {
      target: { value: 'Clave-Nueva-2026!' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Cambiar contraseña' }));

    await waitFor(() =>
      expect(cambiarClavePerfil).toHaveBeenCalledWith({
        clave_actual: 'Clave-Vieja-2026!',
        clave_nueva: 'Clave-Nueva-2026!',
      })
    );
    expect(salir).toHaveBeenCalled();
    expect(
      await screen.findByText('Contraseña actualizada. Por seguridad, inicia sesión nuevamente.')
    ).toBeTruthy();
  });
});
