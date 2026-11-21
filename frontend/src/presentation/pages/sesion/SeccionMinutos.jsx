import { useEffect, useState } from 'react';

import { minutosApi } from '../../../adapters/http/minutosApi.js';
import {
  ETIQUETAS_ESTADO,
  MAX_MINUTOS,
  minutosValidos,
  motivoValido,
  puedeDisputar,
} from '../../../domain/sesion/minutos.js';

// I17: el tutor propone los minutos dictados y el escolar (o el tutor ante un
// ajuste a la baja) puede disputarlos. Solo la coordinación los verifica.
export default function SeccionMinutos({ sesion, usuario, api = minutosApi }) {
  const [registro, setRegistro] = useState(null);
  const [cargando, setCargando] = useState(true);
  const [minutos, setMinutos] = useState('');
  const [motivo, setMotivo] = useState('');
  const [enviando, setEnviando] = useState(false);
  const [mensaje, setMensaje] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    let activo = true;
    setCargando(true);
    api.obtener(sesion.id)
      .then(datos => { if (activo) setRegistro(datos); })
      .catch(err => {
        if (!activo) return;
        if (err?.estado === 404) setRegistro(null);
        else setError(err?.message || 'No se pudieron cargar los minutos.');
      })
      .finally(() => { if (activo) setCargando(false); });
    return () => { activo = false; };
  }, [api, sesion.id]);

  async function ejecutar(accion, textoExito) {
    setEnviando(true);
    setMensaje('');
    setError('');
    try {
      setRegistro(await accion());
      setMensaje(textoExito);
      setMinutos('');
      setMotivo('');
    } catch (err) {
      setError(err?.message || 'No se pudo completar la acción.');
    } finally {
      setEnviando(false);
    }
  }

  const esTutor = usuario?.id === sesion.tutor_id;
  const finalizada = new Date(sesion.fin) <= new Date();
  const disputable = puedeDisputar(registro, usuario?.id);

  return (
    <section className="sesion-tarjeta sesion-minutos" aria-labelledby="minutos-sesion">
      <h2 id="minutos-sesion">Minutos verificados</h2>
      {cargando ? <p>Cargando minutos...</p> : registro ? (
        <>
          <p><strong>Estado:</strong> {ETIQUETAS_ESTADO[registro.estado] || registro.estado}</p>
          <p><strong>Propuestos por el tutor:</strong> {registro.minutos_propuestos} min</p>
          {registro.minutos_verificados !== null && (
            <p><strong>Verificados por coordinación:</strong> {registro.minutos_verificados} min</p>
          )}
          <p><strong>Suman al total:</strong> {registro.minutos_computables} min</p>
          {disputable && (
            <form onSubmit={evento => {
              evento.preventDefault();
              ejecutar(() => api.disputar(sesion.id, motivo.trim()), 'Disputa registrada. Los minutos quedan congelados.');
            }}>
              <label htmlFor="motivo-disputa">Motivo de la disputa</label>
              <textarea id="motivo-disputa" rows={3} value={motivo} onChange={e => setMotivo(e.target.value)} disabled={enviando} />
              <button className="sesion-declarar" type="submit" disabled={enviando || !motivoValido(motivo)}>
                Disputar minutos
              </button>
            </form>
          )}
        </>
      ) : esTutor && finalizada ? (
        <form onSubmit={evento => {
          evento.preventDefault();
          ejecutar(() => api.proponer(sesion.id, Number(minutos)), 'Minutos propuestos. Coordinación los revisará.');
        }}>
          <label htmlFor="minutos-dictados">Minutos realmente dictados</label>
          <input
            id="minutos-dictados"
            type="number"
            min={1}
            max={MAX_MINUTOS}
            step={1}
            value={minutos}
            onChange={e => setMinutos(e.target.value)}
            disabled={enviando}
          />
          <button className="sesion-declarar" type="submit" disabled={enviando || !minutosValidos(minutos)}>
            Proponer minutos
          </button>
          <p><small>Solo se pueden proponer si tutor y escolar declararon asistencia.</small></p>
        </form>
      ) : (
        <p>{esTutor ? 'Podrás proponer los minutos cuando termine la sesión.' : 'El tutor aún no ha propuesto los minutos.'}</p>
      )}
      <p><small>Nada suma al total hasta que coordinación lo verifica. Una disputa congela los minutos y una cancelación los deja en cero.</small></p>
      {mensaje && <p className="sesion-exito" role="status">{mensaje}</p>}
      {error && <p role="alert" className="error">{error}</p>}
    </section>
  );
}
