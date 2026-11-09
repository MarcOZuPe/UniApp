import { api } from '../../adapters/http/apiAcceso.js';

export const solicitarRecuperacion = (email) =>
  api.post('/api/auth/recuperar', { email });

export const restablecerClave = ({ token, clave }) =>
  api.post('/api/auth/restablecer', { token, clave });
