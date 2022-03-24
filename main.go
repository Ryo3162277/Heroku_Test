package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"

	//_ "github.com/mattn/go-sqlite3"
	"github.com/saintfish/chardet"
	"golang.org/x/net/html/charset"

	"github.com/stretchr/gomniauth"
	//"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
)

func main() {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int { return a / b },
	}
	github_cli := os.Getenv("GITHUB_CLI")
	github_password := os.Getenv("GITHUB_PASSWORD")
	google_cli := os.Getenv("GOOGLE_CLI")
	google_password := os.Getenv("GOOGLE_PASSWORD")

	//readEvent(6535)
	router := gin.Default()
	router.SetFuncMap(funcMap)
	gomniauth.SetSecurityKey("セキュリティキー")
	gomniauth.WithProviders(
		//facebook.New("クライアントid", "秘密の値", "http://localhost:8080/auth/callback/facebook"),

		//github.New(github_cli, github_password, "http://localhost:8080/auth/callback/github"),
		github.New(github_cli, github_password, "https://fierce-citadel-16696.herokuapp.com/auth/callback/github"),
		google.New(google_cli, google_password, "https://fierce-citadel-16696.herokuapp.com/auth/callback/google"),
		//google.New(google_cli, google_password, "http://localhost:8080/auth/callback/google"),
	)
	router.LoadHTMLGlob("templates/*.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})
	/*
		router.GET("/auth/login/google", LoginHandler)
		router.GET("/auth/login/facebook", LoginHandler)
		router.GET("/auth/login/github", LoginHandler)
		router.GET("/auth/callback/google", LoginHandler)
		router.GET("/auth/callback/facebook", LoginHandler)
		router.GET("/auth/callback/github", LoginHandler)
	*/
	router.GET("/myanalysis", MustAuth(MyAnalysisHandler).ServeHTTP)
	router.GET("/everyanalysis", MustAuth(EveryAnalysisHandler).ServeHTTP)
	router.GET("/auth/:type/:provider", LoginHandler)
	router.GET("/analysis", MustAuth(AnalysisSubmitHandler).ServeHTTP)
	router.GET("/analysis/:event_id", MustAuth(EventHandler).ServeHTTP)
	router.GET("/analysis/:event_id/:class_num", MustAuth(ClassHandler).ServeHTTP)
	router.GET("/analysis/:event_id/:class_num/:index", MustAuth(RecordHandler).ServeHTTP)

	//authGroup := router.Group("/auth/*")
	//authGroup.Use(LoginHandler)
	router.GET("/top", MustAuth(func(c *gin.Context) {
		c.HTML(http.StatusOK, "top.html", gin.H{})
	}).ServeHTTP)
	router.GET("/submitted_analysis/:event_id/:class_num/:index/:no", MustAuth(SubmittedAnalysisHandler).ServeHTTP)
	router.POST("/analysis/:event_id/:class_num/:index/submitted", MustAuth(AnalysisPostHandler).ServeHTTP)
	router.POST("/submitted_analysis/:event_id/:class_num/:index/:no/change", MustAuth(AnalysisChangeHandler).ServeHTTP)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port
	router.Run(port)
	EventDBInit()
}

func EventDBInit() {
	databaseUrl := os.Getenv("DATABASE_URL")
	db, err := gorm.Open("postgres", databaseUrl)
	//db, err := gorm.Open("postgres", "event.db")
	if err != nil {
		panic("I cannot open database")
	}
	db.AutoMigrate(&Event{})

	/*
	 *
	 */
	databaseUrl1 := os.Getenv("HEROKU_POSTGRESQL_AQUA_URL")
	dbr, err1 := gorm.Open("postgres", databaseUrl1)
	//dbr, err1 := gorm.Open("postgres", "Race.db")
	if err1 != nil {
		panic("I cannot open database")
	}
	dbr.AutoMigrate(&Race{})
	/*
	 *
	 */
	databaseUrl2 := os.Getenv("HEROKU_POSTGRESQL_CYAN_URL")
	dbRec, err2 := gorm.Open("postgres", databaseUrl2)
	//dbRec, err2 := gorm.Open("postgres", "Record.db")
	if err2 != nil {
		panic("I cannot open database")
	}
	dbr.AutoMigrate(&Race{})
	dbr.Close()
	dbRec.AutoMigrate(&Record{})
	dbRec.Close()
	/*
	 *
	 */
	databaseUrl3 := os.Getenv("HEROKU_POSTGRESQL_PINK_URL")
	db_ana, err3 := gorm.Open("postgres", databaseUrl3)
	//db_ana, err3 := gorm.Open("postgres", "Analysis.db")
	if err3 != nil {
		panic("I cannot open database")
	}
	db_ana.DropTableIfExists(&Analysis{}, &Plan{}, "analysis_plans")
	db_ana.DropTableIfExists(&Analysis{}, &Execution{}, "analysis_executions")
	db_ana.DropTableIfExists(&Analysis{}, &User{}, "user_analysises")

	db_ana.AutoMigrate(&Analysis{})
	db_ana.AutoMigrate(&Plan{})
	db_ana.AutoMigrate(&Execution{})
	db_ana.AutoMigrate(&User{})
	db_ana.Close()
	url := "https://mulka2.com/lapcenter/"
	// Getリクエスト
	res, _ := http.Get(url)
	defer res.Body.Close()

	// 読み取り
	buf, _ := ioutil.ReadAll(res.Body)

	// 文字コード判定
	det := chardet.NewTextDetector()
	detRslt, _ := det.DetectBest(buf)
	//fmt.Println(detRslt.Charset)
	// => EUC-JP

	// 文字コード変換
	bReader := bytes.NewReader(buf)
	reader, _ := charset.NewReaderLabel(detRslt.Charset, bReader)

	// HTMLパース
	doc, _ := goquery.NewDocumentFromReader(reader)

	eventnum := 0
	// eventを抜き出し
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		Url, _ := s.Attr("href")
		//fmt.Println(url)
		if strings.Contains(Url, "event=") {
			strs := strings.Split(Url, "=")
			//fmt.Println(len(strs))

			var i int
			//fmt.Println(strs[1])
			i, _ = strconv.Atoi(strs[1])
			if eventnum < i {
				eventnum = i
			}

		}

	})
	fmt.Println(eventnum)
	var top Event
	er := db.Order("event_id desc").First(&top).Error
	var I int
	if er != nil {
		I = 0
	} else {
		I = top.EventID + 1
	}
	for i := I; i <= eventnum; i++ {
		//readEvent(i)
	}

	defer db.Close()
}

func dbGetAll() []Event {
	databaseUrl := os.Getenv("DATABASE_URL")
	db, err := gorm.Open("postgres", databaseUrl)
	//db, err := gorm.Open("postgres", "event.db")
	if err != nil {
		panic("データベース開けず！(dbGetAll())")
	}
	var todos []Event
	db.Order("event_id desc").Find(&todos)
	db.Close()
	return todos
}
