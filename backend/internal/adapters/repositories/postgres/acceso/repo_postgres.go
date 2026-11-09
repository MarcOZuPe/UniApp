package acceso

import "database/sql"

type RepositorioPostgres struct{ DB *sql.DB }
