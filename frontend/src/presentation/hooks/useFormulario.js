import { useState } from 'react';

export function useFormulario(valoresIniciales) {
  const [formulario, setFormulario] = useState(valoresIniciales);

  function actualizarCampo(campo) {
    return (evento) => {
      setFormulario((formularioActual) => ({
        ...formularioActual,
        [campo]: evento.target.value,
      }));
    };
  }

  function restablecerFormulario() {
    setFormulario(valoresIniciales);
  }

  return { formulario, actualizarCampo, restablecerFormulario };
}
