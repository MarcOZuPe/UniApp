export const ROLES = {
  escolar: 'Escolar',
  apoderado: 'Apoderado',
  tutor: 'Tutor UNI',
  coordinador: 'Coordinación',
};
export const ROLES_REGISTRO = ['escolar', 'apoderado', 'tutor'];

export const puedeAcceder = (rol, permitidos) => !permitidos || permitidos.includes(rol);
