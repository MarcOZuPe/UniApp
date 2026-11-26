import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

import {
  listarAuditoriaCoordinacion,
  obtenerMetricasCoordinacion,
  obtenerPanelCoordinacion,
} from '../../../application/acceso/coordinacion.js';
import { useSesion } from '../../hooks/useSesion.js';
import '../../styles/panel-coordinacion.css';

const ETIQUETAS_ACCION = {
  'incidente.revisado': 'Incidente revisado',
  'expediente.revisado': 'Expediente revisado',
  'curso.revisado': 'Curso revisado',
  'autorizacion.cambiada': 'Autorización cambiada',
  'cupo.actualizado': 'Cupo actualizado',
};

export default function Coordinacion() {
  const { usuario } = useSesion();
  const { panel, metricas, auditoria, mensajeError } = usePanelCoordinacion();

  return (
    <main className="pagina-coordinacion">
      <h1>Zona de coordinación</h1>
      <p>Hola, {usuario?.nombre_uso}.</p>
      {mensajeError && <p className="error" role="alert">{mensajeError}</p>}
      {!panel && !mensajeError && <p role="status">Cargando panel…</p>}
      {panel && <PanelCoordinacion panel={panel} />}
      <p><Link to="/incidentes">Revisar incidentes confidenciales</Link></p>
      {metricas && <MetricasAgregadas metricas={metricas} />}
      {auditoria && <Auditoria eventos={auditoria} />}
      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}

function usePanelCoordinacion() {
  const [panel, setPanel] = useState(null);
  const [metricas, setMetricas] = useState(null);
  const [auditoria, setAuditoria] = useState(null);
  const [mensajeError, setMensajeError] = useState('');

  useEffect(() => {
    let consultaVigente = true;

    Promise.all([
      obtenerPanelCoordinacion(),
      obtenerMetricasCoordinacion(),
      listarAuditoriaCoordinacion(50),
    ])
      .then(([panelRecibido, metricasRecibidas, auditoriaRecibida]) => {
        if (!consultaVigente) return;
        setPanel(panelRecibido);
        setMetricas(metricasRecibidas);
        setAuditoria(auditoriaRecibida?.eventos ?? []);
      })
      .catch((errorConsulta) => {
        if (consultaVigente) setMensajeError(errorConsulta.message);
      });

    return () => {
      consultaVigente = false;
    };
  }, []);

  return { panel, metricas, auditoria, mensajeError };
}

function PanelCoordinacion({ panel }) {
  return (
    <section aria-labelledby="titulo-panel">
      <h2 id="titulo-panel">{panel.mensaje}</h2>
      <p><Link to="/coordinacion/tutores">Revisar expedientes de tutores</Link></p>
      <p><Link to="/coordinacion/tutores/aprobados">Ver expedientes aprobados</Link></p>
      <p><Link to="/coordinacion/catalogo">Importar y aprobar el catálogo escolar</Link></p>
    </section>
  );
}

function MetricasAgregadas({ metricas }) {
  return (
    <section aria-labelledby="titulo-metricas">
      <h2 id="titulo-metricas">Métricas agregadas</h2>
      <p className="panel-actualizado">
        Actualizado: {formatearFecha(metricas.actualizado_en)}
      </p>
      <div className="panel-metricas">
        <article>
          <h3>Reservas</h3>
          <dl>
            <Dato etiqueta="Totales" valor={metricas.reservas.totales} />
            <Dato etiqueta="Próximas" valor={metricas.reservas.futuras} />
            <Dato etiqueta="Finalizadas" valor={metricas.reservas.finalizadas} />
          </dl>
        </article>
        <article>
          <h3>Asistencia</h3>
          <dl>
            <Dato etiqueta="Sesiones finalizadas" valor={metricas.asistencia.sesiones_finalizadas} />
            <Dato
              etiqueta="Asistencia escolar"
              valor={`${metricas.asistencia.con_asistencia_escolar} (${metricas.asistencia.tasa_escolar_porcentaje}%)`}
            />
            <Dato
              etiqueta="Asistencia tutor"
              valor={`${metricas.asistencia.con_asistencia_tutor} (${metricas.asistencia.tasa_tutor_porcentaje}%)`}
            />
          </dl>
        </article>
        <article>
          <h3>Horas internas</h3>
          <p className="panel-cifra">{formatearHoras(metricas.horas.internas)}</p>
          <small>Horas de sesiones finalizadas.</small>
        </article>
        <article>
          <h3>Pendientes de gestión</h3>
          <dl>
            <Dato etiqueta="Expedientes" valor={metricas.gestion.expedientes_pendientes} />
            <Dato etiqueta="Incidentes" valor={metricas.gestion.incidentes_pendientes} />
            <Dato etiqueta="Cursos" valor={metricas.gestion.cursos_pendientes} />
            <Dato etiqueta="Solicitudes de colegio" valor={metricas.gestion.solicitudes_colegio_pendientes} />
          </dl>
        </article>
      </div>
    </section>
  );
}

function Dato({ etiqueta, valor }) {
  return (
    <>
      <dt>{etiqueta}</dt>
      <dd>{valor}</dd>
    </>
  );
}

function Auditoria({ eventos }) {
  return (
    <section aria-labelledby="titulo-auditoria">
      <h2 id="titulo-auditoria">Auditoría de decisiones y reglas</h2>
      <p>
        Registro de las últimas decisiones y cambios.
      </p>
      {eventos.length === 0 ? (
        <p>Aún no hay eventos de auditoría registrados.</p>
      ) : (
        <div className="panel-tabla-scroll">
          <table className="panel-auditoria">
            <thead>
              <tr>
                <th scope="col">Fecha</th>
                <th scope="col">Acción</th>
                <th scope="col">Actor</th>
                <th scope="col">Entidad</th>
                <th scope="col">Detalle</th>
              </tr>
            </thead>
            <tbody>
              {eventos.map((evento) => (
                <tr key={evento.id}>
                  <td>{formatearFecha(evento.ocurrido_en)}</td>
                  <td>{ETIQUETAS_ACCION[evento.accion] ?? evento.accion}</td>
                  <td>
                    {evento.actor_nombre || 'Sistema'}
                    <small>{evento.actor_rol}</small>
                  </td>
                  <td>
                    {nombreEntidad(evento.entidad)}
                    <small>{evento.entidad_id}</small>
                  </td>
                  <td>{resumirDetalle(evento)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

function formatearFecha(fecha) {
  if (!fecha) return '—';
  const valor = new Date(fecha);
  if (Number.isNaN(valor.getTime())) return '—';
  return valor.toLocaleString('es-PE', {
    timeZone: 'America/Lima',
    dateStyle: 'short',
    timeStyle: 'short',
  });
}

function formatearHoras(valor) {
  const numero = Number(valor ?? 0);
  return `${numero.toLocaleString('es-PE', { maximumFractionDigits: 2 })} h`;
}

function nombreEntidad(entidad) {
  return {
    incidente: 'Incidente',
    tutor_expediente: 'Expediente',
    catalogo_curso: 'Curso',
    autorizacion: 'Autorización',
    cupo_semanal: 'Cupo semanal',
  }[entidad] ?? entidad;
}

function resumirDetalle(evento) {
  const detalle = evento.detalle || {};
  switch (evento.accion) {
    case 'incidente.revisado':
      return `Resultado: ${detalle.resultado || '—'}. Estado: ${detalle.estado || '—'}.`;
    case 'expediente.revisado':
      return `Estado: ${detalle.estado || '—'}.`;
    case 'curso.revisado':
      return `${detalle.curso || 'Curso'} (${detalle.grado || '—'}): ${detalle.estado || '—'}.`;
    case 'autorizacion.cambiada':
      return `Escolar ${detalle.escolar_id || '—'}: ${detalle.estado || '—'}.`;
    case 'cupo.actualizado':
      return detalle.minutos_anteriores
        ? `${detalle.minutos_anteriores} → ${detalle.minutos} minutos.`
        : `${detalle.minutos || '—'} minutos.`;
    default:
      return 'Evento registrado.';
  }
}
