package models

import(
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DbConnection *sql.DB

//店舗構造体
type Shop struct {
	ID			int    `form:"id"`
	OwnerName	string `form:"owner_name"`
	Password	string `form:"pass"`
}

//shoplistDB作成
func CreateShopListDB() error{
	DbConnection, _ := sql.Open("sqlite3", "DB/shoplist.sql")
	defer DbConnection.Close()
	cmd := `CREATE TABLE IF NOT EXISTS shoplist(
		id			INT,
		owner_name	STRING,
		pass		STRING);`
	_, err := DbConnection.Exec(cmd)
	return err
}

//店舗をDBに新規登録（店舗リスト）
func (shop *Shop) SignupShop() error {
	DbConnection, _ := sql.Open("sqlite3", "DB/shoplist.sql")
	defer DbConnection.Close()
	shop.GetShopID(DbConnection)	
	cmd := "INSERT INTO shoplist (id, owner_name, pass) VALUES (?, ?, ?);"
	_, err := DbConnection.Exec(cmd, shop.ID, shop.OwnerName, shop.Password)
	return err
}

//ShopId発行(
func (shop *Shop) GetShopID(DbConnection *sql.DB) error {
	cmd := "SELECT id FROM shoplist ORDER BY id DESC LIMIT 1;"
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

//idをキーにしてshoplistDBからオーナー名、パスを返す
func FindOwner(id int) (string, string, error){
	DbConnection, _ := sql.Open("sqlite3", "DB/shoplist.sql")
	defer DbConnection.Close()
	cmd := "SELECT owner_name, pass FROM shoplist WHERE id = ?;"
	row := DbConnection.QueryRow(cmd, id)
	var ownerName	string
	var password	string
	err := row.Scan(&ownerName, &password)
	return ownerName, password, err
}

//利用者構造体
type User struct{
	Position	int		`form:"position"`  //0:オーナー, 1:被雇用者
	ShopID		int		`form:"id"`
	Name		string	`form:"name"`
	Password	string	`form:"pass"`
}

//shopinfoDB作成（被雇用者リスト）
func CreateShopInfoDB(id int) error {
	DbConnection, _ := sql.Open("sqlite3", "DB/shopinfo.sql")
	defer DbConnection.Close()
	cmd := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (name STRING, pass STRING);", "shopID_"+strconv.Itoa(id))
	_, err := DbConnection.Exec(cmd)
	return err
}

//idをキーにしてshopinfoDBから被雇用者名、パスを返す/
func FindUser(id int, name string) (string, string, error){
	DbConnection, _ := sql.Open("sqlite3", "DB/shopinfo.sql")
	defer DbConnection.Close()
	cmd := fmt.Sprintf("SELECT name, pass FROM %s WHERE name = ?;", "shopID_" + strconv.Itoa(id))
	row := DbConnection.QueryRow(cmd, name)
	var userName	string
	var password	string
	err := row.Scan(&userName, &password)
	return userName, password, err
}

//従業員名の重複確認(重複あればtrue、なければfalse)
func ConfirmName(id int, name string) bool{
	DbConnection, _ := sql.Open("sqlite3", "DB/shopinfo.sql")
	defer DbConnection.Close()
	cmd := fmt.Sprintf("SELECT name FROM %s WHERE name = ?;", "shopID_" + strconv.Itoa(id))
	row := DbConnection.QueryRow(cmd, name)
	err := row.Scan(&name)
	if err != nil{
		log.Println(err)
		return false
	}
	return true
}

//shopinfoDB従業員登録
func (user *User) SignupUser() error {
	DbConnection, _ := sql.Open("sqlite3", "DB/shopinfo.sql")
	defer DbConnection.Close()
	cmd := fmt.Sprintf("INSERT INTO %s (name, pass) VALUES (?, ?);", "shopID_"+strconv.Itoa(user.ShopID))
	_, err := DbConnection.Exec(cmd, user.Name, user.Password)
	return err
}

//shopinfoDBから従業員情報削除(名前とパスの組み合わせ揃える必要あり)
//提出前のshopshiftDBからも削除
func (user *User) DeleteUser() error{
	DbConnection, _ := sql.Open("sqlite3", "DB/shopinfo.sql")
	defer DbConnection.Close()
	cmd := fmt.Sprintf("DELETE FROM %s WHERE name = ? AND pass = ?;", "shopID_"+strconv.Itoa(user.ShopID))
	if _, err := DbConnection.Exec(cmd, user.Name, user.Password); err != nil{
		return err
	}
	//削除完了したかの確認
	cmd = fmt.Sprintf("SELECT name FROM %s WHERE name = ?;", "shopID_"+strconv.Itoa(user.ShopID))
	row := DbConnection.QueryRow(cmd, user.Name)
	if err := row.Scan(user.Name); err != sql.ErrNoRows{
		return err
	}
	deleteShift(user.ShopID, user.Name)	//shopshiftDBからも削除
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

//名前とパスの組のユーザーリストをDBから返す
func GetUsers(id int) (Users, error){
	DbConnection, _ := sql.Open("sqlite3", "DB/shopinfo.sql")
	defer DbConnection.Close()
	cmd := fmt.Sprintf("SELECT * FROM %s;", "shopID_"+strconv.Itoa(id))
	rows, _ := DbConnection.Query(cmd)
	defer rows.Close()
	var uu = Users{
		List: []User{},
	}
	for rows.Next(){
		var u = User{}
		if err := rows.Scan(&u.Name, &u.Password); err != nil {
			return uu, err
		}
		uu.List = append(uu.List, u)
	}	
	if err := rows.Err(); err != nil{
		return uu, err
	}
	return uu, nil
}

//shopshiftDBに２ヶ月分のテーブル作成（シフトリスト）
func CreateShopShift(id int) error {
	DbConnection, _ := sql.Open("sqlite3", "DB/shopshift.sql")
	defer DbConnection.Close()
	year, month := GetYearandMonth()
	text := `(name STRING, day1 INT, day2 INT, day3 INT, day4 INT, day5 INT, day6 INT, day7 INT,
			 day8 INT, day9 INT, day10 INT, day11 INT, day12 INT, day13 INT, day14 INT,
			 day15 INT, day16 INT, day17 INT, day18 INT, day19 INT, day20 INT, day21 INT,
			 day22 INT, day23 INT, day24 INT, day25 INT, day26 INT, day27 INT, day28 INT`
	//30日の月, 2月, それ以外の月で場合分けして2ヶ月分テーブル作成
	for i := 0; i < 2; i++{
		tablename := strconv.Itoa(id) + strconv.Itoa(year) + strconv.Itoa(month)
		if month == 4 || month == 6 || month == 9 || month == 11 {
			cmd := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s " + text + `day29 INT, day30 INT);`, "shopID_" + tablename)
			if _, err := DbConnection.Exec(cmd); err != nil{
				return err
			}
			month++
	
		} else if month == 2 {
			//閏年の判定
			if (year % 400 == 0 || (year % 4 == 0 && year % 100 != 0)){
				cmd := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s " + text + ", day29 INT);", "shopID_" + tablename)
				if _, err := DbConnection.Exec(cmd); err != nil{
					return err
				}
				month++
			} else {
				cmd := fmt.Sprintf("REATE TABLE IF NOT EXISTS %s ` + text + `);", "shopID_" + tablename)
				if _, err := DbConnection.Exec(cmd); err != nil{
					return err
				}
				month++
			}

		} else {
			cmd := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s " + text + ", day29 INT, day30 INT, day31 INT);", "shopID_" + tablename)
			if _, err := DbConnection.Exec(cmd); err != nil{
				return err
			}
			if(month == 12){
				month = 1
				year++
			} else {
				month++
			}
		}
	}
	return nil	
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

//名前をキーにしてシフト削除
func deleteShift(id int, name string)error{
	DbConnection, _ := sql.Open("sqlite3", "DB/shopshift.sql")
	defer DbConnection.Close()
	year, month := GetNextYearandMonth()
	tablename := strconv.Itoa(id) + strconv.Itoa(year) + strconv.Itoa(month)
	cmd := fmt.Sprintf("DELETE FROM %s WHERE name = ?;", "shopID_" + tablename)
	_, err := DbConnection.Exec(cmd, name)
	return err
}

//1ヶ月先のシフト取得
func UserShift(id int, name string) (ShiftData, error){
	DbConnection, _ := sql.Open("sqlite3", "DB/shopshift.sql")
	defer DbConnection.Close()
	var shift ShiftData
	shift.Year, shift.Month = GetNextYearandMonth()

	tablename := strconv.Itoa(id) + strconv.Itoa(shift.Year) + strconv.Itoa(shift.Month)
	cmd := fmt.Sprintf("SELECT * FROM %s WHERE name = ?;", "shopID_" + tablename)
	row := DbConnection.QueryRow(cmd, name)

	if shift.Month == 4 || shift.Month == 6 || shift.Month == 9 || shift.Month ==11 {
		shift.Day = make([]DayShift, 30, 35)

		err := row.Scan(&name, &shift.Day[0].Shift, &shift.Day[1].Shift, &shift.Day[2].Shift, &shift.Day[3].Shift,
			&shift.Day[4].Shift, &shift.Day[5].Shift, &shift.Day[6].Shift, &shift.Day[7].Shift, &shift.Day[8].Shift,
			&shift.Day[9].Shift, &shift.Day[10].Shift, &shift.Day[11].Shift, &shift.Day[12].Shift, &shift.Day[13].Shift,
			&shift.Day[14].Shift, &shift.Day[15].Shift, &shift.Day[16].Shift, &shift.Day[17].Shift, &shift.Day[18].Shift,
			&shift.Day[19].Shift, &shift.Day[20].Shift, &shift.Day[21].Shift, &shift.Day[22].Shift, &shift.Day[23].Shift,
			&shift.Day[24].Shift, &shift.Day[25].Shift, &shift.Day[26].Shift, &shift.Day[27].Shift, &shift.Day[28].Shift,
			&shift.Day[29].Shift)

		for i := 1; i <= 30; i++{
			shift.Day[i-1].Day = i
		}
		if err != nil{
			return shift, err
		} 

		} else if shift.Month == 2 {
			//閏年の判定
			if (shift.Year % 400 == 0 || (shift.Year % 4 == 0 && shift.Year % 100 != 0)){
				shift.Day = make([]DayShift, 29, 35)
				err := row.Scan(&name, &shift.Day[0].Shift, &shift.Day[1].Shift, &shift.Day[2].Shift, &shift.Day[3].Shift,
					&shift.Day[4].Shift, &shift.Day[5].Shift, &shift.Day[6].Shift, &shift.Day[7].Shift, &shift.Day[8].Shift,
					&shift.Day[9].Shift, &shift.Day[10].Shift, &shift.Day[11].Shift, &shift.Day[12].Shift, &shift.Day[13].Shift,
					&shift.Day[14].Shift, &shift.Day[15].Shift, &shift.Day[16].Shift, &shift.Day[17].Shift, &shift.Day[18].Shift,
					&shift.Day[19].Shift, &shift.Day[20].Shift, &shift.Day[21].Shift, &shift.Day[22].Shift, &shift.Day[23].Shift,
					&shift.Day[24].Shift, &shift.Day[25].Shift, &shift.Day[26].Shift, &shift.Day[27].Shift, &shift.Day[28].Shift)

				for i := 1; i <= 29; i++{
					shift.Day[i-1].Day = i
				}
				if err != nil{
					return shift, err
				} 
			} else{
				shift.Day = make([]DayShift, 28, 35)
				err := row.Scan(&name, &shift.Day[0].Shift, &shift.Day[1].Shift, &shift.Day[2].Shift, &shift.Day[3].Shift,
					&shift.Day[4].Shift, &shift.Day[5].Shift, &shift.Day[6].Shift, &shift.Day[7].Shift, &shift.Day[8].Shift,
					&shift.Day[9].Shift, &shift.Day[10].Shift, &shift.Day[11].Shift, &shift.Day[12].Shift, &shift.Day[13].Shift,
					&shift.Day[14].Shift, &shift.Day[15].Shift, &shift.Day[16].Shift, &shift.Day[17].Shift, &shift.Day[18].Shift,
					&shift.Day[19].Shift, &shift.Day[20].Shift, &shift.Day[21].Shift, &shift.Day[22].Shift, &shift.Day[23].Shift,
					&shift.Day[24].Shift, &shift.Day[25].Shift, &shift.Day[26].Shift, &shift.Day[27].Shift)
				for i := 1; i <= 28; i++{
					shift.Day[i-1].Day = i
				}
				if err != nil{
					return shift, err
				} 
			}
		} else {
			shift.Day = make([]DayShift, 31, 35)
			err := row.Scan(&name, &shift.Day[0].Shift, &shift.Day[1].Shift, &shift.Day[2].Shift, &shift.Day[3].Shift,
				&shift.Day[4].Shift, &shift.Day[5].Shift, &shift.Day[6].Shift, &shift.Day[7].Shift, &shift.Day[8].Shift,
				&shift.Day[9].Shift, &shift.Day[10].Shift, &shift.Day[11].Shift, &shift.Day[12].Shift, &shift.Day[13].Shift,
				&shift.Day[14].Shift, &shift.Day[15].Shift, &shift.Day[16].Shift, &shift.Day[17].Shift, &shift.Day[18].Shift,
				&shift.Day[19].Shift, &shift.Day[20].Shift, &shift.Day[21].Shift, &shift.Day[22].Shift, &shift.Day[23].Shift,
				&shift.Day[24].Shift, &shift.Day[25].Shift, &shift.Day[26].Shift, &shift.Day[27].Shift, &shift.Day[28].Shift,
				&shift.Day[29].Shift, &shift.Day[30].Shift)
			for i := 1; i <= 31; i++{
				shift.Day[i-1].Day = i
			}
			if err != nil{
				return shift, err
			} 
		}
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
func(ms *MonthShift) RegisterMonthShift(id int, name string) error {
	DbConnection, _ := sql.Open("sqlite3", "DB/shopshift.sql")
	defer DbConnection.Close()
	year, month := GetNextYearandMonth()
	tablename := strconv.Itoa(id) + strconv.Itoa(year) + strconv.Itoa(month)
	text := `INSERT INTO %s (name, day1, day2, day3, day4, day5, day6, day7,
			day8, day9, day10, day11, day12, day13, day14,
			day15, day16, day17, day18, day19, day20, day21,
			day22, day23, day24, day25, day26, day27, day28`
	value := " VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?"
	
	if month == 4 || month == 6 || month == 9 || month ==11 {
		cmd := fmt.Sprintf(text + ", day29, day30)" + value + ", ?, ?);", "shopID_" + tablename)
		_, err := DbConnection.Exec(cmd, &name, ms.Day1, ms.Day2, ms.Day3, ms.Day4, ms.Day5, ms.Day6, ms.Day7, ms.Day8,
			ms.Day9, ms.Day10, ms.Day11, ms.Day12, ms.Day13, ms.Day14, ms.Day15, ms.Day16, ms.Day17, ms.Day18, ms.Day19,
			ms.Day20, ms.Day21, ms.Day22, ms.Day23, ms.Day24, ms.Day25, ms.Day26, ms.Day27, ms.Day28, ms.Day29, ms.Day30)
		if err != nil{
			return err
		}
	} else if month == 2 {
		//閏年の判定
		if (year % 400 == 0 || (year % 4 == 0 && year % 100 != 0)){
			cmd := fmt.Sprintf(text + ", day29)"+ value + ", ?);", "shopID_" + tablename)
			_, err := DbConnection.Exec(cmd, &name, ms.Day1, ms.Day2, ms.Day3, ms.Day4, ms.Day5, ms.Day6, ms.Day7, ms.Day8,
				ms.Day9, ms.Day10, ms.Day11, ms.Day12, ms.Day13, ms.Day14, ms.Day15, ms.Day16, ms.Day17, ms.Day18, ms.Day19,
				ms.Day20, ms.Day21, ms.Day22, ms.Day23, ms.Day24, ms.Day25, ms.Day26, ms.Day27, ms.Day28, ms.Day29)
			if err != nil{
				return err
			}
		} else {
			cmd := fmt.Sprintf(text + ");" + value, "shopID_" + tablename)
			_, err := DbConnection.Exec(cmd, &name, ms.Day1, ms.Day2, ms.Day3, ms.Day4, ms.Day5, ms.Day6, ms.Day7, ms.Day8,
				ms.Day9, ms.Day10, ms.Day11, ms.Day12, ms.Day13, ms.Day14, ms.Day15, ms.Day16, ms.Day17, ms.Day18, ms.Day19,
				ms.Day20, ms.Day21, ms.Day22, ms.Day23, ms.Day24, ms.Day25, ms.Day26, ms.Day27, ms.Day28)
			if err != nil{
				return err
			}
		}
	} else {
		cmd := fmt.Sprintf(text + ", day29, day30, day31)" + value + ", ?, ?, ?);", "shopID_" + tablename)
		_, err := DbConnection.Exec(cmd, &name, ms.Day1, ms.Day2, ms.Day3, ms.Day4, ms.Day5, ms.Day6, ms.Day7, ms.Day8,
			ms.Day9, ms.Day10, ms.Day11, ms.Day12, ms.Day13, ms.Day14, ms.Day15, ms.Day16, ms.Day17, ms.Day18, ms.Day19,
			ms.Day20, ms.Day21, ms.Day22, ms.Day23, ms.Day24, ms.Day25, ms.Day26, ms.Day27, ms.Day28, ms.Day29, ms.Day30, ms.Day31)
		if err != nil{
			return err
		}
	}
	return nil
}

//来月のシフト提出有無確認(提出されていればtrue, 提出されていなければfalse)
func ConfirmShift(id int, name string) bool{
	DbConnection, _ := sql.Open("sqlite3", "DB/shopshift.sql")
	defer DbConnection.Close()
	year, month := GetNextYearandMonth()
	tablename := strconv.Itoa(id) + strconv.Itoa(year) + strconv.Itoa(month)
	cmd := fmt.Sprintf("SELECT name FROM %s WHERE name = ?;", "shopID_" + tablename)
	row := DbConnection.QueryRow(cmd, name)
	if 	err := row.Scan(&name); err != nil{
		log.Println(err)
		return false
	}
	return true
}

//提出されたシフトアップデート(削除して挿入)
func(ms *MonthShift) UpdateMonthShift(id int, name string) error{
	DbConnection, _ := sql.Open("sqlite3", "DB/shopshift.sql")
	defer DbConnection.Close()
	year, month := GetNextYearandMonth()
	tablename := strconv.Itoa(id) + strconv.Itoa(year) + strconv.Itoa(month)
	cmd := fmt.Sprintf("DELETE FROM %s WHERE name = ?;", "shopID_" + tablename)
	if _, err := DbConnection.Exec(cmd, name); err != nil{
		return err
	}
	if err := ms.RegisterMonthShift(id, name); err != nil{
		return err
	}
	return nil
}


//全被雇用者のシフトリスト構造体
type ShiftList struct{
	Month		int
	Day			[]int
	DataList	[]ShiftData		//人ごとのデータ
}

//ユーザーの全シフト取得
func GetShiftList(id int) (ShiftList, error){
	DbConnection, _ := sql.Open("sqlite3", "DB/shopshift.sql")
	defer DbConnection.Close()
	year, month := GetNextYearandMonth()

	tablename := strconv.Itoa(id) + strconv.Itoa(year) + strconv.Itoa(month)
	cmd := fmt.Sprintf("SELECT * FROM %s ;", "shopID_" + tablename)
	rows, _ := DbConnection.Query(cmd)
	defer rows.Close()
	var sl = ShiftList{
		Month:		month,
		DataList:	[]ShiftData{},
	}

	if month == 4 || month == 6 || month == 9 || month == 11 {
		for rows.Next(){
			var shift ShiftData
			shift.Day = make([]DayShift, 30, 35)
			err := rows.Scan(&shift.Name, &shift.Day[0].Shift, &shift.Day[1].Shift, &shift.Day[2].Shift, &shift.Day[3].Shift,
				&shift.Day[4].Shift, &shift.Day[5].Shift, &shift.Day[6].Shift, &shift.Day[7].Shift, &shift.Day[8].Shift,
				&shift.Day[9].Shift, &shift.Day[10].Shift, &shift.Day[11].Shift, &shift.Day[12].Shift, &shift.Day[13].Shift,
				&shift.Day[14].Shift, &shift.Day[15].Shift, &shift.Day[16].Shift, &shift.Day[17].Shift, &shift.Day[18].Shift,
				&shift.Day[19].Shift, &shift.Day[20].Shift, &shift.Day[21].Shift, &shift.Day[22].Shift, &shift.Day[23].Shift,
				&shift.Day[24].Shift, &shift.Day[25].Shift, &shift.Day[26].Shift, &shift.Day[27].Shift, &shift.Day[28].Shift,
				&shift.Day[29].Shift)
			if err != nil {
				return sl, err
			}
			sl.DataList = append(sl.DataList, shift)
		}
		if err := rows.Err(); err != nil{
			return ShiftList{}, err
		}
		for i := 1; i <= 30; i++{
			sl.Day = append(sl.Day, i)
		}

		} else if month == 2 {
			//閏年の判定
			if (year % 400 == 0 || (year % 4 == 0 && year % 100 != 0)){
				for rows.Next(){
					var shift ShiftData
					shift.Day = make([]DayShift, 29, 35)
					err := rows.Scan(&shift.Name, &shift.Day[0].Shift, &shift.Day[1].Shift, &shift.Day[2].Shift, &shift.Day[3].Shift,
						&shift.Day[4].Shift, &shift.Day[5].Shift, &shift.Day[6].Shift, &shift.Day[7].Shift, &shift.Day[8].Shift,
						&shift.Day[9].Shift, &shift.Day[10].Shift, &shift.Day[11].Shift, &shift.Day[12].Shift, &shift.Day[13].Shift,
						&shift.Day[14].Shift, &shift.Day[15].Shift, &shift.Day[16].Shift, &shift.Day[17].Shift, &shift.Day[18].Shift,
						&shift.Day[19].Shift, &shift.Day[20].Shift, &shift.Day[21].Shift, &shift.Day[22].Shift, &shift.Day[23].Shift,
						&shift.Day[24].Shift, &shift.Day[25].Shift, &shift.Day[26].Shift, &shift.Day[27].Shift, &shift.Day[28].Shift)
					if err != nil {
						return sl, err
					}
					sl.DataList = append(sl.DataList, shift)
				}
				if err := rows.Err(); err != nil{
					return sl, err
				}
				for i := 1; i <= 29; i++{
					sl.Day = append(sl.Day, i)
				}
			} else{
				for rows.Next(){
					var shift ShiftData
					shift.Day = make([]DayShift, 28, 35)
					err := rows.Scan(&shift.Name, &shift.Day[0].Shift, &shift.Day[1].Shift, &shift.Day[2].Shift, &shift.Day[3].Shift,
						&shift.Day[4].Shift, &shift.Day[5].Shift, &shift.Day[6].Shift, &shift.Day[7].Shift, &shift.Day[8].Shift,
						&shift.Day[9].Shift, &shift.Day[10].Shift, &shift.Day[11].Shift, &shift.Day[12].Shift, &shift.Day[13].Shift,
						&shift.Day[14].Shift, &shift.Day[15].Shift, &shift.Day[16].Shift, &shift.Day[17].Shift, &shift.Day[18].Shift,
						&shift.Day[19].Shift, &shift.Day[20].Shift, &shift.Day[21].Shift, &shift.Day[22].Shift, &shift.Day[23].Shift,
						&shift.Day[24].Shift, &shift.Day[25].Shift, &shift.Day[26].Shift, &shift.Day[27].Shift)
					if err != nil {
						return sl, err
					}
					for i := 1; i <= 28; i++{
						shift.Day[i-1].Day = i
					}
					sl.DataList = append(sl.DataList, shift)
				}
				if err := rows.Err(); err != nil{
					return sl, err
				}
				for i := 1; i <= 28; i++{
					sl.Day = append(sl.Day, i)
				}
			}
		} else {
			for rows.Next(){
				var shift ShiftData
				shift.Day = make([]DayShift, 31, 35)
				err := rows.Scan(&shift.Name, &shift.Day[0].Shift, &shift.Day[1].Shift, &shift.Day[2].Shift, &shift.Day[3].Shift,
					&shift.Day[4].Shift, &shift.Day[5].Shift, &shift.Day[6].Shift, &shift.Day[7].Shift, &shift.Day[8].Shift,
					&shift.Day[9].Shift, &shift.Day[10].Shift, &shift.Day[11].Shift, &shift.Day[12].Shift, &shift.Day[13].Shift,
					&shift.Day[14].Shift, &shift.Day[15].Shift, &shift.Day[16].Shift, &shift.Day[17].Shift, &shift.Day[18].Shift,
					&shift.Day[19].Shift, &shift.Day[20].Shift, &shift.Day[21].Shift, &shift.Day[22].Shift, &shift.Day[23].Shift,
					&shift.Day[24].Shift, &shift.Day[25].Shift, &shift.Day[26].Shift, &shift.Day[27].Shift, &shift.Day[28].Shift,
					&shift.Day[29].Shift, &shift.Day[30].Shift)
				if err != nil {
					return sl, err
				}
				sl.DataList = append(sl.DataList, shift)
			}
			if err := rows.Err(); err != nil{
				return sl, err
			}
			for i := 1; i <= 31; i++{
				sl.Day = append(sl.Day, i)
			}
		}
		return sl, nil
}
