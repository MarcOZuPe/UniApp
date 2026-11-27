import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('./apiAcceso.js', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    postFormulario: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

import { api } from './apiAcceso.js';
import {
  aceptarVinculacion,
  cambiarAutorizacion,
  consultarVinculacion,
  historialAutorizaciones,
  listarVinculaciones,
  miApoderado,
  solicitarVinculacion,
} from './apoderadoApi.js';
import {
  importarCatalogo,
  obtenerCatalogo,
  obtenerCursosPendientes,
  obtenerSolicitudesColegio,
  revisarCurso,
  solicitarColegio,
} from './catalogoApi.js';
import {
  obtenerAprobados,
  obtenerMiExpediente,
  obtenerPendientes,
  obtenerPerfilesPublicos,
  presentarExpediente,
  revisarExpediente,
} from './tutorApi.js';

beforeEach(() => {
  vi.clearAllMocks();
});

describe('adaptadores de apoderado', () => {
  it('construye las rutas y cuerpos esperados', () => {
    solicitarVinculacion('escolar@example.test');
    listarVinculaciones();
    cambiarAutorizacion('escolar/1', false);
    historialAutorizaciones('escolar/1');
    consultarVinculacion('token');
    aceptarVinculacion('token');
    miApoderado();

    expect(api.post).toHaveBeenCalledWith('/api/apoderado/vinculaciones', { email_escolar: 'escolar@example.test' });
    expect(api.get).toHaveBeenCalledWith('/api/apoderado/vinculaciones');
    expect(api.post).toHaveBeenCalledWith('/api/apoderado/escolares/escolar%2F1/autorizacion', { autorizado: false });
    expect(api.get).toHaveBeenCalledWith('/api/apoderado/escolares/escolar%2F1/autorizaciones');
    expect(api.post).toHaveBeenCalledWith('/api/escolar/vinculaciones/consultar', { token: 'token' });
    expect(api.post).toHaveBeenCalledWith('/api/escolar/vinculaciones/aceptar', { token: 'token' });
    expect(api.get).toHaveBeenCalledWith('/api/escolar/apoderado');
  });
});

describe('adaptadores de catálogo', () => {
  it('construye las rutas esperadas', () => {
    obtenerCatalogo();
    importarCatalogo({ fuente: 'origen' });
    obtenerCursosPendientes();
    revisarCurso(7, { estado: 'aprobado' });
    solicitarColegio({ nombre: 'Colegio' });
    obtenerSolicitudesColegio();

    expect(api.get).toHaveBeenCalledWith('/api/catalogo');
    expect(api.post).toHaveBeenCalledWith('/api/catalogo/importaciones', { fuente: 'origen' });
    expect(api.get).toHaveBeenCalledWith('/api/catalogo/cursos/pendientes');
    expect(api.post).toHaveBeenCalledWith('/api/catalogo/cursos/7/revision', { estado: 'aprobado' });
    expect(api.post).toHaveBeenCalledWith('/api/catalogo/colegios/solicitudes', { nombre: 'Colegio' });
    expect(api.get).toHaveBeenCalledWith('/api/catalogo/colegios/solicitudes');
  });
});

describe('adaptadores de tutor', () => {
  it('construye las rutas y el formulario esperado', () => {
    const formulario = new FormData();
    obtenerMiExpediente();
    obtenerPerfilesPublicos();
    presentarExpediente(formulario);
    obtenerPendientes();
    obtenerAprobados();
    revisarExpediente('tutor/1', { estado: 'aprobado' });

    expect(api.get).toHaveBeenCalledWith('/api/tutor/mi-expediente');
    expect(api.get).toHaveBeenCalledWith('/api/tutor/perfiles');
    expect(api.postFormulario).toHaveBeenCalledWith('/api/tutor/expediente', formulario);
    expect(api.get).toHaveBeenCalledWith('/api/tutor/revision');
    expect(api.get).toHaveBeenCalledWith('/api/tutor/revision/aprobados');
    expect(api.post).toHaveBeenCalledWith('/api/tutor/tutor%2F1/revision', { estado: 'aprobado' });
  });
});
