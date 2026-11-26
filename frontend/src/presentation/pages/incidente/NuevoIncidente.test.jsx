// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

import NuevoIncidente from './NuevoIncidente.jsx';
import { crearIncidente, listarTutoriasIncidentes } from '../../../adapters/http/incidenteApi.js';

vi.mock('../../hooks/useSesion.js', () => ({
  useSesion: () => ({
    usuario: { id: 'escolar-1', rol: 'escolar', nombre_uso: 'Escolar Uno' },
  }),
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
  asistio: true,
};

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/incidentes/nuevo']}>
      <Routes>
        <Route path="/incidentes/nuevo" element={<NuevoIncidente />} />
        <Route path="/incidentes" element={<h1>Incidentes</h1>} />
      </Routes>
    </MemoryRouter>
  );
}

function completarFormulario(descripcion) {
  fireEvent.change(screen.getByLabelText('Tutoría atendida'), {
    target: { value: 'reserva-1' },
  });
  fireEvent.change(screen.getByLabelText('Descripción de lo ocurrido'), {
    target: { value: descripcion },
  });
}

beforeEach(() => {
  vi.resetAllMocks();
  listarTutoriasIncidentes.mockResolvedValue({ tutorias: [tutoria] });
  crearIncidente.mockResolvedValue({ id: 'incidente-nuevo' });
});

afterEach(cleanup);

describe('F-24: validación de evidencia en incidentes', () => {
  it('rechaza un archivo con extensión PDF pero MIME que no corresponde', async () => {
    renderizar();
    await screen.findByRole('heading', { name: 'Registrar incidente' });
    completarFormulario('El tutor hizo un comentario ofensivo durante la tutoría.');
    fireEvent.change(screen.getByLabelText(/Evidencias obligatorias/), {
      target: { files: [new File(['contenido'], 'evidencia.pdf', { type: 'text/plain' })] },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Registrar incidente' }));

    expect(await screen.findByText(/PDF, JPG o PNG/)).toBeTruthy();
    expect(crearIncidente).not.toHaveBeenCalled();
  });
});

describe('B-03: accesibilidad del error de descripción', () => {
  it('asocia la descripción inválida con su mensaje mediante aria-describedby', async () => {
    renderizar();
    await screen.findByRole('heading', { name: 'Registrar incidente' });
    completarFormulario('corta');
    fireEvent.change(screen.getByLabelText(/Evidencias obligatorias/), {
      target: { files: [new File(['%PDF-1.4'], 'chat.pdf', { type: 'application/pdf' })] },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Registrar incidente' }));

    const descripcion = screen.getByLabelText('Descripción de lo ocurrido');
    const mensaje = await screen.findByText('Describe lo ocurrido entre 20 y 2000 caracteres.');
    const idMensaje = mensaje.getAttribute('id');
    expect(idMensaje).toBeTruthy();
    expect(descripcion.getAttribute('aria-describedby')).toBe(idMensaje);
  });
});
