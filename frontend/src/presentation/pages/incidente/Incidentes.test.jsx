// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import Incidentes from './Incidentes.jsx';
import {
  listarIncidentes,
  revisarIncidente,
} from '../../../adapters/http/incidenteApi.js';

const estado = vi.hoisted(() => ({
  sesion: { usuario: { id: 'escolar-1', rol: 'escolar', nombre_uso: 'Escolar Uno' } },
}));

vi.mock('../../hooks/useSesion.js', () => ({
  useSesion: () => estado.sesion,
}));

vi.mock('../../../adapters/http/incidenteApi.js', () => ({
  listarIncidentes: vi.fn(),
  revisarIncidente: vi.fn(),
  urlEvidencia: (incidenteID, evidenciaID) =>
    `/api/incidentes/${incidenteID}/evidencias/${evidenciaID}`,
}));

const incidentePendiente = {
  id: 'incidente-1',
  escolar_id: 'escolar-1',
  escolar_nombre: 'Escolar Uno',
  reportado_por: 'escolar-1',
  reportado_por_rol: 'escolar',
  tutor_nombre: 'Tutor Uno',
  curso: 'Álgebra',
  inicio_lima: '2026-11-17 09:00',
  tipo: 'trato_inadecuado',
  tipo_etiqueta: 'Trato inadecuado',
  descripcion: 'El tutor hizo un comentario ofensivo durante la tutoría.',
  estado: 'pendiente',
  evidencias: [
    { id: 'evidencia-1', nombre: 'chat.pdf', tamano_bytes: 2048, descargar_en: '/api/incidentes/incidente-1/evidencias/evidencia-1' },
  ],
};

function renderizar(entradas = ['/incidentes']) {
  return render(
    <MemoryRouter initialEntries={entradas}>
      <Incidentes />
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
  estado.sesion = { usuario: { id: 'escolar-1', rol: 'escolar', nombre_uso: 'Escolar Uno' } };
  listarIncidentes.mockResolvedValue({ incidentes: [incidentePendiente] });
  revisarIncidente.mockResolvedValue({ id: 'incidente-1', estado: 'revisado' });
});

afterEach(cleanup);

describe('I13 - Listado de incidentes', () => {
  it('muestra los reportes y enlaza al formulario en una ruta separada', async () => {
    renderizar();

    expect(await screen.findByRole('heading', { name: 'Mis reportes' })).toBeTruthy();
    expect(screen.queryByRole('heading', { name: 'Nuevo reporte' })).toBeNull();
    expect(screen.getByRole('link', { name: 'Registrar un incidente' }).getAttribute('href')).toBe(
      '/incidentes/nuevo'
    );

    expect(screen.getByText(/Reportado por: Escolar/)).toBeTruthy();

    const disparador = screen.getByRole('button', { name: /Trato inadecuado/ });
    expect(disparador.getAttribute('aria-expanded')).toBe('false');
    expect(screen.queryByText('El tutor hizo un comentario ofensivo durante la tutoría.')).toBeNull();
    expect(screen.queryByRole('link', { name: 'chat.pdf' })).toBeNull();

    fireEvent.click(disparador);
    expect(disparador.getAttribute('aria-expanded')).toBe('true');
    expect(screen.getByText('El tutor hizo un comentario ofensivo durante la tutoría.')).toBeTruthy();
    expect(screen.getByRole('link', { name: 'chat.pdf' }).getAttribute('href')).toBe(
      '/api/incidentes/incidente-1/evidencias/evidencia-1'
    );

    fireEvent.click(disparador);
    expect(screen.queryByText('El tutor hizo un comentario ofensivo durante la tutoría.')).toBeNull();
  });

  it('distingue un reporte registrado por el apoderado', async () => {
    listarIncidentes.mockResolvedValue({
      incidentes: [
        {
          ...incidentePendiente,
          id: 'incidente-2',
          reportado_por: 'apoderado-1',
          reportado_por_rol: 'apoderado',
        },
      ],
    });
    renderizar();

    expect(await screen.findByText(/Reportado por: Apoderado/)).toBeTruthy();
  });

  it('muestra la confirmación recibida luego de registrar un incidente', async () => {
    renderizar([
      {
        pathname: '/incidentes',
        state: { mensaje: 'Reporte registrado. La evidencia queda disponible solo para ti, tu apoderado y coordinación.' },
      },
    ]);

    expect(await screen.findByText(/Reporte registrado/)).toBeTruthy();
  });

  it('muestra a coordinación los reportes y registra la revisión', async () => {
    estado.sesion = { usuario: { id: 'coord-1', rol: 'coordinador', nombre_uso: 'Coordinación' } };
    renderizar();

    expect(await screen.findByRole('heading', { name: 'Reportes recibidos' })).toBeTruthy();
    expect(screen.getByText(/Revisa los incidentes reportados por escolares y apoderados/)).toBeTruthy();
    expect(screen.queryByText(/Reporta una posible falta/)).toBeNull();
    expect(screen.queryByRole('link', { name: 'Registrar un incidente' })).toBeNull();
    expect(screen.queryByLabelText('Comentario de revisión')).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: /Trato inadecuado/ }));
    fireEvent.change(screen.getByLabelText('Comentario de revisión'), {
      target: { value: 'La evidencia sustenta el reporte.' },
    });
    fireEvent.click(screen.getByRole('button', { name: /Registrar revisión de Escolar Uno/ }));

    await waitFor(() =>
      expect(revisarIncidente).toHaveBeenCalledWith('incidente-1', {
        resultado: 'procede',
        comentario: 'La evidencia sustenta el reporte.',
      })
    );
    expect(await screen.findByText('Revisión registrada.')).toBeTruthy();
  });
});
