// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import App from '../../../app/App.jsx';
import Inicio from '../acceso/Inicio.jsx';
import { obtenerSesion, declararAsistencia, reportarAusencia, registrarResultado } from '../../../adapters/http/sesionApi.js';
const estado = vi.hoisted(() => ({ usuario: null, cargando: false }));
vi.mock('../../hooks/useSesion.js', () => ({ useSesion: () => estado }));
vi.mock('../../../application/acceso/sesion.js', () => ({ cargarSesion: vi.fn(), salir: vi.fn() }));
vi.mock('../../../adapters/http/sesionApi.js', () => ({ obtenerSesion: vi.fn(), declararAsistencia: vi.fn(), reportarAusencia: vi.fn(), registrarResultado: vi.fn() }));
const sesion = { id: 'sesion-i05-demo', reserva_id: 'reserva-i05-demo', tutor_id: 'tutor-001', escolar_id: 'escolar-001', inicio: '2026-10-20T15:00:00Z', fin: '2026-10-20T16:00:00Z', asistencias: [] };
const asistencia = rol => ({ usuario_id: rol + '-001', rol, declarada_en: sesion.inicio });
function abrir(ruta = '/sesiones/sesion-i05-demo') { window.history.replaceState({}, '', ruta); return render(<App />); }
beforeEach(() => {
  vi.resetAllMocks();
  estado.usuario = { id: 'tutor-001', rol: 'tutor', nombre_uso: 'Tutor Prueba' };
  obtenerSesion.mockResolvedValue(sesion);
});
afterEach(() => { cleanup(); vi.unstubAllEnvs(); });
describe('Detalle I05 y acceso por ruta real', () => {
  it('muestra carga, participantes pendientes, horario de Lima y demostración', async () => {
    abrir();
    expect(screen.getByRole('status').textContent).toContain('Cargando');
    await screen.findByRole('heading', { name: 'Detalle de tutoría' });
    expect(screen.getByText('Tutor:')).toBeTruthy();
    expect(screen.getByText('Escolar:')).toBeTruthy();
    expect(screen.getAllByText('Pendiente')).toHaveLength(2);
    expect(screen.getByText(/Sesión sintética/)).toBeTruthy();
    const fecha = new Intl.DateTimeFormat('es-PE', { timeZone: 'America/Lima', dateStyle: 'medium', timeStyle: 'short' }).format(new Date(sesion.inicio));
    expect(screen.getByText(fecha, { exact: false })).toBeTruthy();
    expect(screen.getByText('60 minutos', { exact: false, selector: 'p' })).toBeTruthy();
    expect(obtenerSesion).toHaveBeenCalledWith(sesion.id);
  });
  it.each(['tutor', 'escolar'])('%s declara solo su asistencia y vuelve a cargarla', async rol => {
    estado.usuario = { id: rol + '-001', rol, nombre_uso: rol };
    const guardada = { ...sesion, asistencias: [asistencia(rol)] };
    obtenerSesion.mockResolvedValueOnce(sesion).mockResolvedValue(guardada);
    declararAsistencia.mockResolvedValue(asistencia(rol));
    const vista = abrir();
    fireEvent.click(await screen.findByRole('button', { name: 'Declarar mi asistencia' }));
    const boton = await screen.findByRole('button', { name: 'Asistencia ya declarada' });
    expect(boton.disabled).toBe(true);
    fireEvent.click(boton);
    expect(declararAsistencia).toHaveBeenCalledTimes(1);
    expect(declararAsistencia).toHaveBeenCalledWith(sesion.id);
    expect(await screen.findByText('Tu asistencia fue declarada correctamente.')).toBeTruthy();
    vista.unmount();
    abrir();
    expect((await screen.findByRole('button', { name: 'Asistencia ya declarada' })).disabled).toBe(true);
  });
  it('no duplica peticiones mientras se registra', async () => {
    declararAsistencia.mockImplementation(() => new Promise(() => {}));
    abrir();
    const boton = await screen.findByRole('button', { name: 'Declarar mi asistencia' });
    fireEvent.click(boton); fireEvent.click(boton);
    expect(boton.disabled).toBe(true);
    expect(declararAsistencia).toHaveBeenCalledTimes(1);
  });
  it('mantiene deshabilitada la asistencia ya guardada aunque falle la recarga', async () => {
    obtenerSesion.mockResolvedValueOnce(sesion).mockRejectedValue(new Error('Sin red'));
    declararAsistencia.mockResolvedValue(asistencia('tutor'));
    abrir();
    fireEvent.click(await screen.findByRole('button', { name: 'Declarar mi asistencia' }));
    expect((await screen.findByRole('alert')).textContent).toContain('Tu asistencia se guardó');
    expect(screen.getByRole('button', { name: 'Asistencia ya declarada' }).disabled).toBe(true);
  });
  it('muestra ambas declaraciones persistidas', async () => {
    obtenerSesion.mockResolvedValue({ ...sesion, asistencias: [asistencia('tutor'), asistencia('escolar')] });
    abrir();
    expect(await screen.findAllByText('Declarada')).toHaveLength(2);
    expect(screen.getByRole('button', { name: 'Asistencia ya declarada' }).disabled).toBe(true);
  });
  it('muestra el rechazo de una declaración duplicada del servidor', async () => {
    declararAsistencia.mockRejectedValue(new Error('Asistencia ya declarada en otra pestaña.'));
    abrir();
    fireEvent.click(await screen.findByRole('button', { name: 'Declarar mi asistencia' }));
    expect((await screen.findByRole('alert')).textContent).toContain('otra pestaña');
    expect(screen.queryByText('Tu asistencia fue declarada correctamente.')).toBeNull();
  });
  it('muestra error de consulta privada y enlace de retorno', async () => {
    obtenerSesion.mockRejectedValue(new Error('No participas en esta sesión.'));
    abrir();
    expect((await screen.findByRole('alert')).textContent).toContain('No participas');
    expect(screen.getByRole('link', { name: 'Volver al inicio' }).getAttribute('href')).toBe('/');
  });
  it('no etiqueta otras sesiones como demostración', async () => {
    obtenerSesion.mockResolvedValue({ ...sesion, id: 'otra' });
    abrir('/sesiones/otra');
    await screen.findByRole('heading', { name: 'Detalle de tutoría' });
    expect(screen.queryByText(/Sesión sintética/)).toBeNull();
  });
  it.each(['apoderado', 'coordinador'])('bloquea la ruta al %s', async rol => {
    estado.usuario = { id: 'otro', rol };
    abrir();
    await waitFor(() => expect(window.location.pathname).toBe('/no-autorizado'));
    expect(obtenerSesion).not.toHaveBeenCalled();
  });
  it('redirige al ingreso sin autenticación', async () => {
    estado.usuario = null; abrir();
    await waitFor(() => expect(window.location.pathname).toBe('/ingreso'));
    expect(obtenerSesion).not.toHaveBeenCalled();
  });
});
describe('Inicio y modo demo', () => {
  it.each(['tutor', 'escolar'])('permite navegar al %s en modo demo', async rol => {
    vi.stubEnv('VITE_I05_DEMO', 'true');
    estado.usuario = { id: rol + '-001', rol, nombre_uso: rol };
    abrir('/');
    fireEvent.click(screen.getByRole('link', { name: 'Ver sesión de demostración' }));
    await screen.findByRole('heading', { name: 'Detalle de tutoría' });
    expect(window.location.pathname).toBe('/sesiones/sesion-i05-demo');
  });
  it.each(['false', '', undefined])('oculta demo cuando la opción vale %s', valor => {
    vi.stubEnv('VITE_I05_DEMO', valor);
    render(<MemoryRouter><Inicio /></MemoryRouter>);
    expect(screen.queryByRole('link', { name: 'Ver sesión de demostración' })).toBeNull();
  });
  it.each(['apoderado', 'coordinador'])('no ofrece demo al %s', rol => {
    vi.stubEnv('VITE_I05_DEMO', 'true'); estado.usuario = { id: 'otro', rol };
    render(<MemoryRouter><Inicio /></MemoryRouter>);
    expect(screen.queryByRole('link', { name: 'Ver sesión de demostración' })).toBeNull();
  });
  it.each([['tutor', ['Mi expediente de tutor', 'Gestionar disponibilidad semanal']], ['escolar', ['Reservar una tutoría']], ['coordinador', ['Zona de coordinación']]])('conserva navegación del %s', (rol, enlaces) => {
    estado.usuario = { id: 'otro', rol }; render(<MemoryRouter><Inicio /></MemoryRouter>);
    enlaces.forEach(name => expect(screen.getByRole('link', { name })).toBeTruthy());
  });
});

describe('Ausencias I11', () => {
  it.each(['tutor', 'escolar'])('%s reporta su ausencia y bloquea declaraciones contradictorias', async rol => {
    estado.usuario = { id: rol + '-001', rol };
    const ausencia = { usuario_id: estado.usuario.id, rol, reportada_en: sesion.fin };
    obtenerSesion.mockResolvedValueOnce(sesion).mockResolvedValue({ ...sesion, ausencias: [ausencia] });
    let completar;
    reportarAusencia.mockImplementation(() => new Promise(resolve => { completar = resolve; }));
    abrir();
    const boton = await screen.findByRole('button', { name: 'Reportar mi ausencia' });
    expect(reportarAusencia).not.toHaveBeenCalled();
    expect(declararAsistencia).not.toHaveBeenCalled();
    fireEvent.click(boton);
    fireEvent.click(boton);
    expect(boton.disabled).toBe(true);
    expect(reportarAusencia).toHaveBeenCalledTimes(1);
    expect(reportarAusencia).toHaveBeenCalledWith(sesion.id);
    completar(ausencia);
    expect(await screen.findByText('Tu ausencia fue reportada correctamente.')).toBeTruthy();
    expect((await screen.findByRole('button', { name: 'Ausencia ya reportada' })).disabled).toBe(true);
    expect(screen.getByRole('button', { name: 'Declarar mi asistencia' }).disabled).toBe(true);
    expect(screen.getByText('Ausencia reportada')).toBeTruthy();
  });
  it('conserva la ausencia confirmada aunque falle la recarga', async () => {
    obtenerSesion.mockResolvedValueOnce(sesion).mockRejectedValue(new Error('Sin red'));
    reportarAusencia.mockResolvedValue({ usuario_id: 'tutor-001', rol: 'tutor', reportada_en: sesion.fin });
    abrir();
    fireEvent.click(await screen.findByRole('button', { name: 'Reportar mi ausencia' }));
    expect((await screen.findByRole('alert')).textContent).toContain('Tu ausencia se guardó');
    expect(screen.getByRole('button', { name: 'Ausencia ya reportada' }).disabled).toBe(true);
    expect(screen.getByRole('button', { name: 'Declarar mi asistencia' }).disabled).toBe(true);
  });
  it('muestra conflictos del servidor y no anuncia éxito', async () => {
    reportarAusencia.mockRejectedValue(new Error('Ausencia ya reportada.'));
    abrir();
    fireEvent.click(await screen.findByRole('button', { name: 'Reportar mi ausencia' }));
    expect((await screen.findByRole('alert')).textContent).toContain('Ausencia ya reportada');
    expect(screen.queryByText('Tu ausencia fue reportada correctamente.')).toBeNull();
  });
  it('impide reportar ausencia cuando la asistencia ya está guardada', async () => {
    obtenerSesion.mockResolvedValue({ ...sesion, asistencias: [asistencia('tutor')] });
    abrir();
    const boton = await screen.findByRole('button', { name: 'Reportar mi ausencia' });
    expect(boton.disabled).toBe(true);
    fireEvent.click(boton);
    expect(reportarAusencia).not.toHaveBeenCalled();
  });
});

describe('Resultado I11', () => {
  const resultado = { usuario_id: 'tutor-001', resumen: 'Se resolvieron ecuaciones.', registrado_en: sesion.fin };
  it('el tutor registra una vez y conserva el resultado al volver a entrar', async () => {
    let completar;
    registrarResultado.mockImplementation(() => new Promise(resolve => { completar = resolve; }));
    const vista = abrir();
    const campo = await screen.findByLabelText('Resumen del resultado');
    expect(registrarResultado).not.toHaveBeenCalled();
    expect(screen.getByRole('button', { name: 'Registrar resultado' }).disabled).toBe(true);
    fireEvent.change(campo, { target: { value: '  Se resolvieron ecuaciones.  ' } });
    const boton = screen.getByRole('button', { name: 'Registrar resultado' });
    fireEvent.click(boton); fireEvent.click(boton);
    expect(screen.getByRole('button', { name: 'Guardando resultado...' }).disabled).toBe(true);
    expect(screen.getByRole('button', { name: 'Reportar mi ausencia' }).disabled).toBe(true);
    expect(registrarResultado).toHaveBeenCalledTimes(1);
    expect(registrarResultado).toHaveBeenCalledWith(sesion.id, resultado.resumen);
    completar(resultado);
    expect(await screen.findByText('El resultado de la tutoría fue registrado correctamente.')).toBeTruthy();
    expect(screen.getByText(resultado.resumen)).toBeTruthy();
    expect(screen.queryByLabelText('Resumen del resultado')).toBeNull();
    expect(declararAsistencia).not.toHaveBeenCalled();
    expect(reportarAusencia).not.toHaveBeenCalled();
    vista.unmount();
    obtenerSesion.mockResolvedValue({ ...sesion, resultado });
    abrir();
    expect(await screen.findByText(resultado.resumen)).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Registrar resultado' })).toBeNull();
  });
  it('el escolar consulta el resultado sin disponer de formulario', async () => {
    estado.usuario = { id: 'escolar-001', rol: 'escolar' };
    const vista = abrir();
    expect(await screen.findByText('El tutor aún no ha registrado el resultado.')).toBeTruthy();
    expect(screen.queryByLabelText('Resumen del resultado')).toBeNull();
    vista.unmount();
    obtenerSesion.mockResolvedValue({ ...sesion, resultado });
    abrir();
    expect(await screen.findByText(resultado.resumen)).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Registrar resultado' })).toBeNull();
    expect(registrarResultado).not.toHaveBeenCalled();
  });
  it('conserva el texto ante error o conflicto para no perder el trabajo', async () => {
    registrarResultado.mockRejectedValue(new Error('Resultado ya registrado en otra pestaña.'));
    abrir();
    const campo = await screen.findByLabelText('Resumen del resultado');
    fireEvent.change(campo, { target: { value: resultado.resumen } });
    fireEvent.click(screen.getByRole('button', { name: 'Registrar resultado' }));
    expect((await screen.findByRole('alert')).textContent).toContain('otra pestaña');
    expect(campo.value).toBe(resultado.resumen);
    expect(screen.queryByText('El resultado de la tutoría fue registrado correctamente.')).toBeNull();
  });
  it('rechaza texto vacío y más de 2.000 caracteres', async () => {
    abrir();
    const campo = await screen.findByLabelText('Resumen del resultado');
    const boton = screen.getByRole('button', { name: 'Registrar resultado' });
    for (const value of ['   ', 'ñ'.repeat(2001)]) {
      fireEvent.change(campo, { target: { value } });
      expect(boton.disabled).toBe(true);
    }
    expect(registrarResultado).not.toHaveBeenCalled();
  });
});
