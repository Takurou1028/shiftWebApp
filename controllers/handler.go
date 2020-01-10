package controllers

import(
	"github.com/labstack/echo"
	"log"
	"net/http"
	"shift-webapp/models"
)

func MethodOverride(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        if c.Request().Method == "POST" {
            method := c.Request().PostFormValue("_method")
            if method == "PUT" || method == "DELETE" {
                c.Request().Method = method
            }
        }
        return next(c)
    }
}

//店舗登録
func Signup(c echo.Context) error{
	newShop := new(models.Shop)
	if err := c.Bind(newShop); err != nil {
		return err
	}
	//名前もしくはパスがない場合、エラーを返す
	if newShop.OwnerName == "" || newShop.Password == "" {
		return c.String(http.StatusBadRequest, "Invalid name or password")
	}
	
	if err := models.CreateOwnerDB(); err != nil {
		return c.String(http.StatusInternalServerError, "Failure to create owner table")
	}

	if err := newShop.SignupShop(); err != nil {
		return c.String(http.StatusInternalServerError, "Failure to sign up")
	}

	if err := models.CreateUserListDB(); err != nil {
		return c.String(http.StatusInternalServerError, "Failure to create userlist table")
	}

	if err := models.CreateShiftDB(); err != nil {
		return c.String(http.StatusInternalServerError, "Failure to create shift table")
	}
	
	newShop.Password = ""
	return c.Render(http.StatusOK, "finishing_signup.html", newShop)
}

//ログイン
func Login(c echo.Context) error{
	user := new(models.User)
	if err := c.Bind(user); err != nil {
		return err
	}

	//オーナーの場合
	if user.Position == 0{
	name, pass, err := models.FindOwner(user.ShopID)
	if err != nil{
		log.Println(err)	
	}	
	if user.Name != name || user.Password != pass {
		return c.String(http.StatusUnauthorized, "Invalid id, name or password")
	}
	c, err := CreateSessions(c, user.Position, user.ShopID, user.Name, user.ID)	//セッション作成
	if err != nil{
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.Redirect(http.StatusSeeOther, "/owner")
	}
	
	//従業員の場合の処理
	if user.Position == 1{
		name, pass, err := models.FindUser(user.ID)
		user.Name = name
		if err != nil{
			log.Println(err)	
		}	
		if user.Password != pass {
			return c.String(http.StatusUnauthorized, "Invalid id, name or password")
		}
	
		c, err := CreateSessions(c, user.Position, user.ShopID, user.Name, user.ID)	//セッション作成
		if err != nil{
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.Redirect(http.StatusSeeOther, "/user")
	}
	return c.String(http.StatusBadRequest, "BadRequest")
}

//ログアウト
func Logout(c echo.Context) error{
	c, err := DisableSessions(c)
	if err != nil{
		return c.NoContent(http.StatusInternalServerError)		
	}
	return c.Render(http.StatusOK, "logout.html", nil)
}

//以下は/owner
func ShowOwnerPage(c echo.Context) error{
	status, _ := ConfirmOwnerAuth(c)	//session確認
	if status != http.StatusOK{
		return c.String(status, "Authorization error")
	}
	return c.Render(http.StatusOK, "owner.html", nil)
}

//被雇用者を新規登録
func RegisterUser(c echo.Context) error{
	status, shopID := ConfirmOwnerAuth(c)	//session確認
	if status != http.StatusOK {
		return c.String(status, "Authorization error")
	}

	u := new(models.User)	//shopID, userID以外の入力されたユーザー情報をバインド
	if err := c.Bind(u); err != nil {
		return err
	}
	u.ShopID = shopID
	if u.Name == "" || u.Password == "" {
		return c.String(http.StatusBadRequest, "invalid name or password")
	}

	//ユーザー登録
	if err := u.SignupUser(); err != nil{	//IDも入る
		return c.String(http.StatusInternalServerError, "Failure to signup User")
	}

	monthShift := new(models.MonthShift)
	if err := monthShift.RegisterMonthShift(u.ID, u.Name, u.ShopID); err != nil{
		return c.String(http.StatusInternalServerError, "Failure to insert data") 
	}
	
	return CreateUserList(c)
}

func CreateUserList(c echo.Context) error{
	status, shopID := ConfirmOwnerAuth(c)	//session確認
	if status != http.StatusOK{
		return c.String(status, "Authorization error")
	}

	users, err := models.GetUsers(shopID)
	if err != nil{
		log.Println(err)
	}
	return c.Render(http.StatusOK, "registration.html", users)
}

func DeleteUser(c echo.Context) error{
	status, shopID := ConfirmOwnerAuth(c)	//session確認
	if status != http.StatusOK{
		return c.String(status, "Authorization error")
	}

	u := new(models.User)
	if err := c.Bind(u); err != nil {
		return err
	}
	u.ShopID = shopID
	//パスがないもしくはユーザーidが正しくない場合、エラーを返す
	if u.ID == 0 || u.Name =="" || u.Password == "" {
		return c.String(http.StatusBadRequest, "Delete failed. Confirm name, pass and user_id")
	}
	err := u.DeleteUser()	//ユーザー情報削除
	if err != nil{
		return c.String(http.StatusBadRequest, "Delete failed. Confirm name, pass and user_id")
	}

	users, err:= models.GetUsers(shopID)
	if err != nil{
		log.Println(err)
	}
	return c.Render(http.StatusOK, "registration.html", users)
}

func CreateShiftList(c echo.Context) error{
	status, shopID := ConfirmOwnerAuth(c)	//session確認
	if status != http.StatusOK{
		return c.String(status, "Authorization error")
	}

	shiftlist, err := models.GetShiftList(shopID)
	if err != nil{
		return err
	}

	return c.Render(http.StatusOK, "shiftlist.html", shiftlist)

}

//以下は/user
func ShowUserPage(c echo.Context) error{
	status, userID, name, _ := ConfirmUserAuth(c)	//session確認
	if status != http.StatusOK{
		return c.String(status, "Authorization error")
	}

	shift, err := models.UserShift(userID, name)	//DBからshift取得
	if err != nil{
		log.Println(err)
	}
	return c.Render(http.StatusOK, "user.html", shift)
}

func SubmitShift(c echo.Context) error{
	status, userID, name, shopID := ConfirmUserAuth(c)	//session確認
	if status != http.StatusOK{
		return c.String(status, "Authorization error")
	}

	monthShift := new(models.MonthShift)	//提出されたシフトバインド
	if err := c.Bind(monthShift); err != nil {
		return err
	}
	if models.ConfirmShift(userID, name){	//シフト提出有無の確認
		if err := monthShift.UpdateMonthShift(userID, name, shopID); err != nil{	//提出されていたらアップデート
			return c.String(http.StatusInternalServerError, "Incomplete update")
		} else {
		monthShift.RegisterMonthShift(userID, name, shopID)
		}
	}
	return ShowUserPage(c)	//ページ整形
}