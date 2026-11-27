// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

vi.mock('../hooks/useSesion.js', () => ({ useSesion: vi.fn() }));
vi.mock('../../application/acceso/sesion.js', () => ({ salir: vi.fn() }));

import { useSesion } from '../hooks/useSesion.js';
import { salir } from '../../application/acceso/sesion.js';
import Icono from './Icono.jsx';
import CalendarioSemanal from './CalendarioSemanal.jsx';
import RutaProtegida from './RutaProtegida.jsx';
import RutaProtegidaConRetorno from './RutaProtegidaConRetorno.jsx';
import DisenoApp from './DisenoApp.jsx';

beforeEach(() => {
  vi.clearAllMocks();
  useSesion.mockReturnValue({ usuario: null, cargando: false });
});

afterEach(cleanup);

describe('Icono', () => {
  it('renderiza un trazo decorativo', () => {
    const { container } = render(<Icono nombre="inicio" className="size-8" />);
    const svg = container.querySelector('svg');
    expect(svg.getAttribute('aria-hidden')).toBe('true');
    expect(svg.className.baseVal).toContain('size-8');
    expect(svg.querySelector('path').getAttribute('d')).toContain('m2.25 12');
  });
});

describe('CalendarioSemanal', () => {
  it('agrupa por día, ordena y permite eliminar', () => {
    const alEliminar = vi.fn();
    render(
      <CalendarioSemanal
        slots={[
          { id: 's2', dia_semana: 1, inicio: '10:00', fin: '11:00' },
          { id: 's1', dia_semana: 1, inicio: '09:00', fin: '10:00' },
        ]}
        alEliminar={alEliminar}
        eliminandoId="s2"
      />
    );
    expect(screen.getAllByText('Lunes').length).toBe(1);
    expect(screen.getByText('09:00–10:00')).toBeTruthy();
    expect(screen.getAllByText('Eliminando…').length).toBe(1);
    fireEvent.click(screen.getByRole('button', { name: /Eliminar horario 09:00 del Lunes/i }));
    expect(alEliminar).toHaveBeenCalledWith(expect.objectContaining({ id: 's1' }));
  });
});

describe('RutaProtegida', () => {
  function renderRuta(roles = ['escolar'], hijo = <p>contenido</p>) {
    return render(
      <MemoryRouter initialEntries={['/protegida']}>
        <Routes>
          <Route path="/protegida" element={<RutaProtegida roles={roles}>{hijo}</RutaProtegida>} />
          <Route path="/ingreso" element={<p>ingreso</p>} />
          <Route path="/no-autorizado" element={<p>no autorizado</p>} />
        </Routes>
      </MemoryRouter>
    );
  }

  it('muestra carga mientras la sesión está pendiente', () => {
    useSesion.mockReturnValue({ usuario: null, cargando: true });
    renderRuta();
    expect(screen.getByRole('status').textContent).toContain('Cargando');
  });

  it('redirige al ingreso sin usuario', () => {
    renderRuta();
    expect(screen.getByText('ingreso')).toBeTruthy();
  });

  it('redirige cuando el rol no está permitido', () => {
    useSesion.mockReturnValue({ usuario: { rol: 'escolar' }, cargando: false });
    renderRuta(['coordinador']);
    expect(screen.getByText('no autorizado')).toBeTruthy();
  });

  it('renderiza el contenido cuando el rol está permitido', () => {
    useSesion.mockReturnValue({ usuario: { rol: 'escolar' }, cargando: false });
    renderRuta();
    expect(screen.getByText('contenido')).toBeTruthy();
  });

  it('conserva el destino en RutaProtegidaConRetorno', () => {
    useSesion.mockReturnValue({ usuario: null, cargando: false });
    render(
      <MemoryRouter initialEntries={['/privado?token=abc']}>
        <Routes>
          <Route
            path="/privado"
            element={<RutaProtegidaConRetorno roles={['escolar']}><p>privado</p></RutaProtegidaConRetorno>}
          />
          <Route path="/ingreso" element={<p>ingreso</p>} />
        </Routes>
      </MemoryRouter>
    );
    expect(screen.getByText('ingreso')).toBeTruthy();
  });
});

describe('DisenoApp', () => {
  it('muestra el ingreso sin sesión', () => {
    render(
      <MemoryRouter>
        <DisenoApp />
      </MemoryRouter>
    );
    expect(screen.getByRole('link', { name: 'Ingresar' })).toBeTruthy();
  });

  it('muestra el menú del usuario y permite cerrar sesión', async () => {
    useSesion.mockReturnValue({
      usuario: { nombre_uso: 'Ana Pérez', rol: 'escolar' },
      cargando: false,
    });
    render(
      <MemoryRouter>
        <DisenoApp />
      </MemoryRouter>
    );
    expect(screen.getByText('Ana Pérez')).toBeTruthy();
    expect(screen.getByText('Escolar')).toBeTruthy();
    expect(screen.getByRole('link', { name: /Reservar tutoría/i })).toBeTruthy();
    fireEvent.click(screen.getAllByRole('button', { name: /Cerrar sesión/i })[0]);
    expect(salir).toHaveBeenCalled();
  });
});
