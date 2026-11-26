// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import Coordinacion from './Coordinacion.jsx';
import {
  listarAuditoriaCoordinacion,
  obtenerMetricasCoordinacion,
  obtenerPanelCoordinacion,
} from '../../../application/acceso/coordinacion.js';

const estado = vi.hoisted(() => ({
  usuario: { id: 'coord-1', rol: 'coordinador', nombre_uso: 'Coordinación Prueba' },
}));

vi.mock('../../hooks/useSesion.js', () => ({
  useSesion: () => ({ usuario: estado.usuario }),
}));

vi.mock('../../../application/acceso/coordinacion.js', () => ({
  obtenerPanelCoordinacion: vi.fn(),
  obtenerMetricasCoordinacion: vi.fn(),
  listarAuditoriaCoordinacion: vi.fn(),
}));

const metricas = {
  reservas: { totales: 12, futuras: 3, finalizadas: 9 },
  asistencia: {
    sesiones_finalizadas: 9,
    con_asistencia_escolar: 8,
    con_asistencia_tutor: 7,
    tasa_escolar_porcentaje: 88.89,
    tasa_tutor_porcentaje: 77.78,
  },
  horas: { internas: 9 },
  gestion: {
    expedientes_pendientes: 2,
    incidentes_pendientes: 1,
    cursos_pendientes: 4,
    solicitudes_colegio_pendientes: 3,
  },
  actualizado_en: '2026-11-18T18:00:00Z',
};

const eventos = [
  {
    id: 2,
    ocurrido_en: '2026-11-18T18:05:00Z',
    actor_id: 'coord-1',
    actor_nombre: 'Coordinación Prueba',
    actor_rol: 'coordinador',
    accion: 'incidente.revisado',
    entidad: 'incidente',
    entidad_id: 'incidente-1',
    detalle: { resultado: 'procede', estado: 'revisado' },
  },
  {
    id: 1,
    ocurrido_en: '2026-11-18T18:00:00Z',
    actor_id: 'tutor-1',
    actor_nombre: 'Tutor Prueba',
    actor_rol: 'tutor',
    accion: 'cupo.actualizado',
    entidad: 'cupo_semanal',
    entidad_id: 'tutor-1',
    detalle: { minutos_anteriores: 180, minutos: 240 },
  },
];

function renderizar() {
  return render(
    <MemoryRouter>
      <Coordinacion />
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
  estado.usuario = { id: 'coord-1', rol: 'coordinador', nombre_uso: 'Coordinación Prueba' };
  obtenerPanelCoordinacion.mockResolvedValue({ mensaje: 'Panel de coordinación' });
  obtenerMetricasCoordinacion.mockResolvedValue(metricas);
  listarAuditoriaCoordinacion.mockResolvedValue({ eventos });
});

afterEach(cleanup);

describe('I14 - panel de coordinación', () => {
  it('muestra métricas agregadas y la auditoría sin datos sensibles', async () => {
    renderizar();

    expect(await screen.findByRole('heading', { name: 'Métricas agregadas' })).toBeTruthy();
    expect(screen.getByText('12')).toBeTruthy();
    expect(screen.getByText('9 h')).toBeTruthy();
    expect(screen.getByText('88.89%', { exact: false })).toBeTruthy();
    expect(listarAuditoriaCoordinacion).toHaveBeenCalledWith(50);

    expect(await screen.findByRole('heading', { name: 'Auditoría de decisiones y reglas' })).toBeTruthy();
    expect(screen.getByText('Incidente revisado')).toBeTruthy();
    expect(screen.getByText('Cupo actualizado')).toBeTruthy();
    expect(screen.getByText('180 → 240 minutos.')).toBeTruthy();
    expect(screen.queryByText(/clave_almacenamiento|huella/i)).toBeNull();
    expect(screen.queryByRole('link', { name: 'Gestionar coordinadores' })).toBeNull();
  });

  it('solo ofrece gestión de coordinadores al rol general', async () => {
    estado.usuario = {
      id: 'general-1',
      rol: 'coordinador_general',
      nombre_uso: 'Coordinación General',
    };
    renderizar();

    expect(
      await screen.findByRole('link', { name: 'Gestionar coordinadores' })
    ).toBeTruthy();
  });
});
