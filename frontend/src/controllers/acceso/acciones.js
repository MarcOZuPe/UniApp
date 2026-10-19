import { almacenSesion } from './almacenSesion.js';
import { crearClienteHttp } from './apiCliente.js';
import { conSesionExpirada } from './decoradores.js';

const cerrarSesionLocal = () => almacenSesion.set({ usuario: null, cargando: false });

export const api = conSesionExpirada(crearClienteHttp(), cerrarSesionLocal);

export async function cargarSesion() {
  try {
    const usuario = await api.get('/api/auth/yo');
    almacenSesion.set({ usuario, cargando: false });
  } catch {
    cerrarSesionLocal();
  }
}

export const registrar = (datos) =>
  api.post('/api/auth/registro', {
    email: datos.email,
    nombre_uso: datos.nombre_uso,
    clave: datos.clave,
    rol: datos.rol,
  });

export async function ingresar(email, clave) {
  const usuario = await api.post('/api/auth/ingreso', { email, clave });
  almacenSesion.set({ usuario, cargando: false });
  return usuario;
}

export async function salir() {
  try {
    await api.post('/api/auth/salida');
  } finally {
    cerrarSesionLocal();
  }
}
