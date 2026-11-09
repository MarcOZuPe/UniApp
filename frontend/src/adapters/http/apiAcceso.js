import { almacenSesion } from '../../application/acceso/almacenSesion.js';
import { crearClienteHttp } from '../../infrastructure/http/fetchClient.js';
import { conSesionExpirada } from './decoradores.js';

const cerrarSesionLocal = () => almacenSesion.set({ usuario: null, cargando: false });

// Cliente HTTP compartido por los adapters del flujo A.
export const api = conSesionExpirada(crearClienteHttp(), cerrarSesionLocal);
