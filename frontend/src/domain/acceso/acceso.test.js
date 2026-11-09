import { describe, expect, it, vi } from 'vitest';
import { ROLES, ROLES_REGISTRO, etiquetaRol, politicaDeRol, puedeAcceder, requiereCorreoInstitucional } from './roles.js';
import { crearPolitica, validarRegistro, validarRestablecer } from './validadores.js';
import { getInstance } from '../../application/acceso/almacenSesion.js';
import { crearClienteHttp, ErrorApi } from '../../infrastructure/http/fetchClient.js';
import { conSesionExpirada } from '../../adapters/http/decoradores.js';

describe('modelo: catálogo de roles (registro centralizado)', () => {
  it('deriva los roles públicos y expone sus metadatos', () => {
    expect(ROLES_REGISTRO).toEqual(['escolar', 'apoderado', 'tutor']);
    expect(ROLES_REGISTRO).not.toContain('coordinador');
    for (const rol of Object.keys(ROLES)) {
      expect(etiquetaRol(rol)).toBeTruthy();
      expect(['basica', 'estricta']).toContain(politicaDeRol(rol));
    }
    expect(etiquetaRol('desconocido')).toBe('desconocido');
    expect(politicaDeRol('desconocido')).toBe('basica');
    expect(requiereCorreoInstitucional('tutor')).toBe(true);
    expect(requiereCorreoInstitucional('escolar')).toBe(false);
    expect(requiereCorreoInstitucional('desconocido')).toBe(false);
  });
});

describe('modelo: validadores (Strategy + Factory)', () => {
  it('tutor usa política estricta y escolar la básica', () => {
    expect(crearPolitica('escolar')({ clave: 'larga-pero-simple' })).toEqual([]);
    expect(crearPolitica('tutor')({ clave: 'larga-pero-simple' }).length).toBeGreaterThan(0);
  });
  it('tutor exige correo institucional', () => {
    const datosBaseTutor = { nombre_uso: 'Tito', clave: 'Clave-Segura-2026!', confirmar: 'Clave-Segura-2026!', rol: 'tutor' };
    expect(validarRegistro({ ...datosBaseTutor, email: 't@gmail.com' }).email).toBeTruthy();
    expect(validarRegistro({ ...datosBaseTutor, email: 't@uni.edu.pe.evil.com' }).email).toBeTruthy();
    expect(validarRegistro({ ...datosBaseTutor, email: 't@uni.edu.pe' })).toEqual({});
    expect(validarRegistro({ ...datosBaseTutor, email: 't@uni.pe' })).toEqual({});
  });
  it('restablecer exige la política básica y confirma la clave', () => {
    expect(validarRestablecer({ clave: 'corta', confirmar: 'corta' }).clave).toBeTruthy();
    expect(validarRestablecer({ clave: 'clave-larga-ok', confirmar: 'clave-larga-ok' })).toEqual({});
    expect(validarRestablecer({ clave: 'clave-larga-ok', confirmar: 'otra' }).confirmar).toBeTruthy();
  });
  it('detecta contraseñas distintas', () => {
    const erroresRegistro = validarRegistro({ email: 'a@b.co', nombre_uso: 'Ana', clave: 'clave-larga-ok', confirmar: 'otra', rol: 'escolar' });
    expect(erroresRegistro.confirmar).toBeTruthy();
  });
  it('un escolar no puede acceder a rutas de coordinación', () => {
    expect(puedeAcceder('escolar', ['coordinador'])).toBe(false);
    expect(puedeAcceder('coordinador', ['coordinador'])).toBe(true);
    expect(puedeAcceder('escolar', undefined)).toBe(true);
  });
});

describe('controlador: almacén (Observer)', () => {
  it('devuelve siempre la misma instancia', () => {
    expect(getInstance()).toBe(getInstance());
  });

  it('notifica a suscriptores y permite darse de baja', () => {
    const almacen = getInstance();
    const observador = vi.fn();
    const desuscribir = almacen.subscribe(observador);
    almacen.set({ usuario: { rol: 'escolar' }, cargando: false });
    expect(observador).toHaveBeenCalledTimes(1);
    desuscribir();
    almacen.set({ usuario: null });
    expect(observador).toHaveBeenCalledTimes(1);
  });
});

describe('controlador: cliente (Adapter + Decorator)', () => {
  it('normaliza errores HTTP y de red en ErrorApi', async () => {
    const cliente = crearClienteHttp(async () => crearRespuestaHttp(422, { codigo: 'validacion', mensaje: 'Revisa', campos: { clave: 'x' } }));
    await expect(cliente.post('/x', {})).rejects.toMatchObject({ estado: 422, campos: { clave: 'x' } });
    const clienteCaido = crearClienteHttp(async () => { throw new Error('boom'); });
    await expect(clienteCaido.get('/x')).rejects.toMatchObject({ estado: 0, codigo: 'red' });
  });

  it('el decorador conserva postFormulario del cliente decorado', async () => {
    const formulario = new FormData();
    const clienteBase = {
      get: vi.fn(),
      post: vi.fn(),
      postFormulario: vi.fn(async () => ({ adjunto_id: 'a-1' })),
      put: vi.fn(),
      delete: vi.fn(),
    };
    const clienteDecorado = conSesionExpirada(clienteBase, vi.fn());
    await expect(clienteDecorado.postFormulario('/api/reservas/r-1/adjuntos', formulario)).resolves.toEqual({ adjunto_id: 'a-1' });
    expect(clienteBase.postFormulario).toHaveBeenCalledWith('/api/reservas/r-1/adjuntos', formulario);
  });

  it('el decorador avisa cuando una ruta protegida devuelve 401, pero no en el ingreso', async () => {
    const clienteBase = crearClienteHttp(async () => crearRespuestaHttp(401, { mensaje: 'no' }));
    const alExpirar = vi.fn();
    const clienteDecorado = conSesionExpirada(clienteBase, alExpirar);
    await expect(clienteDecorado.get('/api/auth/yo')).rejects.toBeInstanceOf(ErrorApi);
    expect(alExpirar).toHaveBeenCalledTimes(1);
    await expect(clienteDecorado.post('/api/auth/ingreso', {})).rejects.toBeInstanceOf(ErrorApi);
    expect(alExpirar).toHaveBeenCalledTimes(1);
  });
});

describe('controlador: acciones de confirmación (Adapter)', () => {
  it('envía el token y el correo a los endpoints correctos', async () => {
    const llamadasHttp = [];
    vi.stubGlobal('fetch', vi.fn(async (url, opciones) => {
      llamadasHttp.push({ url, opciones });
      return { ok: true, status: 200, text: async () => JSON.stringify({ mensaje: 'ok' }) };
    }));
    const { confirmarCorreo, reenviarConfirmacion } = await import('../../application/acceso/confirmacion.js');
    await confirmarCorreo('tok-123');
    await reenviarConfirmacion('ana@uni.pe');
    vi.unstubAllGlobals();
    expect(llamadasHttp.map((llamadaHttp) => llamadaHttp.url)).toEqual([
      '/api/auth/confirmar',
      '/api/auth/reenviar-confirmacion',
    ]);
    expect(JSON.parse(llamadasHttp[0].opciones.body)).toEqual({ token: 'tok-123' });
    expect(JSON.parse(llamadasHttp[1].opciones.body)).toEqual({ email: 'ana@uni.pe' });
    expect(llamadasHttp[0].opciones.credentials).toBe('same-origin');
  });
});

describe('controlador: panel de coordinación (Adapter)', () => {
  it('consulta el endpoint del panel con la sesión', async () => {
    const llamadasHttp = [];
    vi.stubGlobal('fetch', vi.fn(async (url, opciones) => {
      llamadasHttp.push({ url, opciones });
      return { ok: true, status: 200, text: async () => JSON.stringify({ mensaje: 'Panel de coordinación' }) };
    }));
    const { obtenerPanelCoordinacion } = await import('../../application/acceso/coordinacion.js');
    await obtenerPanelCoordinacion();
    vi.unstubAllGlobals();
    expect(llamadasHttp).toHaveLength(1);
    expect(llamadasHttp[0].url).toBe('/api/coordinacion/panel');
    expect(llamadasHttp[0].opciones.method).toBe('GET');
    expect(llamadasHttp[0].opciones.credentials).toBe('same-origin');
  });
});

function crearRespuestaHttp(estado, cuerpo) {
  return { ok: estado < 400, status: estado, text: async () => (cuerpo ? JSON.stringify(cuerpo) : '') };
}
