package frontend

import "embed"

// Директива //go:embed теперь находится в пакете 'frontend'
// и может безопасно смотреть на папку 'dist' внутри себя.
//
//go:embed all:dist
var StaticFiles embed.FS
