import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { api } from '../../adapters/http/apiAcceso.js';
import { almacenSesion } from './almacenSesion.js';
import { cargarSesion } from './sesion.js';

vi.mock('../../adapters/http/apiAcceso.js', () => ({
  api: { get: vi.fn(), post: vi.fn() },
}));

const usuario = { id: 'u-1', email: 'u@example.test', nombre_uso: 'U', rol: 'escolar' };

describe('F-19: carga inicial de sesión', () => {
  beforeEach(() => {
    api.get.mockReset();
    almacenSesion.set({ usuario, cargando: false });
  });

  afterEach(() => {
    almacenSesion.set({ usuario: null, cargando: false });
  });

  it('debe conservar el usuario local ante un fallo que no es 401', async () => {
    api.get.mockRejectedValue(Object.assign(new Error('fallo del servidor'), { estado: 500 }));

    await cargarSesion();

    expect(api.get).toHaveBeenCalledWith('/api/auth/yo');
    expect(almacenSesion.getState().usuario).not.toBeNull();
    expect(almacenSesion.getState().usuario).toEqual(usuario);
  });

  it('debe conservar el usuario local ante un error de red', async () => {
    api.get.mockRejectedValue(Object.assign(new Error('sin conexión'), { estado: 0 }));

    await cargarSesion();

    expect(almacenSesion.getState().usuario).toEqual(usuario);
  });
});
