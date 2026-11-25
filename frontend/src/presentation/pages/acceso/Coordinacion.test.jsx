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

const eventos = [{
  id: 2,
  ocurrido_en: '2026-11-18T18:05:00Z',
  actor_nombre: 'Coordinación Prueba',
  actor_rol: 'coordinador',
  accion: 'incidente.revisado',
  entidad: 'incidente',
  entidad_id: 'incidente-1',
  detalle: { resultado: 'procede', estado: 'revisado' },
}];

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
  obtenerMetricasCoordinacion.mockResolvedValue({
    reservas: { totales: 1, futuras: 1, finalizadas: 0 },
    asistencia: {
      sesiones_finalizadas: 0,
      con_asistencia_escolar: 0,
      con_asistencia_tutor: 0,
      tasa_escolar_porcentaje: 0,
      tasa_tutor_porcentaje: 0,
    },
    horas: { internas: 0 },
    gestion: {
      expedientes_pendientes: 0,
      incidentes_pendientes: 0,
      cursos_pendientes: 0,
      solicitudes_colegio_pendientes: 0,
    },
    actualizado_en: '2026-11-18T18:00:00Z',
  });
  listarAuditoriaCoordinacion.mockResolvedValue({ eventos });
});

afterEach(cleanup);

describe('F-03: acceso visible a minutos', () => {
  it('enlaza a la bandeja de revisión de minutos', async () => {
    renderizar();

    expect(
      await screen.findByRole('link', { name: /revisión de minutos|revisar minutos/i })
    ).toBeTruthy();
  });
});

describe('F-20: fallo parcial en el panel de coordinación', () => {
  it('conserva el panel y la auditoría cuando solo fallan las métricas', async () => {
    obtenerMetricasCoordinacion.mockRejectedValue(new Error('Métricas no disponibles'));

    renderizar();

    expect(
      await screen.findByRole('heading', { name: 'Panel de coordinación' })
    ).toBeTruthy();
    expect(
      await screen.findByRole('heading', { name: 'Auditoría de decisiones y reglas' })
    ).toBeTruthy();
    expect(screen.getByText('Incidente revisado')).toBeTruthy();
    expect(screen.getByRole('alert').textContent).toBe('Métricas no disponibles');
  });
});
