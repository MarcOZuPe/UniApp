import { Link } from 'react-router-dom';

import { usePanelEscolares } from '../../hooks/usePanelEscolares.js';

export default function PanelEscolares() {
  const {
    vinculaciones,
    escolarSeleccionadoID,
    historialesPorEscolar,
    escolarCargandoHistorial,
    mensajeEstado,
    mensajeEsError,
    cargandoVinculaciones,
    escolarCambiandoAutorizacion,
    seleccionarEscolar,
    alternarAutorizacion,
  } = usePanelEscolares();

  const vinculacionSeleccionada = vinculaciones.find(
    (vinculacionCandidata) => vinculacionCandidata.escolar.id === escolarSeleccionadoID
  );
  const estaAutorizado = vinculacionSeleccionada?.autorizacion?.autorizado === true;
  const historialSeleccionado = vinculacionSeleccionada
    ? historialesPorEscolar[vinculacionSeleccionada.escolar.id]
    : undefined;

  return (
    <main className="pagina-apoderado">
      <h1>Panel de escolares</h1>
      <p>
        Selecciona un escolar para administrar su autorización de pedir nuevas
        tutorías y consultar el historial de cambios.
      </p>

      {mensajeEstado && (
        <p role={mensajeEsError ? 'alert' : 'status'}>{mensajeEstado}</p>
      )}
      {cargandoVinculaciones && <p role="status">Cargando…</p>}

      {!cargandoVinculaciones && vinculaciones.length === 0 && (
        <p>
          Todavía no has vinculado a ningún escolar.{' '}
          <Link to="/apoderado/vincular">Vincula al primero</Link>.
        </p>
      )}

      {!cargandoVinculaciones && vinculaciones.length > 0 && (
        <SelectorEscolar
          vinculaciones={vinculaciones}
          escolarSeleccionadoID={escolarSeleccionadoID}
          alSeleccionar={seleccionarEscolar}
        />
      )}

      {!cargandoVinculaciones && vinculaciones.length > 0 && !vinculacionSeleccionada && (
        <p>Selecciona un escolar de la lista para ver su información.</p>
      )}

      {!cargandoVinculaciones && vinculacionSeleccionada && (
        <DetalleEscolar
          vinculacion={vinculacionSeleccionada}
          autorizacion={{
            estaAutorizado,
            historial: historialSeleccionado,
            cargandoHistorial: escolarCargandoHistorial === vinculacionSeleccionada.escolar.id,
            cambiando: escolarCambiandoAutorizacion === vinculacionSeleccionada.escolar.id,
            alAlternar: alternarAutorizacion,
          }}
        />
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

function SelectorEscolar({ vinculaciones, escolarSeleccionadoID, alSeleccionar }) {
  return (
    <div className="selector-escolar">
      <label htmlFor="selector-escolar">Escolar</label>
      <select
        id="selector-escolar"
        value={escolarSeleccionadoID}
        onChange={alSeleccionar}
      >
        <option value="">Selecciona un escolar…</option>
        {vinculaciones.map((vinculacionCandidata) => (
          <option key={vinculacionCandidata.id} value={vinculacionCandidata.escolar.id}>
            {vinculacionCandidata.escolar.nombre_uso} ({vinculacionCandidata.escolar.email})
            {vinculacionCandidata.estado === 'pendiente' ? ' — pendiente' : ''}
          </option>
        ))}
      </select>
    </div>
  );
}

function DetalleEscolar({ vinculacion, autorizacion }) {
  return (
    <section className="escolar-panel" aria-labelledby="detalle-escolar">
      <h2 id="detalle-escolar">{vinculacion.escolar.nombre_uso}</h2>
      <p>{vinculacion.escolar.email}</p>
      <p>
        <small>
          {vinculacion.estado === 'aceptada'
            ? 'Vinculación aceptada'
            : 'Pendiente de aceptación'}
        </small>
      </p>

      {vinculacion.estado === 'aceptada' && (
        <>
          <p>
            <small>
              {autorizacion.estaAutorizado
                ? 'Puede solicitar nuevas tutorías.'
                : 'No puede solicitar nuevas tutorías.'}
            </small>
          </p>
          <button
            type="button"
            onClick={() => autorizacion.alAlternar(vinculacion)}
            disabled={autorizacion.cambiando}
          >
            {autorizacion.cambiando
              ? 'Guardando…'
              : autorizacion.estaAutorizado
                ? 'Revocar autorización'
                : 'Autorizar nuevas tutorías'}
          </button>

          <HistorialAutorizaciones
            historial={autorizacion.historial}
            cargando={autorizacion.cargandoHistorial}
            nombreEscolar={vinculacion.escolar.nombre_uso}
          />
        </>
      )}
    </section>
  );
}

function HistorialAutorizaciones({ historial, cargando, nombreEscolar }) {
  return (
    <section
      className="historial-autorizaciones"
      aria-labelledby="historial-autorizaciones"
    >
      <h3 id="historial-autorizaciones">
        Historial de autorizaciones y revocaciones
      </h3>
      {historial === undefined || cargando ? (
        <p role="status">Cargando historial…</p>
      ) : historial === null ? (
        <p>No se pudo cargar el historial.</p>
      ) : historial.length === 0 ? (
        <p>Sin cambios registrados.</p>
      ) : (
        <table>
          <caption>
            Cambios de permiso para {nombreEscolar}
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
  );
}

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
