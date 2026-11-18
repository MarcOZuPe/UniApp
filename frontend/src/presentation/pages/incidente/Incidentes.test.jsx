// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import Incidentes from './Incidentes.jsx';
import {
  crearIncidente,
  listarIncidentes,
  listarTutoriasIncidentes,
  revisarIncidente,
} from '../../../adapters/http/incidenteApi.js';

const estado = vi.hoisted(() => ({
  sesion: { usuario: { id: 'escolar-1', rol: 'escolar', nombre_uso: 'Escolar Uno' } },
}));

vi.mock('../../hooks/useSesion.js', () => ({
  useSesion: () => estado.sesion,
}));

vi.mock('../../../adapters/http/incidenteApi.js', () => ({
  crearIncidente: vi.fn(),
  listarIncidentes: vi.fn(),
  listarTutoriasIncidentes: vi.fn(),
  revisarIncidente: vi.fn(),
  urlEvidencia: (incidenteID, evidenciaID) =>
    `/api/incidentes/${incidenteID}/evidencias/${evidenciaID}`,
}));

const tutoria = {
  reserva_id: 'reserva-1',
  escolar_id: 'escolar-1',
  escolar_nombre: 'Escolar Uno',
  tutor_id: 'tutor-1',
  tutor_nombre: 'Tutor Uno',
  curso: 'Álgebra',
  inicio_lima: '2026-11-17 09:00',
  fin_lima: '2026-11-17 10:00',
  asistio: true,
};

const incidentePendiente = {
  id: 'incidente-1',
  escolar_id: 'escolar-1',
  escolar_nombre: 'Escolar Uno',
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

function renderizar() {
  return render(
    <MemoryRouter>
      <Incidentes />
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
  estado.sesion = { usuario: { id: 'escolar-1', rol: 'escolar', nombre_uso: 'Escolar Uno' } };
  listarIncidentes.mockResolvedValue({ incidentes: [] });
  listarTutoriasIncidentes.mockResolvedValue({ tutorias: [tutoria] });
  crearIncidente.mockResolvedValue({ id: 'incidente-nuevo' });
  revisarIncidente.mockResolvedValue({ id: 'incidente-1', estado: 'revisado' });
});

afterEach(cleanup);

function completarFormulario() {
  fireEvent.change(screen.getByLabelText('Tutoría atendida'), {
    target: { value: 'reserva-1' },
  });
  fireEvent.change(screen.getByLabelText('Descripción de lo ocurrido'), {
    target: { value: 'El tutor hizo un comentario ofensivo durante la tutoría.' },
  });
}

describe('I13 - Incidentes confidenciales', () => {
  it('registra un reporte con evidencia PDF y muestra que el tutor no accede', async () => {
    renderizar();
    expect(await screen.findByRole('heading', { name: 'Nuevo reporte' })).toBeTruthy();
    expect(screen.getByText(/El tutor denunciado no tiene acceso/)).toBeTruthy();

    completarFormulario();
    const archivo = new File(['%PDF-1.4'], 'chat.pdf', { type: 'application/pdf' });
    fireEvent.change(screen.getByLabelText(/Evidencias obligatorias/), {
      target: { files: [archivo] },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Registrar incidente' }));

    await waitFor(() => expect(crearIncidente).toHaveBeenCalledTimes(1));
    const formulario = crearIncidente.mock.calls[0][0];
    expect(formulario.get('reserva_id')).toBe('reserva-1');
    expect(formulario.getAll('evidencias')).toHaveLength(1);
    expect(await screen.findByText(/Reporte registrado/)).toBeTruthy();
  });

  it('exige al menos una evidencia antes de enviar', async () => {
    renderizar();
    await screen.findByRole('heading', { name: 'Nuevo reporte' });
    completarFormulario();
    fireEvent.click(screen.getByRole('button', { name: 'Registrar incidente' }));

    expect(await screen.findByText('Adjunta al menos una evidencia del incidente.')).toBeTruthy();
    expect(crearIncidente).not.toHaveBeenCalled();
  });

  it('rechaza en el navegador archivos que no son PDF, JPG o PNG', async () => {
    renderizar();
    await screen.findByRole('heading', { name: 'Nuevo reporte' });
    completarFormulario();
    const archivo = new File(['texto'], 'notas.txt', { type: 'text/plain' });
    fireEvent.change(screen.getByLabelText(/Evidencias obligatorias/), {
      target: { files: [archivo] },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Registrar incidente' }));

    expect(await screen.findByText('La evidencia debe ser PDF, JPG o PNG.')).toBeTruthy();
    expect(crearIncidente).not.toHaveBeenCalled();
  });

  it('permite al apoderado elegir el escolar a través de la tutoría', async () => {
    estado.sesion = { usuario: { id: 'apoderado-1', rol: 'apoderado', nombre_uso: 'Apoderado' } };
    renderizar();
    await screen.findByRole('heading', { name: 'Nuevo reporte' });
    expect(screen.getByText(/Escolar Uno/)).toBeTruthy();
    completarFormulario();
    const archivo = new File(['%PDF-1.4'], 'chat.pdf', { type: 'application/pdf' });
    fireEvent.change(screen.getByLabelText(/Evidencias obligatorias/), {
      target: { files: [archivo] },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Registrar incidente' }));

    await waitFor(() => expect(crearIncidente).toHaveBeenCalledTimes(1));
    expect(crearIncidente.mock.calls[0][0].get('escolar_id')).toBe('escolar-1');
  });

  it('muestra a coordinación los reportes y registra la revisión', async () => {
    estado.sesion = { usuario: { id: 'coord-1', rol: 'coordinador', nombre_uso: 'Coordinación' } };
    listarIncidentes.mockResolvedValue({ incidentes: [incidentePendiente] });
    renderizar();

    expect(await screen.findByRole('heading', { name: 'Reportes recibidos' })).toBeTruthy();
    expect(screen.queryByRole('heading', { name: 'Nuevo reporte' })).toBeNull();
    expect(screen.getByText(/Revisa los incidentes reportados por escolares y apoderados/)).toBeTruthy();
    expect(screen.queryByText(/Reporta una posible falta/)).toBeNull();
    expect(listarTutoriasIncidentes).not.toHaveBeenCalled();

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
  });
});
