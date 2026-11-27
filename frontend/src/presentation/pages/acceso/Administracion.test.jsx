// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import Administracion from './Administracion.jsx';
import {
  crearGeneral,
  eliminarGeneral,
  listarGenerales,
  listarLogsAdministracion,
} from '../../../application/acceso/administracion.js';
import { listarFacultades } from '../../../application/acceso/registro.js';

vi.mock('../../../application/acceso/administracion.js', () => ({
  crearGeneral: vi.fn(),
  eliminarGeneral: vi.fn(),
  listarGenerales: vi.fn(),
  listarLogsAdministracion: vi.fn(),
}));

vi.mock('../../../application/acceso/registro.js', () => ({
  listarFacultades: vi.fn(),
}));

const FACULTAD = { id: 'fac-1', nombre: 'Ingeniería' };
const GENERAL = {
  id: 'u-1',
  nombre_uso: 'Coordinación General',
  email: 'general@example.test',
  facultad_nombre: 'Ingeniería',
};

beforeEach(() => {
  vi.resetAllMocks();
  listarGenerales.mockResolvedValue({ coordinadores_generales: [GENERAL] });
  listarFacultades.mockResolvedValue({ facultades: [FACULTAD] });
  listarLogsAdministracion.mockResolvedValue({ logs: [] });
  crearGeneral.mockResolvedValue({ id: 'u-2' });
  eliminarGeneral.mockResolvedValue(undefined);
});

afterEach(cleanup);

describe('panel de administración', () => {
  it('carga coordinadores, facultades y bitácora al iniciar', async () => {
    render(
      <MemoryRouter>
        <Administracion />
      </MemoryRouter>
    );

    expect(await screen.findByText('Coordinación General')).toBeTruthy();
    expect(screen.getByText(/Ingeniería/, { selector: 'small' })).toBeTruthy();
    expect(listarGenerales).toHaveBeenCalledTimes(1);
    expect(listarFacultades).toHaveBeenCalledTimes(1);
    expect(listarLogsAdministracion).toHaveBeenCalledTimes(1);
    expect(listarLogsAdministracion).toHaveBeenCalledWith(100);
  });

  it('muestra el error del backend y los errores por campo', async () => {
    crearGeneral.mockRejectedValue({
      message: 'Revisa los datos.',
      campos: { email: 'correo no válido' },
    });
    render(
      <MemoryRouter>
        <Administracion />
      </MemoryRouter>
    );
    await screen.findByText('Coordinación General');

    fireEvent.click(screen.getByRole('button', { name: 'Crear coordinador general' }));

    expect((await screen.findByRole('alert')).textContent).toContain('Revisa los datos.');
    expect(screen.getByText('correo no válido')).toBeTruthy();
  });

  it('F-20: un fallo parcial no debe ocultar los datos que sí cargaron', async () => {
    listarFacultades.mockRejectedValue(new Error('facultades caídas'));
    render(
      <MemoryRouter>
        <Administracion />
      </MemoryRouter>
    );

    expect(await screen.findByText('Coordinación General')).toBeTruthy();
    expect((await screen.findByRole('alert')).textContent).toContain('facultades caídas');
  });

  it('envía los datos del formulario y recarga la lista', async () => {
    render(
      <MemoryRouter>
        <Administracion />
      </MemoryRouter>
    );
    await screen.findByText('Coordinación General');

    fireEvent.change(screen.getByLabelText('Correo'), {
      target: { value: 'nuevo@example.test' },
    });
    fireEvent.change(screen.getByLabelText('Nombre de uso'), {
      target: { value: 'Nuevo General' },
    });
    fireEvent.change(screen.getByLabelText('Facultad'), {
      target: { value: FACULTAD.id },
    });
    fireEvent.change(screen.getByLabelText('Clave inicial'), {
      target: { value: 'Clave-Segura-2026!' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Crear coordinador general' }));

    await waitFor(() => {
      expect(crearGeneral).toHaveBeenCalledWith({
        email: 'nuevo@example.test',
        nombre_uso: 'Nuevo General',
        clave: 'Clave-Segura-2026!',
        facultad_id: 'fac-1',
      });
    });
    expect(await screen.findByText('Coordinador general creado.')).toBeTruthy();
    expect(listarGenerales).toHaveBeenCalledTimes(2);
  });
});
