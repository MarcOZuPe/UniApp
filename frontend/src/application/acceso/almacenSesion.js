export function crearAlmacen() {
  let estado = { usuario: null, cargando: true };
  const suscriptores = new Set();
  return {
    getState: () => estado,
    subscribe(suscriptor) {
      suscriptores.add(suscriptor);
      return () => suscriptores.delete(suscriptor);
    },
    set(cambios) {
      estado = { ...estado, ...cambios };
      suscriptores.forEach((suscriptor) => suscriptor());
    },
  };
}
export const almacenSesion = crearAlmacen();
