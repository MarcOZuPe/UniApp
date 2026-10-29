// @vitest-environment jsdom
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import App from '../../../app/App.jsx';
import { obtenerNotificaciones } from '../../../adapters/http/notificacionesApi.js';
const sesion = vi.hoisted(() => ({ usuario: null, cargando: false }));
vi.mock('../../hooks/useSesion.js', () => ({ useSesion: () => sesion }));
vi.mock('../../../application/acceso/acciones.js', () => ({ cargarSesion: vi.fn(), salir: vi.fn() }));
vi.mock('../../../adapters/http/notificacionesApi.js', () => ({ obtenerNotificaciones: vi.fn() }));
const dato = { id: 1, asunto: 'Reserva confirmada', cuerpo: 'Tutoría reservada.', fecha: '2026-10-22T15:00:00Z' };
beforeEach(() => { vi.resetAllMocks(); sesion.usuario = { id: 'tutor', rol: 'tutor', nombre_uso: 'Tutor' }; obtenerNotificaciones.mockResolvedValue([dato]); });
afterEach(cleanup);
function abrir(ruta = '/notificaciones') { window.history.replaceState({}, '', ruta); render(<App />); }
it.each(['tutor', 'escolar'])('navega desde Inicio como %s', async rol => {
 sesion.usuario.rol = rol; abrir('/'); fireEvent.click(screen.getByRole('link', { name: 'Mis notificaciones' }));
 expect(screen.getByRole('status').textContent).toContain('Cargando');
 expect(await screen.findByText('Reserva confirmada')).toBeTruthy();
 expect(obtenerNotificaciones).toHaveBeenCalledOnce();
 expect(screen.getByText(/10:00/)).toBeTruthy();
});
it('muestra un recordatorio recibido', async () => { obtenerNotificaciones.mockResolvedValue([{ ...dato, asunto: 'Recordatorio de tutoría' }]); abrir(); expect(await screen.findByText('Recordatorio de tutoría')).toBeTruthy(); });
it('muestra lista vacía', async () => { obtenerNotificaciones.mockResolvedValue([]); abrir(); expect(await screen.findByText('Aún no tienes notificaciones.')).toBeTruthy(); });
it('muestra fallo de consulta', async () => { obtenerNotificaciones.mockRejectedValue(new Error('fallo')); abrir(); expect(await screen.findByRole('alert')).toBeTruthy(); });
it.each(['coordinador', 'apoderado'])('rechaza al %s y oculta el enlace', async rol => {
 sesion.usuario.rol = rol; abrir('/'); expect(screen.queryByRole('link', { name: 'Mis notificaciones' })).toBeNull(); cleanup();
 abrir(); await waitFor(() => expect(window.location.pathname).toBe('/no-autorizado')); expect(obtenerNotificaciones).not.toHaveBeenCalled();
});
it('requiere sesión', async () => { sesion.usuario = null; abrir(); await waitFor(() => expect(window.location.pathname).toBe('/ingreso')); expect(obtenerNotificaciones).not.toHaveBeenCalled(); });
