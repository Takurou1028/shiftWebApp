# shiftWebApp
シフト集計用アプリ

## 環境
Language: go version go1.11.12 darwin/amd64  
WAF: echo  
DB: SQLite version 3.24.0  

## 設計
MVC

## DB設計(現状は別DB）
| ownerlist |　userlist |userlist |
|:---:|:---:|:---:|
| 店舗ID | ユーザ-ID | シフトID |
| オーナー名 | 店舗ID  | ユーザーID|
| パスワード | ユーザー名 | 店舗ID |
|  |パスワード | 年 |
|  |  | 月 |
|  |  | 日（１〜３１日） |


## 仕様
1. 店舗登録  
2. 雇用者ログイン  
3. 被雇用者ログイン  
4. 被雇用者登録  
5. 被雇用者削除  
6. 被雇用者シフト提出  
7. 被雇用者シフト更新  
8. 提出されたシフト確認  

##  今後実装したいこと
1.提出されたシフトに基づく、シフト作成
2.完成したシフトの公開
3.過去のシフトの閲覧
4.一定期間経過後の過去のシフト削除
5.各操作をログに記録
