# Configuración de cuenta y colaboración

Repositorio: https://github.com/MarcOZuPe/UniApp

## Autenticación personal en la VPS

```bash
cd ~/UniApp
gh auth login --hostname github.com --git-protocol https --web
gh auth setup-git
gh auth refresh -s project
git push -u origin main
```

El código temporal se introduce en el navegador de tu cuenta. No compartas el token ni lo guardes en archivos del repositorio. Si `workflow` falta para publicar Actions, usa `gh auth refresh -s workflow`.

Después se configura el Project público con Status: Backlog, In progress, In review, Done; las variables PROJECT_OWNER=MarcOZuPe y PROJECT_NUMBER con el número real. Guardar el token con permisos repo y project en el secreto GH_TOKEN desde Settings → Secrets and variables → Actions. El comando interactivo `gh secret set GH_TOKEN` permite ingresarlo sin incluirlo en la línea de comandos.

Proteger main con PR obligatorio, una aprobación y conversaciones resueltas; no habilitar force push ni eliminación. Capturar las reglas guardadas y adjuntar la imagen en Evidencias.

## Registro al curso

Hacer fork de https://github.com/AldoLunaBueno/ds-lab-1-gestion y modificar únicamente la línea de Marco Antonio Zúñiga Peralta en registro_estudiantes.md para agregar https://github.com/MarcOZuPe/UniApp. Enviar PR al repositorio del profesor. El nombre del laboratorio es recomendado; aquí se usa UniApp por decisión del propietario.

## Simulación real

Invitar un compañero; crear y asignarle un issue; agregarlo al tablero en Backlog. El compañero crea una rama feature/, captura el rechazo de un commit incorrecto y la aceptación del correcto, agrega su nombre a Colaboradores y abre PR borrador con Closes #N. Al marcarlo listo pasa a In review. Marco comenta, revisa y aprueba; el compañero resuelve la conversación; al fusionar pasa a Done. Adjuntar capturas y enlaces reales. No autoaprobar ni inventar pruebas de participación.
