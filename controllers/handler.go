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
	
	if err := models.CreateShopListDB(); err != nil {
		return c.String(http.StatusInternalServerError, "Failure to create shoplist Table")
	}

	if err := newShop.SignupShop(); err != nil {
		return c.String(http.StatusInternalServerError, "Failure to sign up")
	}

	if err := models.CreateShopInfoDB(newShop.ID); err != nil {
		return c.String(http.StatusInternalServerError, "Failure to create shopinfo Table")
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
	c, err := CreateSessions(c, user.Position, user.ShopID, user.Name)	//セッション作成
	if err != nil{
		return c.NoContent(http.StatusInternalServerError)
	}

	if err := models.CreateShopShift(user.ShopID); err != nil {
		return c.String(http.StatusInternalServerError, "Failure to create shift table")
	}
	return c.Redirect(http.StatusSeeOther, "/owner")
	}
	
	//従業員の場合の処理
	if user.Position == 1{
		name, pass, err := models.FindUser(user.ShopID, user.Name)
		if err != nil{
			log.Println(err)	
		}	
		if user.Name != name || user.Password != pass {
			return c.String(http.StatusUnauthorized, "invalid id, name or password")
		}
	
		c, err := CreateSessions(c, user.Position, user.ShopID, user.Name)	//セッション作成
		if err != nil{
			return c.NoContent(http.StatusInternalServerError)
		}
	
		models.CreateShopShift(user.ShopID)
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
	status, id := ConfirmOwnerAuth(c)	//session確認
	if status != http.StatusOK {
		return c.String(status, "Authorization error")
	}

	u := new(models.User)	//id以外の入力されたユーザー情報をバインド
	if err := c.Bind(u); err != nil {
		return err
	}
	if u.Name == "" || u.Password == "" {
		return c.String(http.StatusBadRequest, "invalid name or password")
	}

	u.ShopID = id
	if models.ConfirmName(u.ShopID, u.Name){
		return c.String(http.StatusBadRequest, "User name has already used")
	}
	//ユーザー登録
	if err := u.SignupUser(); err != nil{
		return c.String(http.StatusInternalServerError, "Failure to signup User")
	}

	monthShift := new(models.MonthShift)
	if err := monthShift.RegisterMonthShift(u.ShopID, u.Name); err != nil{
		return c.String(http.StatusInternalServerError, "Failure to insert data") 
	}
	
	return CreateUserList(c)
}

func CreateUserList(c echo.Context) error{
	status, id := ConfirmOwnerAuth(c)	//session確認
	if status != http.StatusOK{
		return c.String(status, "Authorization error")
	}

	users, err := models.GetUsers(id)
	if err != nil{
		log.Println(err)
	}
	return c.Render(http.StatusOK, "registration.html", users)
}

func DeleteUser(c echo.Context) error{
	status, id := ConfirmOwnerAuth(c)	//session確認
	if status != http.StatusOK{
		return c.String(status, "Authorization error")
	}

	u := new(models.User)
	if err := c.Bind(u); err != nil {
		return err
	}
	//名前もしくはパスがない場合、エラーを返す
	if u.Name == "" || u.Password == "" {
		return c.String(http.StatusBadRequest, "Delete failed. Confirm name and pass")
	}
	u.ShopID = id
	err := u.DeleteUser()	//ユーザー情報削除
	if err != nil{
		return c.String(http.StatusBadRequest, "Delete failed. Confirm name and pass")
	}

	users, err:= models.GetUsers(id)
	if err != nil{
		log.Println(err)
	}
	return c.Render(http.StatusOK, "registration.html", users)
}

func CreateShiftList(c echo.Context) error{
	status, id := ConfirmOwnerAuth(c)	//session確認
	if status != http.StatusOK{
		return c.String(status, "Authorization error")
	}

	shiftlist, err := models.GetShiftList(id)
	if err != nil{
		return err
	}

	return c.Render(http.StatusOK, "shiftlist.html", shiftlist)

}

//以下は/user
func ShowUserPage(c echo.Context) error{
	status, id, name := ConfirmUserAuth(c)	//session確認
	if status != http.StatusOK{
		return c.String(status, "Authorization error")
	}

	shift, err := models.UserShift(id, name)	//DBからshift取得
	if err != nil{
		log.Println(err)
	}
	return c.Render(http.StatusOK, "user.html", shift)
}

func SubmitShift(c echo.Context) error{
	status, id, name := ConfirmUserAuth(c)	//session確認
	if status != http.StatusOK{
		return c.String(status, "Authorization error")
	}

	monthShift := new(models.MonthShift)	//提出されたシフトバインド
	if err := c.Bind(monthShift); err != nil {
		return err
	}
	if models.ConfirmShift(id, name){	//シフト提出有無の確認
		if err := monthShift.UpdateMonthShift(id, name); err != nil{	//提出されていたらアップデート
			return c.String(http.StatusInternalServerError, "Incomplete update")
		}
	} else {
		monthShift.RegisterMonthShift(id, name)
	}

	return ShowUserPage(c)	//ページ整形
}





