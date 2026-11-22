// I18: reglas del cliente para pedir el reporte. La autoridad es el servidor;
// esto evita enviar periodos que se rechazarian.

export const MAX_DIAS_REPORTE = 31;
export const FORMATOS_REPORTE = ['csv', 'pdf'];

const FECHA = /^\d{4}-\d{2}-\d{2}$/;
const UN_DIA_MS = 24 * 60 * 60 * 1000;

function diaUTC(texto) {
  if (!FECHA.test(texto || '')) return null;
  const [anio, mes, dia] = texto.split('-').map(Number);
  const fecha = new Date(Date.UTC(anio, mes - 1, dia));
  return fecha.getUTCDate() === dia ? fecha : null;
}

// Devuelve los errores por campo; {} si el periodo es valido. "hasta" es
// inclusivo: del 1 al 31 son 31 dias.
export function validarPeriodo({ desde, hasta }) {
  const errores = {};
  const inicio = diaUTC(desde);
  const fin = diaUTC(hasta);
  if (!inicio) errores.desde = 'Indica la fecha de inicio.';
  if (!fin) errores.hasta = 'Indica la fecha de fin.';
  if (inicio && fin) {
    const dias = (fin - inicio) / UN_DIA_MS + 1;
    if (dias < 1) errores.hasta = 'La fecha de fin no puede ser anterior a la de inicio.';
    else if (dias > MAX_DIAS_REPORTE) errores.hasta = `El periodo puede abarcar como máximo ${MAX_DIAS_REPORTE} días.`;
  }
  return errores;
}

// Primer y ultimo dia del mes anterior al de la fecha dada (formato AAAA-MM-DD).
export function mesAnterior(hoy = new Date()) {
  const inicio = new Date(Date.UTC(hoy.getFullYear(), hoy.getMonth() - 1, 1));
  const fin = new Date(Date.UTC(hoy.getFullYear(), hoy.getMonth(), 0));
  const texto = (fecha) => fecha.toISOString().slice(0, 10);
  return { desde: texto(inicio), hasta: texto(fin) };
}

// Huella abreviada para mostrar; la completa va en el atributo title.
export function huellaCorta(sha256) {
  return sha256 ? `${sha256.slice(0, 12)}…` : '';
}
