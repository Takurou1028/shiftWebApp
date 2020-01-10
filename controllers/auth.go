package controllers

import(
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"net/http"
)

//セッション作成（認証, 名前, 店id, ユーザーid）
func CreateSessions(c echo.Context, position int, shopID int, name string, userID int) (echo.Context, error){
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
	  Path:     "/",
	  MaxAge:   3600,
	  HttpOnly: true,
	}
	if position == 1{
		sess.Values["user_auth"] = true		
	} else {
		sess.Values["owner_auth"] = true		
	}
	sess.Values["shop_id"] = shopID
	sess.Values["user_name"] = name
	sess.Values["user_id"] = userID

	err := sess.Save(c.Request(), c.Response())
	return c, err
}

//セッション無効化
func DisableSessions(c echo.Context) (echo.Context, error){
	sess, _ := session.Get("session", c)
	//ログアウト
	sess.Values["owner_auth"] = false
	sess.Values["user_auth"] = false
	sess.Values["user_id"] = 0
	sess.Values["shop_id"] = 0
	//状態を保存
	err := sess.Save(c.Request(), c.Response())	
	return c, err
}

//オーナー認証確認し、http.Status(0が正常終了)＋ShopIDをint型で返す
func ConfirmOwnerAuth(c echo.Context) (int, int) {
	sess, err := session.Get("session", c)
	if err != nil {
		return http.StatusInternalServerError, 0
	}
	if b, _ := sess.Values["owner_auth"]; b == true {
		return http.StatusOK, sess.Values["shop_id"].(int)
	}
	return http.StatusUnauthorized , 0
}

//従業員認証確認し、http.Status(0が正常終了)＋ShopID＋Usernameを返す
func ConfirmUserAuth(c echo.Context) (int, int, string, int) {
	sess, err := session.Get("session", c)
	if err != nil {
		return http.StatusInternalServerError, 0, "", 0
	}
	if b, _ :=sess.Values["user_auth"]; b ==true {
		return http.StatusOK , sess.Values["user_id"].(int), sess.Values["user_name"].(string), sess.Values["shop_id"].(int)
	}
	return http.StatusUnauthorized , 0, "", 0
}
