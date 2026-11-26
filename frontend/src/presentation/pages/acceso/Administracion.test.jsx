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
  listarGenerales: vi.fn(),
  crearGeneral: vi.fn(),
  eliminarGeneral: vi.fn(),
  listarLogsAdministracion: vi.fn(),
}));

vi.mock('../../../application/acceso/registro.js', () => ({
  listarFacultades: vi.fn(),
}));

const general = {
  id: 'general-1',
  email: 'general.ciencias@example.test',
  nombre_uso: 'General Ciencias',
  rol: 'coordinador_general',
  facultad_id: '1',
  facultad_nombre: 'Facultad de Ciencias',
  correo_confirmado: true,
};

const log = {
  id: 1,
  creado_en: '2026-11-23T19:00:00Z',
  categoria: 'seguridad',
  evento: 'clave.cambiada',
  actor_id: 'user-1',
  actor_rol: 'tutor',
  ip: '203.0.113.10',
  detalle: { sesiones_revocadas: true },
};

function renderizar() {
  return render(
    <MemoryRouter>
      <Administracion />
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
  listarGenerales.mockResolvedValue({ coordinadores_generales: [general] });
  listarFacultades.mockResolvedValue({
    facultades: [{ id: '2', nombre: 'Facultad de Ingeniería Mecánica' }],
  });
  listarLogsAdministracion.mockResolvedValue({ logs: [log] });
  crearGeneral.mockResolvedValue({ ...general, id: 'general-2' });
  eliminarGeneral.mockResolvedValue(null);
  vi.stubGlobal('confirm', vi.fn(() => true));
});

afterEach(() => {
  vi.unstubAllGlobals();
  cleanup();
});

describe('administración del sistema', () => {
  it('gestiona generales y muestra logs con IP', async () => {
    renderizar();

    expect(await screen.findByText('General Ciencias')).toBeTruthy();
    expect(screen.getByText('203.0.113.10')).toBeTruthy();
    expect(screen.getByText('Cambio de contraseña')).toBeTruthy();

    fireEvent.change(screen.getByLabelText('Correo'), {
      target: { value: 'general.mecanica@example.test' },
    });
    fireEvent.change(screen.getByLabelText('Nombre de uso'), {
      target: { value: 'General Mecánica' },
    });
    fireEvent.change(screen.getByLabelText('Facultad'), {
      target: { value: '2' },
    });
    fireEvent.change(screen.getByLabelText('Clave inicial'), {
      target: { value: 'Clave-Segura-2026!' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Crear coordinador general' }));

    await waitFor(() =>
      expect(crearGeneral).toHaveBeenCalledWith({
        email: 'general.mecanica@example.test',
        nombre_uso: 'General Mecánica',
        facultad_id: '2',
        clave: 'Clave-Segura-2026!',
      })
    );
    expect(await screen.findByText('Coordinador general creado.')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Eliminar' }));
    await waitFor(() => expect(eliminarGeneral).toHaveBeenCalledWith('general-1'));
  });
});
