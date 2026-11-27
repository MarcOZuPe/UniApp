import { afterEach, describe, expect, it, vi } from 'vitest';

import { accesosDeInicio, enlacesDelMenu } from './navegacion.js';

afterEach(() => {
  vi.unstubAllEnvs();
});

describe('navegación por rol', () => {
  it('filtra el menú según el rol', () => {
    const escolar = enlacesDelMenu('escolar').map((enlace) => enlace.ruta);
    expect(escolar).toContain('/reservar');
    expect(escolar).toContain('/incidentes');
    expect(escolar).not.toContain('/coordinacion');

    const coordinador = enlacesDelMenu('coordinador').map((enlace) => enlace.ruta);
    expect(coordinador).toContain('/coordinacion');
    expect(coordinador).toContain('/coordinacion/minutos');
    expect(coordinador).not.toContain('/reservar');

    const administrador = enlacesDelMenu('administrador').map((enlace) => enlace.ruta);
    expect(administrador).toContain('/administracion');
  });

  it('incluye accesos públicos cuando no hay rol', () => {
    const rutas = enlacesDelMenu(undefined).map((enlace) => enlace.ruta);
    expect(rutas).toContain('/catalogo');
    expect(rutas).not.toContain('/reservar');
  });

  it('excluye el menú y los accesos de inicio según su configuración', () => {
    const inicio = accesosDeInicio('escolar').map((acceso) => acceso.ruta);
    expect(inicio).not.toContain('/');
    expect(inicio).toContain('/reservar');
    expect(inicio).toContain('/perfil');
    expect(inicio).not.toContain('/sesiones/sesion-i05-demo');
  });

  it('incluye la sesión demo solo si el modo demo está activo', () => {
    vi.stubEnv('VITE_I05_DEMO', 'true');
    const inicio = accesosDeInicio('escolar').map((acceso) => acceso.ruta);
    expect(inicio).toContain('/sesiones/sesion-i05-demo');
  });
});
