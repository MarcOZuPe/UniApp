const ESTADO_INICIAL = { usuario: null, cargando: true };

export function crearAlmacen() {
  let estadoSesion = { ...ESTADO_INICIAL };
  const suscriptores = new Set();

  function notificarSuscriptores() {
    suscriptores.forEach((suscriptor) => suscriptor());
  }

  return {
    getState: () => estadoSesion,
    subscribe(suscriptor) {
      suscriptores.add(suscriptor);
      return () => suscriptores.delete(suscriptor);
    },
    set(cambios) {
      estadoSesion = { ...estadoSesion, ...cambios };
      notificarSuscriptores();
    },
  };
}

export const almacenSesion = crearAlmacen();
