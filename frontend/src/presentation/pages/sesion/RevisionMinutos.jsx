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

  return (
    <li className="sesion-tarjeta">
      <p><strong>Sesión:</strong> {registro.sesion_id}</p>
      <p><strong>Estado:</strong> {ETIQUETAS_ESTADO[registro.estado] || registro.estado}</p>
      <p><strong>Propuestos:</strong> {registro.minutos_propuestos} min</p>
      {registro.minutos_verificados !== null && <p><strong>Último valor verificado:</strong> {registro.minutos_verificados} min</p>}

      {!enDisputa && (
        <button type="button" className="sesion-declarar" disabled={enviando}
          onClick={() => ejecutar(() => api.aprobar(registro.sesion_id))}>
          Aprobar {registro.minutos_propuestos} min
        </button>
      )}

      <form onSubmit={evento => {
        evento.preventDefault();
        const llamada = enDisputa ? api.resolver : api.ajustar;
        ejecutar(() => llamada(registro.sesion_id, Number(minutos), motivo.trim()));
      }}>
        <label htmlFor={`minutos-${registro.sesion_id}`}>{enDisputa ? 'Minutos finales' : 'Ajustar a'}</label>
        <input id={`minutos-${registro.sesion_id}`} type="number" min={0} max={MAX_MINUTOS} step={1}
          value={minutos} onChange={e => setMinutos(e.target.value)} disabled={enviando} />
        <label htmlFor={`motivo-${registro.sesion_id}`}>Motivo (queda en la auditoría)</label>
        <textarea id={`motivo-${registro.sesion_id}`} rows={2} value={motivo}
          onChange={e => setMotivo(e.target.value)} disabled={enviando} />
        <button type="submit" className="sesion-declarar"
          disabled={enviando || !minutosValidos(minutos, 0) || !motivoValido(motivo)}>
          {enDisputa ? 'Resolver disputa' : 'Ajustar minutos'}
        </button>
      </form>

      <button type="button" onClick={verAuditoria}>Ver auditoría</button>
      {auditoria && (
        <ol aria-label={`Auditoría de ${registro.sesion_id}`}>
          {auditoria.map((entrada, i) => (
            <li key={i}>
              {entrada.accion} por {entrada.actor_id}
              {entrada.valor_nuevo !== null && ` → ${entrada.valor_nuevo} min`}
              {entrada.motivo && ` (${entrada.motivo})`}
            </li>
          ))}
        </ol>
      )}
      {error && <p role="alert" className="error">{error}</p>}
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
    <main className="sesion-pagina">
      <p><Link to="/coordinacion">Volver a coordinación</Link></p>
      <h1>Revisión de minutos</h1>
      <p><small>Propuestos y en disputa. Los minutos solo suman al total después de tu verificación.</small></p>
      {cargando ? <p>Cargando...</p> : pendientes.length === 0 ? (
        <p>No hay minutos pendientes de revisión.</p>
      ) : (
        <ul className="sesion-lista">
          {pendientes.map(registro => (
            <TarjetaPendiente key={registro.sesion_id} registro={registro} api={api} alCambiar={cargar} />
          ))}
        </ul>
      )}
      {error && <p role="alert" className="error">{error}</p>}
    </main>
  );
}
