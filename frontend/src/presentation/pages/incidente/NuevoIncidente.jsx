import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import {
  crearIncidente,
  listarTutoriasIncidentes,
} from '../../../adapters/http/incidenteApi.js';
import {
  MAX_EVIDENCIAS,
  TAMANO_MAXIMO_EVIDENCIA,
  TIPOS_INCIDENTE,
  formatearTamano,
  validarIncidente,
} from '../../../domain/incidente/validadores.js';
import { useSesion } from '../../hooks/useSesion.js';
import '../../styles/incidentes.css';

const FORMULARIO_INICIAL = {
  reserva_id: '',
  tipo: TIPOS_INCIDENTE[0].valor,
  descripcion: '',
};

export default function NuevoIncidente() {
  const { usuario } = useSesion();
  const navigate = useNavigate();
  const esApoderado = usuario?.rol === 'apoderado';
  const inputEvidencias = useRef(null);

  const [tutorias, setTutorias] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [errorGeneral, setErrorGeneral] = useState('');

  const [formulario, setFormulario] = useState(FORMULARIO_INICIAL);
  const [archivos, setArchivos] = useState([]);
  const [errores, setErrores] = useState({});
  const [enviando, setEnviando] = useState(false);

  const tutoriaSeleccionada = tutorias.find(
    (tutoria) => tutoria.reserva_id === formulario.reserva_id
  );

  useEffect(() => {
    let vigente = true;
    setCargando(true);
    setErrorGeneral('');
    listarTutoriasIncidentes()
      .then((respuesta) => {
        if (vigente) setTutorias(respuesta?.tutorias ?? []);
      })
      .catch((error) => {
        if (vigente) setErrorGeneral(error.message);
      })
      .finally(() => {
        if (vigente) setCargando(false);
      });
    return () => {
      vigente = false;
    };
  }, []);

  function cambiarCampo(campo, valor) {
    setFormulario((actual) => ({ ...actual, [campo]: valor }));
    setErrores((actual) => ({ ...actual, [campo]: '' }));
  }

  function cambiarArchivos(evento) {
    const seleccionados = Array.from(evento.target.files ?? []);
    setArchivos(seleccionados);
    setErrores((actual) => ({ ...actual, evidencias: '' }));
  }

  function quitarArchivo(indice) {
    setArchivos((actual) => actual.filter((_, posicion) => posicion !== indice));
  }

  async function enviarReporte(evento) {
    evento.preventDefault();
    setErrorGeneral('');
    const encontrados = validarIncidente({ ...formulario, archivos });
    if (Object.keys(encontrados).length > 0) {
      setErrores(encontrados);
      return;
    }
    const datos = new FormData();
    datos.append('reserva_id', formulario.reserva_id);
    if (esApoderado && tutoriaSeleccionada) {
      datos.append('escolar_id', tutoriaSeleccionada.escolar_id);
    }
    datos.append('tipo', formulario.tipo);
    datos.append('descripcion', formulario.descripcion.trim());
    archivos.forEach((archivo) => datos.append('evidencias', archivo));

    setEnviando(true);
    setErrores({});
    try {
      await crearIncidente(datos);
      setFormulario(FORMULARIO_INICIAL);
      setArchivos([]);
      if (inputEvidencias.current) inputEvidencias.current.value = '';
      const mensaje = esApoderado
        ? 'Reporte registrado. La evidencia queda disponible solo para ti, el escolar y coordinación.'
        : 'Reporte registrado. La evidencia queda disponible solo para ti, tu apoderado y coordinación.';
      navigate('/incidentes', { state: { mensaje } });
    } catch (error) {
      setErrores(error.campos || {});
      setErrorGeneral(error.message);
    } finally {
      setEnviando(false);
    }
  }

  if (cargando) {
    return (
      <main className="pagina-incidentes">
        <h1>Registrar incidente</h1>
        <p role="status">Cargando tutorías…</p>
      </main>
    );
  }

  return (
    <main className="pagina-incidentes">
      <h1>Registrar incidente</h1>
      <p>
        Reporta una posible falta del tutor ocurrida en una tutoría que ya atendiste.
        La evidencia es confidencial y solo la pueden abrir el escolar, su apoderado y
        coordinación. El tutor denunciado no tiene acceso a esos archivos.
      </p>
      <p><Link to="/incidentes">Volver a incidentes</Link></p>

      {errorGeneral && <p className="error" role="alert">{errorGeneral}</p>}

      <section aria-labelledby="titulo-reporte">
        <h2 id="titulo-reporte">Nuevo reporte</h2>
        {tutorias.length === 0 ? (
          <p>
            Aún no hay tutorías finalizadas disponibles para reportar. Cuando termine
            una sesión, podrás registrar un incidente sobre ella.
          </p>
        ) : (
          <form onSubmit={enviarReporte} noValidate>
            <label className="campo">
              <span>Tutoría atendida</span>
              <select
                value={formulario.reserva_id}
                onChange={(evento) => cambiarCampo('reserva_id', evento.target.value)}
                aria-invalid={Boolean(errores.reserva_id)}
              >
                <option value="">Elige una tutoría</option>
                {tutorias.map((tutoria) => (
                  <option key={tutoria.reserva_id} value={tutoria.reserva_id}>
                    {tutoria.inicio_lima} · {tutoria.tutor_nombre} · {tutoria.curso}
                    {esApoderado ? ` · ${tutoria.escolar_nombre}` : ''}
                  </option>
                ))}
              </select>
            </label>
            {errores.reserva_id && <p className="error">{errores.reserva_id}</p>}

            {tutoriaSeleccionada && (
              <p className="incidente-tutoria">
                Tutor: <strong>{tutoriaSeleccionada.tutor_nombre}</strong>. Curso:{' '}
                {tutoriaSeleccionada.curso}. Asistencia declarada:{' '}
                {tutoriaSeleccionada.asistio ? 'sí' : 'no registrada'}.
              </p>
            )}

            <label className="campo">
              <span>Tipo de incidente</span>
              <select
                value={formulario.tipo}
                onChange={(evento) => cambiarCampo('tipo', evento.target.value)}
                aria-invalid={Boolean(errores.tipo)}
              >
                {TIPOS_INCIDENTE.map((opcion) => (
                  <option key={opcion.valor} value={opcion.valor}>{opcion.etiqueta}</option>
                ))}
              </select>
            </label>
            {errores.tipo && <p className="error">{errores.tipo}</p>}

            <label className="campo" htmlFor="incidente-descripcion">
              <span>Descripción de lo ocurrido</span>
              <textarea
                id="incidente-descripcion"
                rows={6}
                maxLength={2000}
                value={formulario.descripcion}
                onChange={(evento) => cambiarCampo('descripcion', evento.target.value)}
                aria-invalid={Boolean(errores.descripcion)}
              />
            </label>
            <small>{formulario.descripcion.trim().length}/2000 caracteres</small>
            {errores.descripcion && <p className="error">{errores.descripcion}</p>}

            <label className="campo" htmlFor="incidente-evidencias">
              <span>Evidencias obligatorias (PDF, JPG o PNG)</span>
              <input
                id="incidente-evidencias"
                ref={inputEvidencias}
                type="file"
                multiple
                accept=".pdf,.jpg,.jpeg,.png,application/pdf,image/jpeg,image/png"
                onChange={cambiarArchivos}
                aria-invalid={Boolean(errores.evidencias)}
              />
            </label>
            <small>
              De 1 a {MAX_EVIDENCIAS} archivos, hasta 10 MB cada uno. Solo se acepta
              contenido PDF, JPG o PNG.
            </small>
            {errores.evidencias && <p className="error" role="alert">{errores.evidencias}</p>}

            {archivos.length > 0 && (
              <ul className="incidente-archivos">
                {archivos.map((archivo, indice) => (
                  <li key={`${archivo.name}-${indice}`}>
                    <span>{archivo.name} ({formatearTamano(archivo.size)})</span>
                    <button type="button" onClick={() => quitarArchivo(indice)}>
                      Quitar
                    </button>
                  </li>
                ))}
              </ul>
            )}
            {archivos.some((archivo) => archivo.size > TAMANO_MAXIMO_EVIDENCIA) && (
              <p className="error">Cada evidencia debe pesar como máximo 10 MB.</p>
            )}

            <button type="submit" disabled={enviando}>
              {enviando ? 'Registrando…' : 'Registrar incidente'}
            </button>
          </form>
        )}
      </section>

      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}
