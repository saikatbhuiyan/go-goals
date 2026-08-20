package sessions

import gorillasessions "github.com/gorilla/sessions"

func NewCookieStore(secret string) *gorillasessions.CookieStore {
	return gorillasessions.NewCookieStore([]byte(secret))
}
