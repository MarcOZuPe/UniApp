import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

import { minutosApi } from '../../../adapters/http/minutosApi.js';
import { ETIQUETAS_ESTADO, MAX_MINUTOS, minutosValidos, motivoValido } from '../../../domain/sesion/minutos.js';

import '../../styles/sesion.css';

function TarjetaPendiente({ registro, api, alCambiar }) {
  const [minutos, setMinutos] = useState('');
  const [motivo, setMotivo] = useState('');
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState('');
  const [auditoria, setAuditoria] = useState(null);
  const enDisputa = registro.estado === 'en_disputa';

  async function ejecutar(accion) {
    setEnviando(true);
    setError('');
    try {
      await accion();
      alCambiar();
    } catch (err) {
      setError(err?.message || 'No se pudo completar la acción.');
      setEnviando(false);
    }
  }

  async function verAuditoria() {
    try {
      const datos = await api.auditoria(registro.sesion_id);
      setAuditoria(datos.auditoria || []);
    } catch (err) {
      setError(err?.message || 'No se pudo cargar la auditoría.');
    }
  }

  const id = registro.sesion_id;
  const tituloId = `titulo-${id}`;
  const formularioId = `formulario-${id}`;

  return (
    <li>
      <article className="sesion-tarjeta minutos-tarjeta" aria-labelledby={tituloId}>
        <header className="minutos-cabecera">
          <h2 id={tituloId}>Sesión {id}</h2>
          <span className={`minutos-estado ${enDisputa ? 'en-disputa' : 'propuesto'}`}>
            {ETIQUETAS_ESTADO[registro.estado] || registro.estado}
          </span>
        </header>

        <dl className="minutos-datos">
          <div>
            <dt>Propuestos</dt>
            <dd>{registro.minutos_propuestos} min</dd>
          </div>
          {registro.minutos_verificados !== null && (
            <div>
              <dt>Último valor verificado</dt>
              <dd>{registro.minutos_verificados} min</dd>
            </div>
          )}
          <div>
            <dt>Suman al total</dt>
            <dd>{enDisputa ? 0 : registro.minutos_computables ?? 0} min</dd>
          </div>
        </dl>

        {!enDisputa && (
          <div className="minutos-aprobar">
            <button type="button" className="sesion-declarar" disabled={enviando}
              onClick={() => ejecutar(() => api.aprobar(id))}>
              Aprobar {registro.minutos_propuestos} min
            </button>
            <p className="sesion-zona">Aprueba los minutos tal como los propuso el tutor.</p>
          </div>
        )}

        <form id={formularioId} className="minutos-formulario" aria-labelledby={`${formularioId}-titulo`}
          onSubmit={evento => {
            evento.preventDefault();
            const llamada = enDisputa ? api.resolver : api.ajustar;
            ejecutar(() => llamada(id, Number(minutos), motivo.trim()));
          }}>
          <h3 id={`${formularioId}-titulo`}>{enDisputa ? 'Resolver la disputa' : 'O ajustar con un motivo'}</h3>
          <div className="minutos-campos">
            <div className="minutos-campo minutos-campo-numero">
              <label htmlFor={`minutos-${id}`}>{enDisputa ? 'Minutos finales' : 'Ajustar a'}</label>
              <input id={`minutos-${id}`} name="minutos" type="number" inputMode="numeric" min={0} max={MAX_MINUTOS} step={1}
                aria-describedby={`minutos-${id}-ayuda`}
                value={minutos} onChange={e => setMinutos(e.target.value)} disabled={enviando} />
              <span id={`minutos-${id}-ayuda`} className="minutos-ayuda">Entre 0 y {MAX_MINUTOS}.</span>
            </div>
            <div className="minutos-campo">
              <label htmlFor={`motivo-${id}`}>Motivo (queda en la auditoría)</label>
              <textarea id={`motivo-${id}`} name="motivo" rows={2} value={motivo}
                onChange={e => setMotivo(e.target.value)} disabled={enviando} />
            </div>
          </div>
          <button type="submit" className="sesion-declarar"
            disabled={enviando || !minutosValidos(minutos, 0) || !motivoValido(motivo)}>
            {enDisputa ? 'Resolver disputa' : 'Ajustar minutos'}
          </button>
        </form>

        <footer className="minutos-pie">
          <button type="button" className="minutos-secundario" onClick={verAuditoria}
            aria-expanded={auditoria !== null} aria-controls={`auditoria-${id}`}>
            Ver auditoría
          </button>
          {auditoria && (
            <ol id={`auditoria-${id}`} className="minutos-auditoria" aria-label={`Auditoría de ${id}`}>
              {auditoria.map((entrada, i) => (
                <li key={i}>
                  {entrada.accion} por {entrada.actor_id}
                  {entrada.valor_nuevo !== null && ` → ${entrada.valor_nuevo} min`}
                  {entrada.motivo && ` (${entrada.motivo})`}
                </li>
              ))}
            </ol>
          )}
        </footer>
        {error && <p role="alert" className="error">{error}</p>}
      </article>
    </li>
  );
}

// I17: bandeja de coordinación. No autoaprueba: cada verificación es una
// acción explícita del coordinador y queda auditada.
export default function RevisionMinutos({ api = minutosApi }) {
  const [pendientes, setPendientes] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState('');

  const cargar = useCallback(async () => {
    setCargando(true);
    setError('');
    try {
      const datos = await api.pendientes();
      setPendientes(datos.pendientes || []);
    } catch (err) {
      setError(err?.message || 'No se pudo cargar la bandeja.');
    } finally {
      setCargando(false);
    }
  }, [api]);

  useEffect(() => { cargar(); }, [cargar]);

  return (
    <main className="pagina-sesion">
      <p><Link to="/coordinacion">Volver a coordinación</Link></p>
      <header className="sesion-encabezado">
        <p className="sesion-sobrelinea">FLUJO E · MINUTOS VERIFICADOS</p>
        <h1>Revisión de minutos</h1>
        <p>Propuestos y en disputa. Los minutos solo suman al total después de tu verificación.</p>
      </header>
      {cargando ? <p role="status">Cargando...</p> : pendientes.length === 0 ? (
        <p>No hay minutos pendientes de revisión.</p>
      ) : (
        <ul className="minutos-lista">
          {pendientes.map(registro => (
            <TarjetaPendiente key={registro.sesion_id} registro={registro} api={api} alCambiar={cargar} />
          ))}
        </ul>
      )}
      {error && <p role="alert" className="error">{error}</p>}
    </main>
  );
}
