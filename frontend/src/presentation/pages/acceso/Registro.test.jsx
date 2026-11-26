// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

import Registro from './Registro.jsx';
import { listarFacultades, registrar } from '../../../application/acceso/registro.js';

vi.mock('../../../application/acceso/registro.js', () => ({
  registrar: vi.fn(),
  listarFacultades: vi.fn(),
}));

vi.mock('../../../application/acceso/confirmacion.js', () => ({
  reenviarConfirmacion: vi.fn(),
}));

const CLAVE_VALIDA = 'Clave-Segura-2026!';

function completarFormulario() {
  fireEvent.change(screen.getByLabelText('Nombre de uso'), { target: { value: 'Escolar Prueba' } });
  fireEvent.change(screen.getByLabelText('Correo'), { target: { value: 'escolar@example.test' } });
  fireEvent.change(screen.getByLabelText('Contraseña'), { target: { value: CLAVE_VALIDA } });
  fireEvent.change(screen.getByLabelText('Repite la contraseña'), { target: { value: CLAVE_VALIDA } });
}

beforeEach(() => {
  vi.resetAllMocks();
  listarFacultades.mockResolvedValue({ facultades: [] });
  registrar.mockReturnValue(new Promise(() => {}));
});

afterEach(cleanup);

describe('F-25: doble envío en registro', () => {
  it('no debe enviar dos registros si se hace doble clic', () => {
    render(
      <MemoryRouter>
        <Registro />
      </MemoryRouter>
    );
    completarFormulario();

    const boton = screen.getByRole('button', { name: 'Crear cuenta' });
    fireEvent.click(boton);
    fireEvent.click(boton);

    expect(registrar).toHaveBeenCalledTimes(1);
    expect(screen.getByRole('button', { name: 'Crear cuenta' }).disabled).toBe(true);
  });
});
