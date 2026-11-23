package admin

import (
	"net/http"

	"github.com/epistemer/i18n"
	"github.com/epistemer/ui"
	"github.com/epistemer/ui/components"
)

// A flash carries one message to the next page, through a redirect:
// components.Reply(...).Go drops toasts, and ui has no flash (findings
// F25). The cookie holds a key, never text.
const flashCookie = "confab-flash"

// goWith redirects to l and shows msg there.
func goWith(w http.ResponseWriter, r *http.Request, msg i18n.TKer, l ui.Link) error {
	http.SetCookie(w, &http.Cookie{Name: flashCookie, Value: msg.K(), Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	return components.Reply(w, r).Go(l)
}

// takeFlash reads and clears the flash.
func takeFlash(w http.ResponseWriter, r *http.Request) i18n.Ter {
	c, err := r.Cookie(flashCookie)
	if err != nil || c.Value == "" {
		return nil
	}
	http.SetCookie(w, &http.Cookie{Name: flashCookie, Path: "/", MaxAge: -1})
	return i18n.Dynamic(c.Value)
}
