import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  cambiarAutorizacion,
  historialAutorizaciones,
  listarVinculaciones,
} from '../../controllers/apoderado/apiApoderado.js';

function formatearFecha(valor) {
  if (!valor) return '—';
  const fecha = new Date(valor);
  if (Number.isNaN(fecha.getTime())) return '—';
  return fecha.toLocaleString('es-PE', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export default function PanelEscolares() {
  const [vinculaciones, setVinculaciones] = useState([]);
  const [historiales, setHistoriales] = useState({});
  const [mensaje, setMensaje] = useState('');
  const [cargando, setCargando] = useState(true);
  const [cambiando, setCambiando] = useState('');

  async function cargarHistoriales(lista) {
    const aceptadas = lista.filter((vinculacion) => vinculacion.estado === 'aceptada');
    const entradas = await Promise.all(
      aceptadas.map(async (vinculacion) => {
        try {
          const respuesta = await historialAutorizaciones(vinculacion.escolar.id);
          return [vinculacion.escolar.id, respuesta.autorizaciones || []];
        } catch {
          return [vinculacion.escolar.id, null];
        }
      })
    );
    setHistoriales(Object.fromEntries(entradas));
  }

  async function cargarVinculaciones() {
    try {
      const respuesta = await listarVinculaciones();
      const lista = respuesta.vinculaciones || [];
      setVinculaciones(lista);
      await cargarHistoriales(lista);
    } catch (error) {
      setMensaje(error.message);
    } finally {
      setCargando(false);
    }
  }

  useEffect(() => {
    cargarVinculaciones();
  }, []);

  async function alternarAutorizacion(vinculacion) {
    const autorizado = vinculacion.autorizacion?.autorizado === true;
    setCambiando(vinculacion.escolar.id);
    setMensaje('');
    try {
      const respuesta = await cambiarAutorizacion(vinculacion.escolar.id, !autorizado);
      setMensaje(respuesta.mensaje);
      await cargarVinculaciones();
    } catch (error) {
      setMensaje(error.message);
    } finally {
      setCambiando('');
    }
  }

  return (
    <main className="pagina-apoderado">
      <h1>Panel de escolares</h1>
      <p>
        Administra la autorización para pedir nuevas tutorías de cada escolar
        vinculado. Debajo de cada uno verás cuándo se otorgó o retiró el permiso.
      </p>

      {mensaje && <p role="alert">{mensaje}</p>}
      {cargando && <p role="status">Cargando…</p>}

      {!cargando && vinculaciones.length === 0 && (
        <p>
          Todavía no has vinculado a ningún escolar.{' '}
          <Link to="/apoderado/vincular">Vincula al primero</Link>.
        </p>
      )}

      {!cargando && vinculaciones.length > 0 && (
        <ul className="lista-escolares">
          {vinculaciones.map((vinculacion) => {
            const autorizado = vinculacion.autorizacion?.autorizado === true;
            const historial = historiales[vinculacion.escolar.id];
            const tituloHistorial = `historial-${vinculacion.id}`;
            return (
              <li key={vinculacion.id} className="escolar-panel">
                <strong>{vinculacion.escolar.nombre_uso}</strong>{' '}
                <span>({vinculacion.escolar.email})</span>
                <br />
                <small>
                  {vinculacion.estado === 'aceptada'
                    ? 'Vinculación aceptada'
                    : 'Pendiente de aceptación'}
                </small>

                {vinculacion.estado === 'aceptada' && (
                  <>
                    <p>
                      <small>
                        {autorizado
                          ? 'Puede solicitar nuevas tutorías.'
                          : 'No puede solicitar nuevas tutorías.'}
                      </small>
                    </p>
                    <button
                      type="button"
                      onClick={() => alternarAutorizacion(vinculacion)}
                      disabled={cambiando === vinculacion.escolar.id}
                    >
                      {cambiando === vinculacion.escolar.id
                        ? 'Guardando…'
                        : autorizado
                          ? 'Revocar autorización'
                          : 'Autorizar nuevas tutorías'}
                    </button>

                    <section
                      className="historial-autorizaciones"
                      aria-labelledby={tituloHistorial}
                    >
                      <h2 id={tituloHistorial}>
                        Historial de autorizaciones y revocaciones
                      </h2>
                      {historial === null ? (
                        <p>No se pudo cargar el historial.</p>
                      ) : (historial || []).length === 0 ? (
                        <p>Sin cambios registrados.</p>
                      ) : (
                        <table>
                          <caption>
                            Cambios de permiso para {vinculacion.escolar.nombre_uso}
                          </caption>
                          <thead>
                            <tr>
                              <th scope="col">Fecha y hora</th>
                              <th scope="col">Cambio</th>
                            </tr>
                          </thead>
                          <tbody>
                            {historial.map((autorizacion) => (
                              <tr key={autorizacion.id}>
                                <td>
                                  <time dateTime={autorizacion.creada_en}>
                                    {formatearFecha(autorizacion.creada_en)}
                                  </time>
                                </td>
                                <td>
                                  {autorizacion.autorizado
                                    ? 'Se otorgó el permiso para pedir tutorías'
                                    : 'Se retiró el permiso para pedir tutorías'}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      )}
                    </section>
                  </>
                )}
              </li>
            );
          })}
        </ul>
      )}

      <p>
        <Link to="/apoderado/vincular">Vincular otro escolar</Link>
      </p>
      <p>
        <Link to="/">Volver al inicio</Link>
      </p>
    </main>
  );
}
