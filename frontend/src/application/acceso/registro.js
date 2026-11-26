import { api } from '../../adapters/http/apiAcceso.js';

export const registrar = (formularioRegistro) =>
  api.post('/api/auth/registro', cuerpoDeRegistro(formularioRegistro));

export const listarFacultades = () => api.get('/api/facultades');

function cuerpoDeRegistro({ email, nombre_uso, clave, rol, facultad_id }) {
  return { email, nombre_uso, clave, rol, facultad_id };
}
