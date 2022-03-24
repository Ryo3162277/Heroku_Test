package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	//"reflect"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	//_ "github.com/jinzhu/gorm/dialects/postgres"
	//_ "github.com/mattn/go-sqlite3"
	"github.com/saintfish/chardet"
	"github.com/stretchr/objx"
	"golang.org/x/net/html/charset"
)

/*
 * 大会について
 */
type Event struct {
	gorm.Model
	EventID int
	Url     string
	Name    string
	Day     string
	Terrain string
}

/*
 * コースについて
 */
type Race struct {
	gorm.Model
	//id       int
	Class    string
	ClassNum int
	EventID  int
	URL      string
	Up       string
	Distance string
}

/*
 * 記録について
 */
type Record struct {
	gorm.Model
	EventID     int
	ClassNum    int
	Index       int
	URL         string
	RunnerName  string
	ClubName    string
	Rank        int
	Result      string
	Speed       string
	LossRate    string
	IdealTime   string
	LapTime     string
	LapRank     string
	ElapsedTime string
	ElapsedRank string
	LegLossTime string
}

func AnalysisSubmitHandler(c *gin.Context) {
	databaseUrl := os.Getenv("DATABASE_URL")
	db, err := gorm.Open("postgres", databaseUrl)
	//db, err := gorm.Open("postgres", "event.db")
	if err != nil {
		panic("I cannot open database")
	}

	db.AutoMigrate(&Event{})
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
	var top Event
	er := db.Order("event_id desc").First(&top).Error
	var I int
	if er != nil {
		I = 0
	} else {
		I = top.EventID + 1
	}
	for i := I; i <= eventnum; i++ {
		readEvent(i)
	}

	events := dbGetAll()
	db.Close()
	c.HTML(http.StatusOK, "analysis.html", gin.H{"event": events})
}
func readEvent(i int) {
	eventURL := "https://mulka2.com/lapcenter/lapcombat2/index.jsp?event=" + strconv.Itoa(i) + "&file=1"
	res, err := http.Get(eventURL)
	if err != nil {
		return
	}
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
	var eventname string
	doc.Find("title").Each(func(_ int, s *goquery.Selection) {
		eventname = s.Text()
	})
	if eventname == "Lap Center (オリエンテーリング大会のリザルト・ラップ解析）" {
		return
	}
	eventday := "eventday"
	terrain := "terrain"
	doc.Find("div.container").Each(func(_ int, s *goquery.Selection) {
		s.Find("div").Each(func(__ int, s1 *goquery.Selection) {
			//fmt.Println(s1.Html())
			s1.Find("span").Each(func(__ int, s2 *goquery.Selection) {

				Str, exi := s2.Attr("style")
				if exi && Str == "padding-right: 10px;" {
					//fmt.Println(Str)
					//fmt.Println(s2.Text())
					if eventday == "eventday" {
						eventday = s2.Text()
					} else if terrain == "terrain" {
						terrain = s2.Text()
					}
				}
			})
		})

		//fmt.Println(s.Text())

	})

	analysisURL := "/analysis/" + strconv.Itoa(i)

	//*events = append(*events, *event)
	databaseUrl := os.Getenv("DATABASE_URL")
	db, err := gorm.Open("postgres", databaseUrl)
	//db, err := gorm.Open("postgres", "event.db")
	if err != nil {
		panic("failed to connect database\n")
	}
	databaseUrl1 := os.Getenv("HEROKU_POSTGRESQL_AQUA_URL")
	db1, err1 := gorm.Open("postgres", databaseUrl1)
	//db1, err1 := gorm.Open("postgres", "Race.db")
	if err1 != nil {
		panic("failed to connect database\n")
	}
	db.Create(&Event{EventID: i, Url: analysisURL, Name: eventname, Terrain: terrain, Day: eventday})
	//fmt.Println(event)
	db.Close()
	Num := 0
	doc.Find("div.row").Each(func(_ int, s *goquery.Selection) {
		s.Find("div").Each(func(_ int, s1 *goquery.Selection) {

			s1.Find("table").Each(func(_ int, s2 *goquery.Selection) {

				s2.Find("tr").Each(func(_ int, s3 *goquery.Selection) {
					//fmt.Print("s3")
					//fmt.Println(s3.Text())
					race := "race"
					dis := "dis"
					up := "up"
					s4 := s3.Find("td")
					race = s4.Find("b").Text()
					s4.Find("span").Each(func(_ int, s5 *goquery.Selection) {
						if dis == "dis" {
							dis = s5.Text()
						} else if up == "up" {
							up = s5.Text()
						}
					})

					//fmt.Print(s4.Find("b").First().Text())
					if up != "up" && dis != "dis" {
						classurl := analysisURL + "/" + strconv.Itoa(Num)
						db1.Create(&Race{URL: classurl, Class: race, Distance: dis, Up: up, EventID: i, ClassNum: Num})

						readRecord(i, Num)
						Num += 1
					}

				})

			})
		})
	})

	db1.Close()
}

func readRecord(eventid int, classnum int) {
	URL := "https://mulka2.com/lapcenter/lapcombat2/split-list.jsp?event=" + strconv.Itoa(eventid) + "&file=1&class=" + strconv.Itoa(classnum) + "&content=analysis"
	res, err := http.Get(URL)
	if err != nil {
		return
	}
	defer res.Body.Close()

	databaseUrl2 := os.Getenv("HEROKU_POSTGRESQL_CYAN_URL")
	db, err2 := gorm.Open("postgres", databaseUrl2)

	//db, err2 := gorm.Open("postgres", "Record.db")
	if err2 != nil {
		panic("I cannot open database")
	}
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
	fmt.Println("EventID:" + strconv.Itoa(eventid) + " , class:" + strconv.Itoa(classnum))
	index := -1
	runnerName := ""
	clubName := ""
	rank := -1
	result := ""
	speed := ""
	lossRate := ""
	idealTime := ""
	lapTime := ""
	lapRank := ""
	elapsedTime := ""
	elapsedRank := ""
	legLossTime := ""
	eventID := -1
	classNum := -1
	url := ""

	URL = "/analysis/" + strconv.Itoa(eventid) + "/" + strconv.Itoa(classnum)

	eventID = eventid
	classNum = classnum

	BOOLTMP := false
	doc.Find("script").Each(func(_ int, s1 *goquery.Selection) {
		html, err := s1.Html()

		if err != nil {
			panic("error")

		}
		lines := strings.Split(html, "\n")

		for _, line := range lines {
			if strings.Contains(line, "runnerData[") {
				line = strings.Replace(line, "&#39;", "", -1)
				line = strings.Replace(line, " ", "", -1)
				if strings.Contains(line, "index") {
					line = strings.Replace(line, ";", "", -1)
					index, _ = strconv.Atoi(strings.Split(line, "=")[1])
				} else if strings.Contains(line, "[runnerName]") {
					line = strings.Replace(line, ";", "", -1)
					runnerName = strings.Split(line, "=")[1]
				} else if strings.Contains(line, "clubName") {
					line = strings.Replace(line, ";", "", -1)
					clubName = strings.Split(line, "=")[1]
				} else if strings.Contains(line, "[rank]") {
					line = strings.Replace(line, ";", "", -1)
					rank, _ = strconv.Atoi(strings.Split(line, "=")[1])
				} else if strings.Contains(line, "result") {
					result = strings.Split(line, "=")[1]
				} else if strings.Contains(line, "speed") {
					speed = strings.Split(line, "=")[1]
				} else if strings.Contains(line, "lossRate") {
					lossRate = strings.Split(line, "=")[1]
				} else if strings.Contains(line, "idealTime") {
					idealTime = strings.Split(line, "=")[1]
				} else if strings.Contains(line, "lapTime") {
					lT := strings.Split(line, "=")[1]
					lT = strings.Replace(lT, "[", "", -1)
					lT = strings.Replace(lT, "]", "", -1)
					lapTime = lT
				} else if strings.Contains(line, "lapRank") {
					lR := strings.Split(line, "=")[1]
					lR = strings.Replace(lR, "[", "", -1)
					lR = strings.Replace(lR, "]", "", -1)
					lapRank = lR
				} else if strings.Contains(line, "elapsedTime") {
					eT := strings.Split(line, "=")[1]
					eT = strings.Replace(eT, "[", "", -1)
					eT = strings.Replace(eT, "]", "", -1)
					elapsedTime = eT
				} else if strings.Contains(line, "elapsedRank") {
					eR := strings.Split(line, "=")[1]
					eR = strings.Replace(eR, "[", "", -1)
					eR = strings.Replace(eR, "]", "", -1)
					elapsedRank = eR
				} else if strings.Contains(line, "legLossTime") {
					lLT := strings.Split(line, "=")[1]
					lLT = strings.Replace(lLT, "[", "", -1)
					lLT = strings.Replace(lLT, "]", "", -1)
					legLossTime = lLT
					BOOLTMP = true
				}
				if BOOLTMP {
					url = URL + "/" + strconv.Itoa(index)
					db.Create(&Record{Index: index, RunnerName: runnerName, ClubName: clubName, Rank: rank, Result: result, Speed: speed, LossRate: lossRate, IdealTime: idealTime, LapTime: lapTime, LapRank: lapRank, ElapsedTime: elapsedTime, ElapsedRank: elapsedRank, LegLossTime: legLossTime, EventID: eventID, ClassNum: classNum, URL: url})
					BOOLTMP = false

				}
				//fmt.Println(line)
			}
		}
		//fmt.Println(html)

	})
	db.Close()
}

func EventHandler(c *gin.Context) {
	r := c.Request
	//w := c.Writer
	segs := strings.Split(r.URL.Path, "/")
	id := segs[2]
	databaseUrl1 := os.Getenv("HEROKU_POSTGRESQL_AQUA_URL")
	db, err := gorm.Open("postgres", databaseUrl1)

	//db, err := gorm.Open("postgres", "Race.db")
	if err != nil {
		panic("failed to connect database\n")
	}
	racearray := []Race{}
	db.Where("event_id = ?", id).Find(&racearray)
	c.HTML(http.StatusOK, "event.html", gin.H{"race": racearray})
	db.Close()
}
func ClassHandler(c *gin.Context) {
	r := c.Request
	segs := strings.Split(r.URL.Path, "/")
	eventid := segs[2]
	classid := segs[3]
	databaseUrl2 := os.Getenv("HEROKU_POSTGRESQL_CYAN_URL")
	db, err := gorm.Open("postgres", databaseUrl2)

	//db, err := gorm.Open("postgres", "Record.db")
	if err != nil {
		panic("failed to connect database")
	}
	recordarray := []Record{}
	db.Where("event_id = ? AND class_num = ?", eventid, classid).Find(&recordarray)
	c.HTML(http.StatusOK, "class.html", gin.H{"records": recordarray})
	db.Close()
}
func RecordHandler(c *gin.Context) {
	r := c.Request
	url := r.URL.Path
	databaseUrl2 := os.Getenv("HEROKU_POSTGRESQL_CYAN_URL")
	db, err := gorm.Open("postgres", databaseUrl2)

	//db, err := gorm.Open("postgres", "Record.db")
	if err != nil {
		panic("failed to connect database")
	}
	var rec Record
	db.Where("url = ?", url).First(&rec)
	lT := strings.Replace(rec.LapTime, ";", "", -1)
	lTs := strings.Split(lT, ",")
	lR := strings.Replace(rec.LapRank, ";", "", -1)
	lRs := strings.Split(lR, ",")

	eT := strings.Replace(rec.ElapsedTime, ";", "", -1)
	eTs := strings.Split(eT, ",")
	eR := strings.Replace(rec.ElapsedRank, ";", "", -1)
	eRs := strings.Split(eR, ",")

	lLT := strings.Replace(rec.LegLossTime, ";", "", -1)
	lLTs := strings.Split(lLT, ",")
	var event Event
	databaseUrl := os.Getenv("DATABASE_URL")
	dbe, err1 := gorm.Open("postgres", databaseUrl)

	//dbe, err1 := gorm.Open("postgres", "event.db")
	if err1 != nil {
		panic("failed to connect database")
	}
	dbe.Where("event_id = ?", rec.EventID).First(&event)
	var race Race
	databaseUrl1 := os.Getenv("HEROKU_POSTGRESQL_AQUA_URL")
	dbR, err2 := gorm.Open("postgres", databaseUrl1)

	//dbR, err2 := gorm.Open("postgres", "Race.db")
	if err2 != nil {
		panic("failed to connect database")
	}
	dbR.Where("event_id = ? AND class_num = ?", rec.EventID, rec.ClassNum).First(&race)

	c.HTML(http.StatusOK, "record.html", gin.H{"record": rec, "lT": lTs, "lR": lRs, "eT": eTs, "eR": eRs, "lLT": lLTs, "race": race, "event": event})
	db.Close()
	dbe.Close()
	dbR.Close()
}

type Analysis struct {
	gorm.Model
	EventID     int
	EventName   string
	ClassNum    int
	ClassName   string
	RunnerIndex int
	RunnerName  string
	No          int
	URL         string
	Plans       []Plan      `gorm:"many2many:analysis_plans;"`
	Executions  []Execution `gorm:"many2many:analysis_executions;"`
	//Comments    []Comment   `gorm:"many2many:analysis_comments;"`
	FavoriteNum int
	Plan        string
	Execution   string
	UserID      string
	UserName    string
}

type Plan struct {
	gorm.Model
	Plan string
}
type Execution struct {
	gorm.Model
	UserID    string
	UserName  string
	Execution string
	//CommentNo int
}

/*
 * User Model
 */
type User struct {
	gorm.Model
	User     string
	Analysis []Analysis `gorm:"many2many:user_analysises;"`
}

func AnalysisPostHandler(c *gin.Context) {
	//c.Request.ParseForm()
	r := c.Request
	//w :=c.Writer
	segs := strings.Split(r.URL.Path, "/")
	//fmt.Println(c.PostForm("plan1"))
	databaseUrl2 := os.Getenv("HEROKU_POSTGRESQL_CYAN_URL")
	db, err := gorm.Open("postgres", databaseUrl2)

	//db, err := gorm.Open("postgres", "Record.db")
	if err != nil {
		panic("failed to connect database")
	}
	var rec Record
	//fmt.Println(segs[4])
	EID, _ := strconv.Atoi(segs[2])
	CNM, _ := strconv.Atoi(segs[3])
	IND, _ := strconv.Atoi(segs[4])
	temp_url := "/analysis/" + segs[2] + "/" + segs[3] + "/" + segs[4]
	db.Where("url = ?", temp_url).First(&rec)
	//db.Where("event_id = ?", EID).Where("class_num = ?",CNM).Where(" 'index' = ?",IND).First(&rec)
	lT := strings.Replace(rec.LapTime, ";", "", -1)
	//fmt.Println(rec.RunnerName)
	lTs := strings.Split(lT, ",")
	//fmt.Println(lTs)

	runner := rec.RunnerName
	//fmt.Println(runner)
	//fmt.Println("runner")
	databaseUrl1 := os.Getenv("HEROKU_POSTGRESQL_AQUA_URL")
	dbRace, err := gorm.Open("postgres", databaseUrl1)
	//dbRace, err := gorm.Open("postgres", "Race.db")
	if err != nil {
		panic("failed to connect database")
	}
	var race Race
	dbRace.Where("event_id = ?", EID).Where("class_num = ?", CNM).First(&race)
	class_name := race.Class
	databaseUrl := os.Getenv("DATABASE_URL")
	dbEvent, err := gorm.Open("postgres", databaseUrl)

	//dbEvent, err := gorm.Open("postgres", "Event.db")
	if err != nil {
		panic("failed to connect database")
	}
	var event Event
	dbEvent.Where("event_id = ?", EID).First(&event)
	event_name := event.Name
	plan := c.PostForm("plan")
	execution := c.PostForm("execution")
	if authCookie, err := r.Cookie("auth"); err == nil {
		k := objx.MustFromBase64(authCookie.Value)
		//fmt.Println(k["userid"])
		//fmt.Println(k)
		databaseUrl3 := os.Getenv("HEROKU_POSTGRESQL_PINK_URL")
		db_ana, err1 := gorm.Open("postgres", databaseUrl3)

		//db_ana, err1 := gorm.Open("postgres", "Analysis.db")
		if err1 != nil {
			panic("failed to connect database")
		}

		user_id := k["userid"]
		user_name := k["name"]

		if User_ID, ok := user_id.(string); ok {
			if User_Name, ok := user_name.(string); ok {
				analysisarray := []Analysis{}

				db_ana.Where("event_id = ?", EID).Where("class_num = ?", CNM).Where("runner_index = ?", IND).Find(&analysisarray)
				size_ana := len(analysisarray)
				url := "/submitted_analysis/" + segs[2] + "/" + segs[3] + "/" + segs[4] + "/" + strconv.Itoa(size_ana)
				ana := &Analysis{EventID: EID, ClassNum: CNM, RunnerIndex: IND, Plan: plan, Execution: execution, URL: url, No: size_ana, UserID: User_ID, UserName: User_Name, EventName: event_name, ClassName: class_name, RunnerName: runner, FavoriteNum: 0}
				db_ana.Create(ana)
				for i := 0; i < len(lTs); i++ {
					p := &Plan{Plan: c.PostForm("plan" + strconv.Itoa(i+1))}
					e := &Execution{Execution: c.PostForm("execution" + strconv.Itoa(i+1))}
					db_ana.Create(p)
					db_ana.Create(e)
					db_ana.Model(ana).Association("Plans").Append(p)
					db_ana.Model(ana).Association("Executions").Append(e)
				}
			}
		}

		c.Redirect(http.StatusSeeOther, "/top")
		db_ana.Close()
	} else {

		c.Redirect(http.StatusSeeOther, "/")
	}
	db.Close()
	dbEvent.Close()
	dbRace.Close()
}

func MyAnalysisHandler(c *gin.Context) {
	r := c.Request
	databaseUrl3 := os.Getenv("HEROKU_POSTGRESQL_PINK_URL")
	db, err := gorm.Open("postgres", databaseUrl3)

	//db, err := gorm.Open("postgres", "Analysis.db")
	if err != nil {
		panic("failed to connect database")
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		k := objx.MustFromBase64(authCookie.Value)

		user_id := k["userid"]
		//user_name:=k["name"]
		if User_ID, ok := user_id.(string); ok {
			var AnalysisArray []Analysis
			db.Where("user_id = ?", User_ID).Find(&AnalysisArray)
			c.HTML(http.StatusOK, "analysislist.html", gin.H{"analysisarray": AnalysisArray})
		}

		//c.Redirect(http.StatusSeeOther,"/myanalysis")

	} else {

		c.Redirect(http.StatusSeeOther, "/")
	}
	db.Close()
}

func EveryAnalysisHandler(c *gin.Context) {
	databaseUrl3 := os.Getenv("HEROKU_POSTGRESQL_PINK_URL")
	db, err := gorm.Open("postgres", databaseUrl3)

	//db, err := gorm.Open("postgres", "Analysis.db")
	if err != nil {
		panic("failed to connect database")
	}

	var AnalysisArray []Analysis
	db.Find(&AnalysisArray)
	c.HTML(http.StatusOK, "analysislist.html", gin.H{"analysisarray": AnalysisArray})
	db.Close()
}

func SubmittedAnalysisHandler(c *gin.Context) {
	r := c.Request
	url := r.URL.Path
	databaseUrl3 := os.Getenv("HEROKU_POSTGRESQL_PINK_URL")
	db, err := gorm.Open("postgres", databaseUrl3)

	//db, err := gorm.Open("postgres", "Analysis.db")
	if err != nil {
		panic("failed to connect database")
	}

	if authCookie, err := r.Cookie("auth"); err == nil {
		k := objx.MustFromBase64(authCookie.Value)

		user_id := k["userid"]
		if User_ID, ok := user_id.(string); ok {
			var analysis Analysis
			db.Where("url = ?", url).First(&analysis)
			db.First(&analysis, analysis.ID).Related(&analysis.Plans, "Plans").Related(&analysis.Executions, "Executions")
			//var tmp_ana Analysis
			//db.First(&tmp_ana,analysis.ID)
			databaseUrl2 := os.Getenv("HEROKU_POSTGRESQL_CYAN_URL")
			db1, err1 := gorm.Open("postgres", databaseUrl2)

			//db1, err1 := gorm.Open("postgres", "Record.db")
			if err1 != nil {
				panic("failed to connect database")
			}
			var rec Record
			//fmt.Println(segs[4])
			EID := analysis.EventID
			CNM := analysis.ClassNum
			IND := analysis.RunnerIndex
			temp_url := "/analysis/" + strconv.Itoa(EID) + "/" + strconv.Itoa(CNM) + "/" + strconv.Itoa(IND)
			db1.Where("url = ?", temp_url).First(&rec)

			lT := strings.Replace(rec.LapTime, ";", "", -1)
			lTs := strings.Split(lT, ",")
			lR := strings.Replace(rec.LapRank, ";", "", -1)
			lRs := strings.Split(lR, ",")

			eT := strings.Replace(rec.ElapsedTime, ";", "", -1)
			eTs := strings.Split(eT, ",")
			eR := strings.Replace(rec.ElapsedRank, ";", "", -1)
			eRs := strings.Split(eR, ",")

			lLT := strings.Replace(rec.LegLossTime, ";", "", -1)
			lLTs := strings.Split(lLT, ",")

			var event Event
			databaseUrl := os.Getenv("DATABASE_URL")
			dbe, err1 := gorm.Open("postgres", databaseUrl)

			//dbe, err1 := gorm.Open("postgres", "event.db")
			if err1 != nil {
				panic("failed to connect database")
			}
			dbe.Where("event_id = ?", rec.EventID).First(&event)
			var race Race

			databaseUrl1 := os.Getenv("HEROKU_POSTGRESQL_AQUA_URL")
			dbR, err2 := gorm.Open("postgres", databaseUrl1)
			//dbR, err2 := gorm.Open("postgres", "Race.db")
			if err2 != nil {
				panic("failed to connect database")
			}
			dbR.Where("event_id = ? AND class_num = ?", rec.EventID, rec.ClassNum).First(&race)

			if analysis.UserID == User_ID {
				c.HTML(http.StatusOK, "myanalysis.html", gin.H{"analysis": analysis, "record": rec, "lT": lTs, "lR": lRs, "eT": eTs, "eR": eRs, "lLT": lLTs, "race": race, "event": event})
			} else {
				c.HTML(http.StatusOK, "viewanalysis.html", gin.H{"analysis": analysis, "record": rec, "lT": lTs, "lR": lRs, "eT": eTs, "eR": eRs, "lLT": lLTs, "race": race, "event": event})
			}

			db1.Close()
			dbR.Close()
			dbe.Close()
			//c.HTML(http.StatusOK, "myanalysis.html", gin.H{"analysisarray":AnalysisArray})
		}

		//c.Redirect(http.StatusSeeOther,"/myanalysis")

	} else {

		c.Redirect(http.StatusSeeOther, "/")
	}
	db.Close()
}

func AnalysisChangeHandler(c *gin.Context) {
	//c.Request.ParseForm()
	r := c.Request
	//w :=c.Writer
	segs := strings.Split(r.URL.Path, "/")

	databaseUrl2 := os.Getenv("HEROKU_POSTGRESQL_CYAN_URL")
	db, err := gorm.Open("postgres", databaseUrl2)
	//db, err := gorm.Open("postgres", "Record.db")

	if err != nil {
		panic("failed to connect database")
	}
	var rec Record

	temp_url := "/submitted_analysis/" + segs[2] + "/" + segs[3] + "/" + segs[4]
	db.Where("url = ?", temp_url).First(&rec)

	db.Close()
	plan := c.PostForm("plan")
	execution := c.PostForm("execution")

	databaseUrl3 := os.Getenv("HEROKU_POSTGRESQL_PINK_URL")
	db_ana, err1 := gorm.Open("postgres", databaseUrl3)
	//db_ana, err1 := gorm.Open("postgres", "Analysis.db")

	if err1 != nil {
		panic("failed to connect database")
	}

	var analysis Analysis
	temp_url = temp_url + "/" + segs[5]
	db_ana.Where("url = ?", temp_url).First(&analysis)
	analysis.Plan = plan
	analysis.Execution = execution
	db_ana.Save(&analysis)
	db_ana.First(&analysis, analysis.ID).Related(&analysis.Plans, "Plans").Related(&analysis.Executions, "Executions")

	for i, p := range analysis.Plans {
		//fmt.Println(p.Plan)
		//new_p:=&Plan{Plan:c.PostForm("plan"+strconv.Itoa(i+1))}
		p.Plan = c.PostForm("plan" + strconv.Itoa(i+1))
		//fmt.Println(p.Plan)

		db_ana.Save(&p)

	}
	for i, e := range analysis.Executions {
		e.Execution = c.PostForm("execution" + strconv.Itoa(i+1))
		db_ana.Save(&e)

	}

	db_ana.Close()
	c.Redirect(http.StatusSeeOther, "/top")

}
