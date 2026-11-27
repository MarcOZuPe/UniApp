import { api } from '../../adapters/http/apiAcceso.js';

export const cambiarClavePerfil = ({ clave_actual, clave_nueva }) =>
  api.put('/api/perfil/clave', { clave_actual, clave_nueva });
