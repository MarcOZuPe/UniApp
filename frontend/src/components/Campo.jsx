import { useId } from 'react';

export default function Campo({ etiqueta, error, ayuda, ...propiedades }) {
  const id = useId();
  const descritoPor = [ayuda && `${id}-ayuda`, error && `${id}-error`].filter(Boolean).join(' ') || undefined;
  return (
    <div className="campo">
      <label htmlFor={id}>{etiqueta}</label>
      <input id={id} aria-invalid={!!error} aria-describedby={descritoPor} {...propiedades} />
      {ayuda && <small id={`${id}-ayuda`}>{ayuda}</small>}
      {error && <small id={`${id}-error`} className="error" role="alert">{error}</small>}
    </div>
  );
}
