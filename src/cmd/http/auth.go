package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func redirectToSavePage(c *gin.Context) {
	c.Redirect(http.StatusFound, fmt.Sprintf("%s/login/error", settingsData.CLIENT_URL))
}

func redirectToClient(c *gin.Context) {
	c.Redirect(http.StatusFound, fmt.Sprintf("%s", settingsData.CLIENT_URL))
}

func initAuthProviders() {
	/*
		store := sessions.NewCookieStore(
			[]byte(settingsData.COOKIE_SESSION_SECRET),
		)
		store.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   60 * 60,
			HttpOnly: true,
			Secure:   true,
		}
		gothic.Store = store

		goth.UseProviders(
			google.New(
				settingsData.GOOGLE_AUTH_CLIENT_ID,
				settingsData.GOOGLE_AUTH_SECRET_CLIENT,
				fmt.Sprintf("%s/auth/google/callback", settingsData.BACKEND_URL),
				"profile",
				"email",
			),
		)
	*/
}
