// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import GestionarCoordinadores from './GestionarCoordinadores.jsx';
import {
  crearCoordinador,
  eliminarCoordinador,
  listarCoordinadores,
} from '../../../application/acceso/coordinacion.js';

vi.mock('../../../application/acceso/coordinacion.js', () => ({
  listarCoordinadores: vi.fn(),
  crearCoordinador: vi.fn(),
  eliminarCoordinador: vi.fn(),
}));

const coordinador = {
  id: 'coord-1',
  email: 'coord@example.test',
  nombre_uso: 'Coordinador Uno',
  rol: 'coordinador',
  correo_confirmado: true,
};

function renderizar() {
  return render(
    <MemoryRouter>
      <GestionarCoordinadores />
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
  listarCoordinadores.mockResolvedValue({ coordinadores: [coordinador] });
  crearCoordinador.mockResolvedValue({
    id: 'coord-2',
    email: 'nuevo@example.test',
    nombre_uso: 'Nuevo Coordinador',
    rol: 'coordinador',
    correo_confirmado: true,
  });
  eliminarCoordinador.mockResolvedValue(null);
  vi.stubGlobal('confirm', vi.fn(() => true));
});

afterEach(() => {
  vi.unstubAllGlobals();
  cleanup();
});

describe('I14 - gestión de coordinadores', () => {
  it('lista, crea y elimina coordinadores', async () => {
    renderizar();

    expect(await screen.findByText('Coordinador Uno')).toBeTruthy();
    expect(screen.getByText('coord@example.test')).toBeTruthy();

    fireEvent.change(screen.getByLabelText('Correo'), {
      target: { value: 'nuevo@example.test' },
    });
    fireEvent.change(screen.getByLabelText('Nombre de uso'), {
      target: { value: 'Nuevo Coordinador' },
    });
    fireEvent.change(screen.getByLabelText('Clave inicial'), {
      target: { value: 'Clave-Segura-2026!' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Crear coordinador' }));

    await waitFor(() =>
      expect(crearCoordinador).toHaveBeenCalledWith({
        email: 'nuevo@example.test',
        nombre_uso: 'Nuevo Coordinador',
        clave: 'Clave-Segura-2026!',
      })
    );
    expect(await screen.findByText('Coordinador creado.')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Eliminar' }));
    await waitFor(() => expect(eliminarCoordinador).toHaveBeenCalledWith('coord-1'));
    expect(await screen.findByText('Coordinador eliminado.')).toBeTruthy();
  });
});
