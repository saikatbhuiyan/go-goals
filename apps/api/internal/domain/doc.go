// Package domain contains enterprise concepts shared by API modules.
//
// Keep this package free of transport, SQL, session, and template concerns.
// If a type starts depending on HTTP or PostgreSQL details, it belongs in an
// application module or infrastructure adapter instead.
package domain
