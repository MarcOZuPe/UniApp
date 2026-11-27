import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../adapters/http/apiAcceso.js', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    postFormulario: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

import { api } from '../../adapters/http/apiAcceso.js';
import {
  crearGeneral,
  eliminarGeneral,
  listarGenerales,
  listarLogsAdministracion,
} from './administracion.js';
import {
  crearCoordinador,
  eliminarCoordinador,
  listarAuditoriaCoordinacion,
  listarCoordinadores,
  obtenerMetricasCoordinacion,
  obtenerPanelCoordinacion,
} from './coordinacion.js';
import { cambiarClavePerfil } from './perfil.js';
import { restablecerClave, solicitarRecuperacion } from './recuperacion.js';
import { listarFacultades, registrar } from './registro.js';
import { confirmarCorreo, reenviarConfirmacion } from './confirmacion.js';

beforeEach(() => {
  vi.clearAllMocks();
});

describe('servicios de administración y coordinación', () => {
  it('delega las operaciones en el cliente HTTP', () => {
    listarGenerales();
    crearGeneral({ email: 'a@b.test' });
    eliminarGeneral('id/1');
    listarLogsAdministracion();
    listarLogsAdministracion(25);
    obtenerPanelCoordinacion();
    obtenerMetricasCoordinacion();
    listarAuditoriaCoordinacion();
    listarAuditoriaCoordinacion(10);
    listarCoordinadores();
    crearCoordinador({ email: 'c@b.test' });
    eliminarCoordinador('id/2');

    expect(api.get).toHaveBeenCalledWith('/api/administracion/coordinadores-generales');
    expect(api.post).toHaveBeenCalledWith('/api/administracion/coordinadores-generales', { email: 'a@b.test' });
    expect(api.delete).toHaveBeenCalledWith('/api/administracion/coordinadores-generales/id%2F1');
    expect(api.get).toHaveBeenCalledWith('/api/administracion/logs?limite=100');
    expect(api.get).toHaveBeenCalledWith('/api/administracion/logs?limite=25');
    expect(api.get).toHaveBeenCalledWith('/api/coordinacion/panel');
    expect(api.get).toHaveBeenCalledWith('/api/coordinacion/metricas');
    expect(api.get).toHaveBeenCalledWith('/api/coordinacion/auditoria?limite=50');
    expect(api.get).toHaveBeenCalledWith('/api/coordinacion/auditoria?limite=10');
    expect(api.get).toHaveBeenCalledWith('/api/coordinacion/coordinadores');
    expect(api.post).toHaveBeenCalledWith('/api/coordinacion/coordinadores', { email: 'c@b.test' });
    expect(api.delete).toHaveBeenCalledWith('/api/coordinacion/coordinadores/id%2F2');
  });
});

describe('servicios de cuenta', () => {
  it('delega perfil, recuperación, registro y confirmación', () => {
    cambiarClavePerfil({ clave_actual: 'a', clave_nueva: 'b' });
    solicitarRecuperacion('a@b.test');
    restablecerClave({ token: 't', clave: 'b' });
    registrar({
      email: 'a@b.test', nombre_uso: 'Nombre', clave: 'Clave-1234', rol: 'escolar',
      facultad_id: 'fac-1', campo_extra: 'ignorado',
    });
    listarFacultades();
    confirmarCorreo('t');
    reenviarConfirmacion('a@b.test');

    expect(api.put).toHaveBeenCalledWith('/api/perfil/clave', { clave_actual: 'a', clave_nueva: 'b' });
    expect(api.post).toHaveBeenCalledWith('/api/auth/recuperar', { email: 'a@b.test' });
    expect(api.post).toHaveBeenCalledWith('/api/auth/restablecer', { token: 't', clave: 'b' });
    expect(api.post).toHaveBeenCalledWith('/api/auth/registro', {
      email: 'a@b.test', nombre_uso: 'Nombre', clave: 'Clave-1234', rol: 'escolar', facultad_id: 'fac-1',
    });
    expect(api.get).toHaveBeenCalledWith('/api/facultades');
    expect(api.post).toHaveBeenCalledWith('/api/auth/confirmar', { token: 't' });
    expect(api.post).toHaveBeenCalledWith('/api/auth/reenviar-confirmacion', { email: 'a@b.test' });
  });
});
