import { describe, expect, it } from 'vitest';

import { validarRegistro } from './validadores.js';

const formularioValido = {
  email: 'ana@example.test',
  nombre_uso: 'Ana',
  clave: 'Clave-Segura-2026!',
  confirmar: 'Clave-Segura-2026!',
  rol: 'escolar',
};

describe('validación de registro: roles no habilitados', () => {
  it('debe rechazar un rol que no admite registro público', () => {
    expect(validarRegistro({ ...formularioValido, rol: 'administrador' }).rol).toBeTruthy();
    expect(validarRegistro({ ...formularioValido, rol: 'coordinador' }).rol).toBeTruthy();
  });

  it('debe rechazar un rol inexistente', () => {
    expect(validarRegistro({ ...formularioValido, rol: 'rol-inexistente' }).rol).toBeTruthy();
  });
});
