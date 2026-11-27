// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../../application/acceso/coordinacion.js', () => ({
  listarCoordinadores: vi.fn(),
  crearCoordinador: vi.fn(),
  eliminarCoordinador: vi.fn(),
}));

import {
  crearCoordinador,
  eliminarCoordinador,
  listarCoordinadores,
} from '../../../application/acceso/coordinacion.js';
import GestionarCoordinadores from './GestionarCoordinadores.jsx';

const COORDINADOR = { id: 'u-1', nombre_uso: 'Coordinador Uno', email: 'coord@example.test', facultad_nombre: 'Ingeniería' };

beforeEach(() => {
  vi.clearAllMocks();
  listarCoordinadores.mockResolvedValue({ coordinadores: [COORDINADOR] });
  crearCoordinador.mockResolvedValue({ id: 'u-2' });
  eliminarCoordinador.mockResolvedValue(undefined);
  window.confirm = vi.fn(() => true);
});

afterEach(cleanup);

describe('Gestionar coordinadores', () => {
  it('lista los coordinadores activos', async () => {
    render(
      <MemoryRouter>
        <GestionarCoordinadores />
      </MemoryRouter>
    );
    expect(await screen.findByText('Coordinador Uno')).toBeTruthy();
    expect(screen.getByText(/Ingeniería/)).toBeTruthy();
    expect(listarCoordinadores).toHaveBeenCalledTimes(1);
  });

  it('crea un coordinador y recarga la lista', async () => {
    render(
      <MemoryRouter>
        <GestionarCoordinadores />
      </MemoryRouter>
    );
    await screen.findByText('Coordinador Uno');
    fireEvent.change(screen.getByLabelText('Correo'), { target: { value: 'nuevo@example.test' } });
    fireEvent.change(screen.getByLabelText('Nombre de uso'), { target: { value: 'Nuevo Coordinador' } });
    fireEvent.change(screen.getByLabelText('Clave inicial'), { target: { value: 'Clave-Segura-2026!' } });
    fireEvent.click(screen.getByRole('button', { name: 'Crear coordinador' }));

    await waitFor(() => expect(crearCoordinador).toHaveBeenCalledWith({
      email: 'nuevo@example.test', nombre_uso: 'Nuevo Coordinador', clave: 'Clave-Segura-2026!',
    }));
    expect(await screen.findByText('Coordinador creado.')).toBeTruthy();
    expect(listarCoordinadores).toHaveBeenCalledTimes(2);
  });

  it('elimina un coordinador confirmado', async () => {
    render(
      <MemoryRouter>
        <GestionarCoordinadores />
      </MemoryRouter>
    );
    await screen.findByText('Coordinador Uno');
    fireEvent.click(screen.getByRole('button', { name: 'Eliminar' }));
    expect(window.confirm).toHaveBeenCalled();
    await waitFor(() => expect(eliminarCoordinador).toHaveBeenCalledWith('u-1'));
    expect(await screen.findByText('Coordinador eliminado.')).toBeTruthy();
  });
});
