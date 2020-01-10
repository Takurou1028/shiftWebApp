package models

import(
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DbConnection *sql.DB

//店舗構造体
type Shop struct {
	ID			int    `form:"shopid"`
	OwnerName	string `form:"owner_name"`
	Password	string `form:"pass"`
}

//shoplistDB作成
func CreateOwnerDB() error{
	DbConnection, _ := sql.Open("sqlite3", "DB/owner.sql")
	defer DbConnection.Close()
	cmd := `CREATE TABLE IF NOT EXISTS owner(
		shopid			INT		PRIMARY KEY,
		owner_name	STRING,
		pass		STRING);`
	_, err := DbConnection.Exec(cmd)
	return err
}

//店舗をDBに新規登録（店舗リスト）
func (shop *Shop) SignupShop() error {
	DbConnection, _ := sql.Open("sqlite3", "DB/owner.sql")
	defer DbConnection.Close()
	shop.GetShopID(DbConnection)	
	cmd := "INSERT INTO owner (shopid, owner_name, pass) VALUES (?, ?, ?);"
	_, err := DbConnection.Exec(cmd, shop.ID, shop.OwnerName, shop.Password)
	return err
}

//ShopId発行(
func (shop *Shop) GetShopID(DbConnection *sql.DB) error {
	cmd := "SELECT shopid FROM owner ORDER BY shopid DESC LIMIT 1;"
	row := DbConnection.QueryRow(cmd)
	var dummyID int
	err := row.Scan(&dummyID)
	if err == sql.ErrNoRows {
		shop.ID = 1
	}
	if err == nil{
		shop.ID = dummyID + 1
	}
	return err
}

//shopidをキーにしてownerDBからオーナー名、パスを返す
func FindOwner(shopID int) (string, string, error){
	DbConnection, _ := sql.Open("sqlite3", "DB/owner.sql")
	defer DbConnection.Close()
	cmd := "SELECT owner_name, pass FROM owner WHERE shopid = ?;"
	row := DbConnection.QueryRow(cmd, shopID)
	var ownerName	string
	var password	string
	err := row.Scan(&ownerName, &password)
	return ownerName, password, err
}

//利用者構造体
type User struct{
	Position	int		`form:"position"`  //0:オーナー, 1:被雇用者
	ShopID		int		`form:"shopid"`
	Name		string	`form:"name"`
	Password	string	`form:"pass"`
	ID			int		`form:"userid"`
}

//userlistDB作成（被雇用者リスト）
func CreateUserListDB() error {
	DbConnection, _ := sql.Open("sqlite3", "DB/userlist.sql")
	defer DbConnection.Close()
	cmd := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
			(userid INT PRIMARY KEY, name STRING, pass STRING, shopid INT);`,
			 "userlist")
	_, err := DbConnection.Exec(cmd)
	return err
}

//UserId発行(
func (user *User) GetUserID(DbConnection *sql.DB) error {
	cmd := "SELECT userid FROM userlist ORDER BY userid DESC LIMIT 1;"
	row := DbConnection.QueryRow(cmd)
	var dummyID int
	err := row.Scan(&dummyID)
	if err == sql.ErrNoRows {
		user.ID = 1
	}
	if err == nil{
		user.ID = dummyID + 1
	}
	return err
}

//useridをキーにしてshopinfoDBから被雇用者名、パスを返す/
func FindUser(userID int) (string, string, error){
	DbConnection, _ := sql.Open("sqlite3", "DB/userlist.sql")
	defer DbConnection.Close()
	cmd := fmt.Sprintf("SELECT name, pass FROM %s WHERE userid = ?;", "userlist")
	row := DbConnection.QueryRow(cmd, userID)
	var userName	string
	var password	string
	err := row.Scan(&userName, &password)
	return userName, password, err
}

//shopinfoDB従業員登録
func (user *User) SignupUser() error {
	DbConnection, _ := sql.Open("sqlite3", "DB/userlist.sql")
	defer DbConnection.Close()
	user.GetUserID(DbConnection)
	cmd := fmt.Sprintf("INSERT INTO %s (userid, name, pass, shopid) VALUES (?, ?, ?, ?);", "userlist")
	_, err := DbConnection.Exec(cmd, user.ID, user.Name, user.Password, user.ShopID)
	return err
}

//shopinfoDBから従業員情報削除(名前とパスの組み合わせ揃える必要あり)
//提出前のshopshiftDBからも削除
func (user *User) DeleteUser() error{
	DbConnection, _ := sql.Open("sqlite3", "DB/userlist.sql")
	defer DbConnection.Close()
	cmd := "SELECT userid FROM userlist WHERE userid = ?;"
	row := DbConnection.QueryRow(cmd, user.ID)
	if err := row.Scan(user.Name); err != nil{
		return err
	}
	cmd = "DELETE FROM userlist WHERE userid = ? AND name = ? AND pass = ?;"
	if _, err := DbConnection.Exec(cmd, user.ID, user.Name, user.Password); err != nil{
		return err
	}
	//削除完了したかの確認
	cmd = "SELECT userid FROM userlist WHERE userid = ?;"
	row = DbConnection.QueryRow(cmd, user.ID)
	if err := row.Scan(user.Name); err != sql.ErrNoRows{
		return err
	}
	deleteShift(user.ShopID)	//shopshiftDBからも削除
	return nil
}

//Ownerのユーザーリスト確認用
type Users struct{
	List []User
}

//現在の年と月を返す
func GetYearandMonth()(int, int){
	t := time.Now()
	var year, month int
	year = int(t.Year())
	month = int(t.Month())
	return year, month
}
//来月の年と月を返す
func GetNextYearandMonth()(int, int){
	var year, month int
	year, month = GetYearandMonth()
	if month == 12 {
		month = 1
		year++
	} else{
		month++
	}
	return year, month
}

//ユーザーIDと名前とパスの組のユーザーリストをDBから返す
func GetUsers(shopID int) (Users, error){
	DbConnection, _ := sql.Open("sqlite3", "DB/userlist.sql")
	defer DbConnection.Close()
	cmd := fmt.Sprintf("SELECT userid, name, pass FROM %s WHERE shopid = ?;", "userlist")
	rows, _ := DbConnection.Query(cmd, shopID)
	defer rows.Close()
	var uu = Users{
		List: []User{},
	}
	for rows.Next(){
		var u = User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Password); err != nil {
			return uu, err
		}
		uu.List = append(uu.List, u)
	}	
	if err := rows.Err(); err != nil{
		return uu, err
	}
	return uu, nil
}


//shiftDB作成
func CreateShiftDB() error {
	DbConnection, _ := sql.Open("sqlite3", "DB/shift.sql")
	defer DbConnection.Close()
	text := `CREATE TABLE IF NOT EXISTS %s 
			(shiftid INT PRIMARY KEY, userid INT, shopid INT, year INT, month INT,
			day1 INT, day2 INT, day3 INT, day4 INT, day5 INT, day6 INT, day7 INT,
			day8 INT, day9 INT, day10 INT, day11 INT, day12 INT, day13 INT, day14 INT,
			day15 INT, day16 INT, day17 INT, day18 INT, day19 INT, day20 INT, day21 INT,
			day22 INT, day23 INT, day24 INT, day25 INT, day26 INT, day27 INT, day28 INT,
			day29 INT, day30 INT, day31 INT)`
	cmd := fmt.Sprintf(text, "shift")
	if _, err := DbConnection.Exec(cmd); err != nil{
		return err
	}
	return nil	
}

func GetShiftID(DbConnection *sql.DB) (int, error) {
	cmd := "SELECT shiftid FROM shift ORDER BY shiftid DESC LIMIT 1;"
	row := DbConnection.QueryRow(cmd)
	var dummyID int
	var shiftID int
	err := row.Scan(&dummyID)
	if err == sql.ErrNoRows {
		shiftID = 1
	}
	if err == nil{
		shiftID = dummyID + 1
	}
	return shiftID, err
}


//被雇用者のシフトデータ構造体
type ShiftData struct{
	Month	int
	Year	int
	Day		[]DayShift
	Name	string
}

type DayShift struct{
	Shift	int
	Day		int
}

//userIDをキーにしてシフト削除
func deleteShift(userID int)error{
	DbConnection, _ := sql.Open("sqlite3", "DB/shift.sql")
	defer DbConnection.Close()
	cmd := fmt.Sprintf("DELETE FROM %s WHERE userid = ?;", "shift")
	_, err := DbConnection.Exec(cmd, userID)
	return err
}


//1ヶ月先のシフト取得
func UserShift(userID int, name string) (ShiftData, error){
	DbConnection, _ := sql.Open("sqlite3", "DB/shift.sql")
	defer DbConnection.Close()
	year, month := GetNextYearandMonth()

	cmd := "SELECT * FROM shift WHERE userid = ? AND year = ? AND month = ?;"
	row := DbConnection.QueryRow(cmd, userID, year, month)

	var shift = ShiftData{}
	shift.Day = make([]DayShift, 27, 31)
	dummySpace := make([]int, 5, 7)
	var dummy1, dummy2, dummy3 int
	err := row.Scan(&dummySpace[0], &dummySpace[1], &dummySpace[2], &shift.Year, &shift.Month,
		&shift.Day[0].Shift, &shift.Day[1].Shift, &shift.Day[2].Shift, &shift.Day[3].Shift,
		&shift.Day[4].Shift, &shift.Day[5].Shift, &shift.Day[6].Shift, &shift.Day[7].Shift, &shift.Day[8].Shift,
		&shift.Day[9].Shift, &shift.Day[10].Shift, &shift.Day[11].Shift, &shift.Day[12].Shift, &shift.Day[13].Shift,
		&shift.Day[14].Shift, &shift.Day[15].Shift, &shift.Day[16].Shift, &shift.Day[17].Shift, &shift.Day[18].Shift,
		&shift.Day[19].Shift, &shift.Day[20].Shift, &shift.Day[21].Shift, &shift.Day[22].Shift, &shift.Day[23].Shift,
		&shift.Day[24].Shift, &shift.Day[25].Shift, &shift.Day[26].Shift, &shift.Day[27].Shift, &dummy1, &dummy2, &dummy3)
		if err != nil {
			return shift, err
		}
		for i := 1; i <= 28; i++{
			shift.Day[i-1].Day = i
		}
		if month == 4 || month == 6 || month == 9 || month == 11 {
			shift.Day = append(shift.Day, DayShift{Shift:dummy1, Day: 29,})
			shift.Day = append(shift.Day, DayShift{Shift:dummy2, Day: 30,})

		} else if month == 2 {
			if (year % 400 == 0 || (year % 4 == 0 && year % 100 != 0)){
				shift.Day = append(shift.Day, DayShift{Shift:dummy1, Day: 29,})
			}
		} else {
			shift.Day = append(shift.Day, DayShift{Shift:dummy1, Day: 29,})
			shift.Day = append(shift.Day, DayShift{Shift:dummy2, Day: 30,})
			shift.Day = append(shift.Day, DayShift{Shift:dummy3, Day: 31,})
		}
		shift.Name = name
	return shift, nil
}


//被雇用者から入力されたのシフト格納用構造体
type MonthShift struct{
	Day1	int	`form:"day1"`
	Day2	int	`form:"day2"`
	Day3	int	`form:"day3"`
	Day4	int	`form:"day4"`
	Day5	int	`form:"day5"`
	Day6	int	`form:"day6"`
	Day7	int	`form:"day7"`
	Day8	int	`form:"day8"`
	Day9	int	`form:"day9"`
	Day10	int	`form:"day10"`
	Day11	int	`form:"day11"`
	Day12	int	`form:"day12"`
	Day13	int	`form:"day13"`
	Day14	int	`form:"day14"`
	Day15	int	`form:"day15"`
	Day16	int	`form:"day16"`
	Day17	int	`form:"day17"`
	Day18	int	`form:"day18"`
	Day19	int	`form:"day19"`
	Day20	int	`form:"day20"`
	Day21	int	`form:"day21"`
	Day22	int	`form:"day22"`
	Day23	int	`form:"day23"`
	Day24	int	`form:"day24"`
	Day25	int	`form:"day25"`
	Day26	int	`form:"day26"`
	Day27	int	`form:"day27"`
	Day28	int	`form:"day28"`
	Day29	int	`form:"day29"`
	Day30	int	`form:"day30"`
	Day31	int	`form:"day31"`
}

//提出されたシフト登録
func(ms *MonthShift) RegisterMonthShift(userID int, name string, shopID int) error {
	DbConnection, _ := sql.Open("sqlite3", "DB/shift.sql")
	defer DbConnection.Close()
	year, month := GetNextYearandMonth()
	shiftID, _ := GetShiftID(DbConnection)
	text := `INSERT INTO %s (shiftid, userid, shopid, year, month,
			day1, day2, day3, day4, day5, day6, day7,
			day8, day9, day10, day11, day12, day13, day14,
			day15, day16, day17, day18, day19, day20, day21,
			day22, day23, day24, day25, day26, day27, day28`
	value := ` VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
					 ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?`
	
	if month == 4 || month == 6 || month == 9 || month ==11 {
		cmd := fmt.Sprintf(text + ", day29, day30)" + value + ", ?, ?);", "shift")
		_, err := DbConnection.Exec(cmd, shiftID, userID, shopID, year, month, ms.Day1, ms.Day2, ms.Day3, ms.Day4, ms.Day5, ms.Day6, ms.Day7, ms.Day8,
			ms.Day9, ms.Day10, ms.Day11, ms.Day12, ms.Day13, ms.Day14, ms.Day15, ms.Day16, ms.Day17, ms.Day18, ms.Day19,
			ms.Day20, ms.Day21, ms.Day22, ms.Day23, ms.Day24, ms.Day25, ms.Day26, ms.Day27, ms.Day28, ms.Day29, ms.Day30)
		if err != nil{
			return err
		}
	} else if month == 2 {
		//閏年の判定
		if (year % 400 == 0 || (year % 4 == 0 && year % 100 != 0)){
			cmd := fmt.Sprintf(text + ", day29)"+ value + ", ?);", "shift")
			_, err := DbConnection.Exec(cmd, shiftID, userID, shopID, year, month, ms.Day1, ms.Day2, ms.Day3, ms.Day4, ms.Day5, ms.Day6, ms.Day7, ms.Day8,
				ms.Day9, ms.Day10, ms.Day11, ms.Day12, ms.Day13, ms.Day14, ms.Day15, ms.Day16, ms.Day17, ms.Day18, ms.Day19,
				ms.Day20, ms.Day21, ms.Day22, ms.Day23, ms.Day24, ms.Day25, ms.Day26, ms.Day27, ms.Day28, ms.Day29)
			if err != nil{
				log.Println(err)
				return err
			}
		} else {
			cmd := fmt.Sprintf(text + ");" + value, "shift")
			_, err := DbConnection.Exec(cmd, shiftID, userID, shopID, year, month, ms.Day1, ms.Day2, ms.Day3, ms.Day4, ms.Day5, ms.Day6, ms.Day7, ms.Day8,
				ms.Day9, ms.Day10, ms.Day11, ms.Day12, ms.Day13, ms.Day14, ms.Day15, ms.Day16, ms.Day17, ms.Day18, ms.Day19,
				ms.Day20, ms.Day21, ms.Day22, ms.Day23, ms.Day24, ms.Day25, ms.Day26, ms.Day27, ms.Day28)
			if err != nil{
				return err
			}
		}
	} else {
		cmd := fmt.Sprintf(text + ", day29, day30, day31)" + value + ", ?, ?, ?);", "shift")
		_, err := DbConnection.Exec(cmd, shiftID, userID, shopID, year, month, ms.Day1, ms.Day2, ms.Day3, ms.Day4, ms.Day5, ms.Day6, ms.Day7, ms.Day8,
			ms.Day9, ms.Day10, ms.Day11, ms.Day12, ms.Day13, ms.Day14, ms.Day15, ms.Day16, ms.Day17, ms.Day18, ms.Day19,
			ms.Day20, ms.Day21, ms.Day22, ms.Day23, ms.Day24, ms.Day25, ms.Day26, ms.Day27, ms.Day28, ms.Day29, ms.Day30, ms.Day31)
		if err != nil{
			return err
		}
	}
	return nil
}

//来月のシフト提出有無確認(提出されていればtrue, 提出されていなければfalse)
func ConfirmShift(userID int, name string) bool{
	DbConnection, _ := sql.Open("sqlite3", "DB/shift.sql")
	defer DbConnection.Close()
	year, month := GetNextYearandMonth()
	cmd := fmt.Sprintf("SELECT shiftid FROM %s WHERE userid = ? AND year = ? AND month = ?;", "shift")
	row := DbConnection.QueryRow(cmd, userID, year, month)
	var shiftID int
	if 	err := row.Scan(&shiftID); err != nil{
		log.Println(err)
		return false
	}
	return true
}

//提出されたシフトアップデート(削除して挿入)
func(ms *MonthShift) UpdateMonthShift(userID int, name string, shopID int) error{
	DbConnection, _ := sql.Open("sqlite3", "DB/shift.sql")
	defer DbConnection.Close()
	year, month := GetNextYearandMonth()
	cmd := fmt.Sprintf("DELETE FROM %s WHERE userid = ? AND year = ? AND month = ?;", "shift")
	if _, err := DbConnection.Exec(cmd, userID, year, month); err != nil{
		return err
	}
	if err := ms.RegisterMonthShift(userID, name, shopID); err != nil{
		return err
	}
	return nil
}


//シフト構造体
type Shift struct{
	ShiftID		int
	UserID		int
	ShopID		int
	Year		int
	Month		int
	Day			[]int	//シフト
	Name		string
}

//シフトリスト構造体
type ShiftList struct{
	Shifts		[]Shift
	Day			[]int
	Month		int
}

//ユーザーの全シフト取得
func GetShiftList(shopID int) (ShiftList, error){
	DbConnection, _ := sql.Open("sqlite3", "DB/shift.sql")
	defer DbConnection.Close()
	year, month := GetNextYearandMonth()

	cmd := fmt.Sprintf("SELECT * FROM %s WHERE shopid = ? AND year = ? AND month = ?;", "shift")
	rows, _ := DbConnection.Query(cmd, shopID, year, month)
	defer rows.Close()
	var sl = ShiftList{}
	sl.Month = month

	for rows.Next(){
		var shift Shift
		shift.Day = make([]int, 27, 31)
		var dummy1, dummy2, dummy3 int
		err := rows.Scan(&shift.ShiftID, &shift.UserID, &shift.ShopID, &shift.Year, &shift.Month,
			&shift.Day[0], &shift.Day[1], &shift.Day[2], &shift.Day[3],&shift.Day[4], &shift.Day[5],
			&shift.Day[6], &shift.Day[7], &shift.Day[8], &shift.Day[9], &shift.Day[10], &shift.Day[11],
			&shift.Day[12], &shift.Day[13],&shift.Day[14], &shift.Day[15], &shift.Day[16], &shift.Day[17],
			&shift.Day[18], &shift.Day[19], &shift.Day[20], &shift.Day[21], &shift.Day[22], &shift.Day[23],
			&shift.Day[24], &shift.Day[25], &shift.Day[26], &shift.Day[27], &dummy1, &dummy2, &dummy3)
		if err != nil {
			return sl, err
		}
		if month == 4 || month == 6 || month == 9 || month == 11 {
			shift.Day = append(shift.Day, dummy1)
			shift.Day = append(shift.Day, dummy2)
		} else if month == 2 {
			if (year % 400 == 0 || (year % 4 == 0 && year % 100 != 0)){
				shift.Day = append(shift.Day, dummy1)
			}
		} else {
			shift.Day = append(shift.Day, dummy1)
			shift.Day = append(shift.Day, dummy2)
			shift.Day = append(shift.Day, dummy3)
		}
		sl.Shifts = append(sl.Shifts, shift)
	}
	for i := range sl.Shifts[0].Day{
		sl.Day = append(sl.Day, i + 1)
	}
	if err := rows.Err(); err != nil{
		return sl, err
	}
	return sl, nil
}
