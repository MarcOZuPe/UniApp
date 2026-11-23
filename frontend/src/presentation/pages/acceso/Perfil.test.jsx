// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import Perfil from './Perfil.jsx';
import { cambiarClavePerfil } from '../../../application/acceso/perfil.js';

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

function renderizar() {
  return render(
    <MemoryRouter>
      <Perfil />
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
  cambiarClavePerfil.mockResolvedValue(null);
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
    expect(await screen.findByText('Contraseña actualizada.')).toBeTruthy();
  });
});
