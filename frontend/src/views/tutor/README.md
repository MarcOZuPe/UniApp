`ExpedienteTutor.jsx` permite presentar el expediente propio y ver su estado.
`RevisarTutores.jsx` es la bandeja de coordinación. Las reglas de negocio y los
permisos se aplican en el backend, no en estas vistas.

El catálogo I07 consulta los perfiles públicos mediante `apiTutor.js`. Esa ruta
solo devuelve expedientes aprobados y vigentes y omite documentos y contactos.
