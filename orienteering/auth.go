package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
)

type authHandler struct {
	next func(*gin.Context)
}

func (h *authHandler) ServeHTTP(c *gin.Context) {
	w := c.Writer
	r := c.Request
	if _, err := r.Cookie("auth"); err == http.ErrNoCookie {
		//未認証
		//fmt.Println("registar")
		w.Header().Set("Location", "login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		//何らかの別のエラーが発生
	} else {
		//成功 ラップされたハンドラーを呼び出す
		h.next(c)
	}
}
func MustAuth(fn func(*gin.Context)) *authHandler {
	AUTH := new(authHandler)
	AUTH.next = fn
	return AUTH
}
func LoginHandler(c *gin.Context) {
	r := c.Request
	w := c.Writer
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		provider, err := gomniauth.Provider(provider)
		fmt.Println("login")
		if err != nil {
			log.Fatalln("認証プロバイダーの取得に失敗しました:", provider, "-", err)
		}

		loginUrl, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			log.Fatalln("GetBeginAuthURLの呼び出し中にエラーが発生しました:", provider, "-", err)
		}

		w.Header().Set("Location", loginUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
		log.Println("TODO: ログイン処理", provider)
	case "callback":
		//log.Println("callback")
		provider, err := gomniauth.Provider(provider)
		fmt.Println("provider")
		if err != nil {

			log.Fatalln("認証プロバイダーの取得に失敗しました", provider, "-", err)
		}
		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			log.Fatalln("認証を完了できませんでした")
		}
		user, err := provider.GetUser(creds)
		if err != nil {
			log.Fatalln("ユーザーの取得に失敗しました", provider, "-", err)
		}
		m := md5.New()
		io.WriteString(m, strings.ToLower(user.Name()))
		userID := fmt.Sprintf("%x", m.Sum(nil))
		authCookieValue := objx.New(map[string]interface{}{
			"userid":     userID,
			"name":       user.Name(),
			"avatar_url": user.AvatarURL(),
			"email":      user.Email(),
		}).MustBase64()
		//log.Println(authCookieValue)
		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: authCookieValue,
			Path:  "/"})
		w.Header()["Location"] = []string{"/top"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "アクション%sには非対応です", action)

	}

}
