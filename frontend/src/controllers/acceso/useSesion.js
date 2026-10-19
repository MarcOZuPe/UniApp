import { useSyncExternalStore } from 'react';
import { almacenSesion } from './almacenSesion.js';

export const useSesion = () =>
  useSyncExternalStore(almacenSesion.subscribe, almacenSesion.getState);
