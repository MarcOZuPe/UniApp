// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';

import NuevoIncidente from './NuevoIncidente.jsx';
import {
  crearIncidente,
  listarTutoriasIncidentes,
} from '../../../adapters/http/incidenteApi.js';

const estado = vi.hoisted(() => ({
  sesion: { usuario: { id: 'escolar-1', rol: 'escolar', nombre_uso: 'Escolar Uno' } },
}));

vi.mock('../../hooks/useSesion.js', () => ({
  useSesion: () => estado.sesion,
}));

vi.mock('../../../adapters/http/incidenteApi.js', () => ({
  crearIncidente: vi.fn(),
  listarTutoriasIncidentes: vi.fn(),
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

function ListaDestino() {
  const location = useLocation();
  return (
    <>
      <h1>Incidentes</h1>
      <p>Lista de incidentes</p>
      {location.state?.mensaje && <p role="status">{location.state.mensaje}</p>}
    </>
  );
}

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/incidentes/nuevo']}>
      <Routes>
        <Route path="/incidentes/nuevo" element={<NuevoIncidente />} />
        <Route path="/incidentes" element={<ListaDestino />} />
      </Routes>
    </MemoryRouter>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
  estado.sesion = { usuario: { id: 'escolar-1', rol: 'escolar', nombre_uso: 'Escolar Uno' } };
  listarTutoriasIncidentes.mockResolvedValue({ tutorias: [tutoria] });
  crearIncidente.mockResolvedValue({ id: 'incidente-nuevo' });
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

describe('I13 - Registro de incidentes', () => {
  it('registra un reporte con evidencia PDF y vuelve al listado', async () => {
    renderizar();
    expect(await screen.findByRole('heading', { name: 'Registrar incidente' })).toBeTruthy();
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
    expect(await screen.findByText('Lista de incidentes')).toBeTruthy();
    expect(screen.getByRole('status').textContent).toMatch(/Reporte registrado/);
  });

  it('exige al menos una evidencia antes de enviar', async () => {
    renderizar();
    await screen.findByRole('heading', { name: 'Registrar incidente' });
    completarFormulario();
    fireEvent.click(screen.getByRole('button', { name: 'Registrar incidente' }));

    expect(await screen.findByText('Adjunta al menos una evidencia del incidente.')).toBeTruthy();
    expect(crearIncidente).not.toHaveBeenCalled();
  });

  it('rechaza en el navegador archivos que no son PDF, JPG o PNG', async () => {
    renderizar();
    await screen.findByRole('heading', { name: 'Registrar incidente' });
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
    await screen.findByRole('heading', { name: 'Registrar incidente' });
    expect(screen.getByText(/Escolar Uno/)).toBeTruthy();
    completarFormulario();
    const archivo = new File(['%PDF-1.4'], 'chat.pdf', { type: 'application/pdf' });
    fireEvent.change(screen.getByLabelText(/Evidencias obligatorias/), {
      target: { files: [archivo] },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Registrar incidente' }));

    await waitFor(() => expect(crearIncidente).toHaveBeenCalledTimes(1));
    expect(crearIncidente.mock.calls[0][0].get('escolar_id')).toBe('escolar-1');
    expect(
      await screen.findByText(
        'Reporte registrado. La evidencia queda disponible solo para ti, el escolar y coordinación.'
      )
    ).toBeTruthy();
  });
});
