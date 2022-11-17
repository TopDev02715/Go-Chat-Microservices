package common

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
)

const (
	OAuthStateCookieName string = "oauthstate"
	SessionIdCookieName  string = "sid"
)

func GenerateStateOauthCookie(c *gin.Context, maxAge int, path, domain string) string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	c.SetCookie(OAuthStateCookieName, state, maxAge, path, domain, false, true)
	return state
}

func SetAuthCookie(c *gin.Context, sessonId string, maxAge int, path, domain string) {
	c.SetCookie(SessionIdCookieName, sessonId, maxAge, path, domain, false, true)
}

func GetCookie(c *gin.Context, name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", fmt.Errorf("get oauth google state cookie error: %w", err)
	}
	unescapedCookie, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return "", fmt.Errorf("unescape oauth google state cookie error: %w", err)
	}
	return unescapedCookie, nil
}
