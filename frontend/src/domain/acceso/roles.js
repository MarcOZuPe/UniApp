const POLITICA_POR_DEFECTO = 'basica';

export const ROLES = {
  escolar: {
    etiqueta: 'Escolar',
    permiteRegistro: true,
    politicaClave: 'basica',
    requiereCorreoInstitucional: false,
  },
  apoderado: {
    etiqueta: 'Apoderado',
    permiteRegistro: true,
    politicaClave: 'basica',
    requiereCorreoInstitucional: false,
  },
  tutor: {
    etiqueta: 'Tutor UNI',
    permiteRegistro: true,
    politicaClave: 'estricta',
    requiereCorreoInstitucional: true,
  },
  coordinador: {
    etiqueta: 'Coordinación',
    permiteRegistro: false,
    politicaClave: 'estricta',
    requiereCorreoInstitucional: false,
  },
};

export const ROLES_REGISTRO = Object.entries(ROLES)
  .filter(([, definicionRol]) => definicionRol.permiteRegistro)
  .map(([rol]) => rol);

export const etiquetaRol = (rol) => ROLES[rol]?.etiqueta ?? rol;

export const politicaDeRol = (rol) =>
  ROLES[rol]?.politicaClave ?? POLITICA_POR_DEFECTO;

export const requiereCorreoInstitucional = (rol) =>
  Boolean(ROLES[rol]?.requiereCorreoInstitucional);

export const puedeAcceder = (rol, rolesPermitidos) =>
  !rolesPermitidos || rolesPermitidos.includes(rol);
