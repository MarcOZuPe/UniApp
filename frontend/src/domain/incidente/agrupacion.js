// Agrupa los incidentes por tutor para que coordinación revise primero a quienes
// acumulan más reportes pendientes. Es lógica de dominio independiente de React.
export function agruparPorTutor(incidentes) {
  const grupos = new Map();
  for (const incidente of incidentes) {
    const clave = incidente.tutor_id || incidente.tutor_nombre;
    if (!grupos.has(clave)) {
      grupos.set(clave, {
        clave,
        tutorNombre: incidente.tutor_nombre,
        incidentes: [],
        pendientes: 0,
      });
    }
    const grupo = grupos.get(clave);
    grupo.incidentes.push(incidente);
    if (incidente.estado === 'pendiente') grupo.pendientes += 1;
  }
  return Array.from(grupos.values()).sort(
    (a, b) => b.pendientes - a.pendientes || a.tutorNombre.localeCompare(b.tutorNombre, 'es')
  );
}

export function contarIncidentes(cantidad) {
  return `${cantidad} ${cantidad === 1 ? 'reporte' : 'reportes'}`;
}

export function contarPendientes(cantidad) {
  return `${cantidad} ${cantidad === 1 ? 'pendiente' : 'pendientes'}`;
}
