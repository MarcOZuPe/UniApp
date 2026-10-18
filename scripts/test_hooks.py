from pathlib import Path
import subprocess, tempfile
root=Path(__file__).resolve().parents[1]
with tempfile.TemporaryDirectory() as t:
 p=Path(t); subprocess.run(['git','init','-q','-b','main',str(p)],check=True)
 for title,ok in [('actualizar readme',False),('docs: actualizar readme',True),('invalid: ejemplo',False),('feat(api)!: cambiar contrato',True),('docs: '+ 'x'*73,False)]:
  f=p/'msg'; f.write_text(title+'\n')
  x=subprocess.run(['bash',str(root/'.githooks/commit-msg'),str(f)],cwd=p,capture_output=True,text=True)
  assert (x.returncode==0)==ok,(title,x.stdout)
  print(('ACEPTADO' if ok else 'RECHAZADO')+': '+title)
 for branch,ok in [('main',True),('feature/sesion-base',True),('bugfix/asistencia',True),('rama_invalida',False)]:
  subprocess.run(['git','symbolic-ref','HEAD','refs/heads/'+branch],cwd=p,check=True)
  x=subprocess.run(['bash',str(root/'.githooks/pre-commit')],cwd=p,capture_output=True,text=True)
  assert (x.returncode==0)==ok,branch
  print(('ACEPTADA' if ok else 'RECHAZADA')+': '+branch)
print('9 casos correctos. Pruebas técnicas, no evidencia de colaboración.')
