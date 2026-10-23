import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

import { obtenerCatalogo, solicitarColegio } from '../../controllers/catalogo/apiCatalogo.js';
import { obtenerPerfilesPublicos } from '../../controllers/tutor/apiTutor.js';
import { useSesion } from '../../controllers/acceso/useSesion.js';

export default function Catalogo() {
  const { usuario } = useSesion();
  const [catalogo, setCatalogo] = useState(null);
  const [perfiles, setPerfiles] = useState([]);
  const [cargando, setCargando] = useState(true);
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState('');
  const [confirmacion, setConfirmacion] = useState('');

  useEffect(() => {
    let vigente = true;
    Promise.all([obtenerCatalogo(), obtenerPerfilesPublicos()])
      .then(([datos, resultadoPerfiles]) => {
        if (!vigente) return;
        setCatalogo(datos);
        setPerfiles(resultadoPerfiles.perfiles);
      })
      .catch((fallo) => {
        if (vigente) setError(fallo.message);
      })
      .finally(() => {
        if (vigente) setCargando(false);
      });
    return () => {
      vigente = false;
    };
  }, []);

  async function enviarSolicitud(evento) {
    evento.preventDefault();
    setError('');
    setConfirmacion('');
    const formularioDOM = evento.currentTarget;
    const formulario = new FormData(formularioDOM);
    setEnviando(true);
    try {
      await solicitarColegio({
        nombre: formulario.get('nombre'),
        distrito: formulario.get('distrito'),
      });
      formularioDOM.reset();
      setConfirmacion('La solicitud quedó registrada para revisión.');
    } catch (fallo) {
      setError(fallo.message);
    } finally {
      setEnviando(false);
    }
  }

  if (cargando) return <main><p role="status">Cargando catálogo…</p></main>;

  return (
    <main className="pagina-catalogo">
      <h1>Catálogo de cursos y tutores</h1>
      {error && <p className="error" role="alert">{error}</p>}
      {catalogo && (
        <>
          <p>
            {catalogo.fuente.toLocaleLowerCase().startsWith('sintetico:')
              ? 'Datos de demostración; no representan un padrón institucional.'
              : <>Origen: <strong>{catalogo.fuente}</strong></>}
            {' · Actualizado '}
            {new Date(catalogo.importado_en).toLocaleString()}
          </p>
          <section aria-labelledby="cursos-titulo">
            <h2 id="cursos-titulo">Cursos aprobados por grado</h2>
            {catalogo.cursos.length === 0 ? (
              <p>Aún no hay cursos aprobados por coordinación.</p>
            ) : (
              <ul>
                {catalogo.cursos.map((curso) => (
                  <li key={curso.id}>{curso.grado} · {curso.nombre}</li>
                ))}
              </ul>
            )}
          </section>
          <section aria-labelledby="colegios-titulo">
            <h2 id="colegios-titulo">Colegios del padrón</h2>
            <ul>
              {catalogo.colegios.map((colegio) => (
                <li key={`${colegio.nombre}-${colegio.distrito}`}>
                  {colegio.nombre} · {colegio.distrito}
                </li>
              ))}
            </ul>
          </section>
        </>
      )}
      <section aria-labelledby="tutores-titulo">
        <h2 id="tutores-titulo">Perfiles públicos de tutores vigentes</h2>
        {perfiles.length === 0 ? (
          <p>No hay perfiles públicos disponibles.</p>
        ) : (
          perfiles.map((perfil) => (
            <article className="perfil-publico" key={perfil.tutor_id}>
              <h3>{perfil.presentacion}</h3>
              <p><strong>Cursos:</strong> {perfil.cursos}</p>
              <p><strong>Experiencia/formación:</strong> {perfil.experiencia}</p>
            </article>
          ))
        )}
        <p>La información de contacto y los documentos de verificación son privados.</p>
      </section>
      {usuario && ['escolar', 'apoderado'].includes(usuario.rol) && (
        <section aria-labelledby="solicitud-colegio-titulo">
          <h2 id="solicitud-colegio-titulo">¿Falta tu colegio?</h2>
          <p>Envía una solicitud para que coordinación revise su incorporación al padrón.</p>
          <form onSubmit={enviarSolicitud}>
            <label className="campo">
              Nombre del colegio
              <input name="nombre" required maxLength={160} />
            </label>
            <label className="campo">
              Distrito
              <input name="distrito" required maxLength={100} />
            </label>
            <button type="submit" disabled={enviando}>
              {enviando ? 'Enviando…' : 'Solicitar incorporación'}
            </button>
          </form>
          {confirmacion && <p role="status">{confirmacion}</p>}
        </section>
      )}
      {!usuario && <p>Inicia sesión como escolar o apoderado para solicitar un colegio faltante.</p>}
      <p><Link to="/">Volver al inicio</Link></p>
    </main>
  );
}
