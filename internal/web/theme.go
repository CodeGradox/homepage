package web

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

// The visitor's theme preference is stored in a cookie so it survives across
// visits without any client-side JavaScript. "system" is never stored — the
// absence of the cookie means "follow the OS preference".
const themeCookie = "theme"

// themeFromRequest reads the theme cookie, treating anything unexpected as
// "system" so a stale or tampered cookie can't break rendering.
func themeFromRequest(r *http.Request) string {
	if c, err := r.Cookie(themeCookie); err == nil {
		if v := c.Value; v == "light" || v == "dark" {
			return v
		}
	}
	return "system"
}

// setTheme stores the visitor's theme choice in a cookie and sends them back to
// the page they came from. Choosing "system" clears the cookie instead, so the
// site falls back to the OS preference.
func (s *Server) setTheme(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     themeCookie,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	switch theme := r.FormValue("theme"); theme {
	case "light", "dark":
		cookie.Value = theme
		cookie.MaxAge = int((365 * 24 * time.Hour).Seconds())
	case "system":
		cookie.MaxAge = -1
	default:
		http.Error(w, "invalid theme", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, cookie)

	// #nosec G710 -- returnPath only follows Referers whose host matches this
	// request's host and strips them to a plain path, mirroring Rails'
	// redirect_back_or_to; everything else falls back to "/".
	http.Redirect(w, r, returnPath(r), http.StatusSeeOther)
}

// returnPath mirrors Rails' redirect_back_or_to: follow the Referer when it
// points at the host serving this request (or is a relative path), otherwise
// fall back to the homepage. Only the path is echoed back, and paths a browser
// could reinterpret as external URLs ("//host", backslashes) are rejected.
func returnPath(r *http.Request) string {
	ref := r.Referer()
	u, err := url.Parse(ref)
	if err != nil {
		return "/"
	}

	sameHost := u.Host == r.Host
	relative := u.Host == "" && u.Scheme == "" && strings.HasPrefix(ref, "/") && !strings.HasPrefix(ref, "//")
	if !sameHost && !relative {
		return "/"
	}

	// Validate the decoded path so encoded variants (%5C, %2F) can't sneak an
	// external-looking location through, then redirect to the escaped form.
	if !strings.HasPrefix(u.Path, "/") || strings.HasPrefix(u.Path, "//") || strings.ContainsAny(u.Path, `\`) {
		return "/"
	}
	return u.EscapedPath()
}
