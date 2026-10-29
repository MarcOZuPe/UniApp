import { useSyncExternalStore } from 'react';
import { almacenSesion } from '../../application/acceso/almacenSesion.js';

export const useSesion = () =>
  useSyncExternalStore(almacenSesion.subscribe, almacenSesion.getState);
