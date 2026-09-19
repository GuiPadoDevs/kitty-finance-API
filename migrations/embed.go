package migrations

import "embed"

// FS contém todos os arquivos .sql de migração embutidos no binário
//go:embed *.sql
var FS embed.FS
