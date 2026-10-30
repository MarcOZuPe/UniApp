const ESTADO_INICIAL = { usuario: null, cargando: true };

function crearAlmacen() {
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

let instancia = null;

export function getInstance() {
  if (!instancia) {
    instancia = crearAlmacen();
  }
  return instancia;
}

export const almacenSesion = getInstance();
