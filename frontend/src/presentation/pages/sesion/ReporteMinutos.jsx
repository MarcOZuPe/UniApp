import { useCallback, useEffect, useId, useRef, useState } from 'react';

import { reporteApi } from '../../../adapters/http/reporteApi.js';
import { huellaCorta, MAX_DIAS_REPORTE, mesAnterior, validarPeriodo } from '../../../domain/sesion/reporte.js';

const fechaHoraLima = new Intl.DateTimeFormat('es-PE', { timeZone: 'America/Lima', dateStyle: 'short', timeStyle: 'short' });
const fechaLima = new Intl.DateTimeFormat('es-PE', { timeZone: 'America/Lima', dateStyle: 'short' });

// "hasta" llega exclusivo (inicio del dia siguiente): se muestra el dia anterior.
function periodo(registro) {
  const fin = new Date(new Date(registro.hasta).getTime() - 1);
  return `${fechaLima.format(new Date(registro.desde))} – ${fechaLima.format(fin)}`;
}

// guardarArchivo descarga el blob con su nombre; se inyecta en las pruebas.
function guardarEnNavegador(archivo, nombre) {
  const url = URL.createObjectURL(archivo);
  const enlace = document.createElement('a');
  enlace.href = url;
  enlace.download = nombre;
  document.body.append(enlace);
  enlace.click();
  enlace.remove();
  URL.revokeObjectURL(url);
}

function CampoFecha({ id, etiqueta, valor, alCambiar, error, campoRef }) {
  return (
    <div className="campo">
      <label htmlFor={id}>{etiqueta}</label>
      <input id={id} ref={campoRef} type="date" required value={valor}
        onChange={e => alCambiar(e.target.value)}
        aria-invalid={error ? 'true' : 'false'}
        aria-describedby={error ? `${id}-error` : undefined} />
      {error && <small id={`${id}-error`} className="error">{error}</small>}
    </div>
  );
}

// I18: coordinacion descarga el reporte de minutos (CSV o PDF). Cada descarga
// queda registrada con la huella SHA-256 del archivo.
export default function ReporteMinutos({ api = reporteApi, guardarArchivo = guardarEnNavegador, hoy = new Date() }) {
  const ids = useId();
  const [filtro, setFiltro] = useState(() => ({ ...mesAnterior(hoy), tutorId: '' }));
  const [errores, setErrores] = useState({});
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState('');
  const [descarga, setDescarga] = useState(null);
  const [registros, setRegistros] = useState(null);
  const formatoRef = useRef('csv');
  const desdeRef = useRef(null);
  const hastaRef = useRef(null);

  const cargarRegistro = useCallback(async () => {
    try {
      const datos = await api.registro();
      setRegistros(datos?.reportes || []);
    } catch {
      setRegistros([]);
    }
  }, [api]);

  useEffect(() => { cargarRegistro(); }, [cargarRegistro]);

  const actualizar = campo => valor => setFiltro(actual => ({ ...actual, [campo]: valor }));

  async function enviar(evento) {
    evento.preventDefault();
    const formato = formatoRef.current;
    formatoRef.current = 'csv';
    const encontrados = validarPeriodo(filtro);
    setErrores(encontrados);
    setError('');
    setDescarga(null);
    if (encontrados.desde) { desdeRef.current?.focus(); return; }
    if (encontrados.hasta) { hastaRef.current?.focus(); return; }

    setEnviando(true);
    try {
      const resultado = await api.descargar({ ...filtro, tutorId: filtro.tutorId.trim(), formato });
      guardarArchivo(resultado.archivo, resultado.nombre);
      setDescarga(resultado);
      cargarRegistro();
    } catch (err) {
      setError(err?.message || 'No se pudo generar el reporte.');
    } finally {
      setEnviando(false);
    }
  }

  return (
    <main className="pagina-reporte">
      <h1>Reporte de minutos</h1>
      <p>Descarga en CSV o PDF los minutos de un periodo de hasta {MAX_DIAS_REPORTE} días. Los escolares aparecen seudonimizados y solo suman los minutos verificados.</p>

      <p role="note" className="rounded-lg border border-amber-300 bg-amber-50 px-4 py-3 text-sm text-amber-950 dark:border-amber-700 dark:bg-amber-950/40 dark:text-amber-100">
        <strong>Sin validez oficial.</strong> No existe convenio ni registro externo verificado con la UNI; el reporte no acredita horas ante ninguna institución.
      </p>

      <form onSubmit={enviar} noValidate aria-describedby={`${ids}-ayuda`}>
        <p id={`${ids}-ayuda`} className="mt-0 text-sm text-stone-600 dark:text-stone-400">Las fechas se interpretan en hora de Lima; el día de fin se incluye.</p>
        <div className="grid gap-x-4 sm:grid-cols-2">
          <CampoFecha id={`${ids}-desde`} etiqueta="Desde" valor={filtro.desde} alCambiar={actualizar('desde')} error={errores.desde} campoRef={desdeRef} />
          <CampoFecha id={`${ids}-hasta`} etiqueta="Hasta" valor={filtro.hasta} alCambiar={actualizar('hasta')} error={errores.hasta} campoRef={hastaRef} />
        </div>
        <div className="campo">
          <label htmlFor={`${ids}-tutor`}>ID del tutor (opcional)</label>
          <input id={`${ids}-tutor`} type="text" autoComplete="off" spellCheck="false" value={filtro.tutorId}
            onChange={e => actualizar('tutorId')(e.target.value)} aria-describedby={`${ids}-tutor-ayuda`} />
          <small id={`${ids}-tutor-ayuda`}>Vacío para incluir a todos los tutores.</small>
        </div>
        <div className="flex flex-wrap gap-2">
          <button type="submit" disabled={enviando} onClick={() => { formatoRef.current = 'csv'; }}>
            {enviando ? 'Generando…' : 'Descargar CSV'}
          </button>
          <button type="submit" disabled={enviando} onClick={() => { formatoRef.current = 'pdf'; }}
            className="border-stone-400 bg-transparent text-stone-800 hover:bg-stone-100 dark:text-stone-100 dark:hover:bg-stone-800">
            Descargar PDF
          </button>
        </div>
      </form>

      {descarga && (
        <p role="status">
          Se descargó <strong>{descarga.nombre}</strong>. Huella SHA-256:{' '}
          <code title={descarga.sha256}>{huellaCorta(descarga.sha256)}</code>
        </p>
      )}
      {error && <p role="alert" className="error">{error}</p>}

      <h2 id={`${ids}-constancias`}>Constancias de descarga</h2>
      {registros === null ? <p role="status">Cargando constancias…</p> : registros.length === 0 ? (
        <p>Aún no se ha descargado ningún reporte.</p>
      ) : (
        <div role="region" aria-labelledby={`${ids}-constancias`} tabIndex={0}
          className="max-h-96 overflow-auto rounded-lg border border-stone-200 dark:border-stone-800">
          <table className="text-sm">
            <thead className="sticky top-0">
              <tr><th>Generado (Lima)</th><th>Periodo</th><th>Tutor</th><th>Formato</th><th>Filas</th><th>Huella SHA-256</th></tr>
            </thead>
            <tbody>
              {registros.map(registro => (
                <tr key={registro.id}>
                  <td>{fechaHoraLima.format(new Date(registro.generado_en))}</td>
                  <td>{periodo(registro)}</td>
                  <td>{registro.tutor_id || 'Todos'}</td>
                  <td>{registro.formato.toUpperCase()}</td>
                  <td>{registro.filas}</td>
                  <td><code title={registro.sha256}>{huellaCorta(registro.sha256)}</code></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </main>
  );
}
