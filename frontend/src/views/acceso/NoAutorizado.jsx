import { Link } from 'react-router-dom';
export default function NoAutorizado() {
  return (
    <main>
      <h1>Sin permiso</h1>
      <p role="alert">Tu rol no tiene acceso a esta sección.</p>
      <Link to="/">Volver al inicio</Link>
    </main>
  );
}
