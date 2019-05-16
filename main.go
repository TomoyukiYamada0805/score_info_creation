package main

import (
  "fmt"
  "github.com/PuerkitoBio/goquery"
  "strings"
  "strconv"
  "regexp"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "time"
  "log"
)

func main() {
  // :=演算子は初期値を宣言する
  doc, err := goquery.NewDocument("https://soccer.yahoo.co.jp/jleague/league/j1")
  //doc, err := goquery.NewDocument("https://soccer.yahoo.co.jp/jleague/schedule/j1/10/all")
  if err != nil {
      fmt.Printf("Failed")
  }

  //db, err := sql.Open("mysql", "root:@/soccor_scoring_development")
  ////db, err := sql.Open("mysql", "be3622c3dba887:72d4f74d@tcp(us-cdbr-iron-east-02.cleardb.net:3306)/heroku_02e85bd01702801?")
  //if err != nil{
  //	  fmt.Printf("Failed")
  //}
  err = Db.Connection()
  if err != nil {
      log.Fatal(err)
  }
	defer Db.Close()

  doc.Find("#modSoccerSchedule02 > .modBody > .partsTable > table > tbody > tr").Each(func(i int, s *goquery.Selection) {
    err = Db.Transaction(func(db *sql.Tx) error {
      if s.Find(".score > .status").Text() == "試合終了" {
        matchUrl, exists := s.Find(".score > a").Attr("href")
        fmt.Println(matchUrl)
        matchUrlSplit    := strings.Split(matchUrl, "/")
        matchId          := matchUrlSplit[len(matchUrlSplit)-1]
        fmt.Println("GameID: " + matchId)
  
        rows, err := db.Query("SELECT count(*) FROM matches WHERE match_id="+matchId)
        if err != nil {
            log.Fatal(err)
        }
  
        columns, err := rows.Columns() // カラム名を取得
        if err != nil {
          panic(err.Error())
        }
            values := make([]sql.RawBytes, len(columns))
      
        //  rows.Scan は引数に `[]interface{}`が必要.
      
        scanArgs := make([]interface{}, len(values))
        for i := range values {
          scanArgs[i] = &values[i]
        }
  
        var value string
      
        for rows.Next() {
          err = rows.Scan(scanArgs...)
          if err != nil {
            panic(err.Error())
          }
      
          for i, col := range values {
            // Here we can check if the value is nil (NULL value)
            if col == nil {
              value = "NULL"
            } else {
              value = string(col)
            }
            fmt.Println(columns[i], ": ", value)
          }
          
        }
  
        if exists == false {
         fmt.Printf("get url failed")
         
        }else if value == "0" {
  
          doc, err := goquery.NewDocument("https://soccer.yahoo.co.jp/"+matchUrl)
  
          if err != nil {
              fmt.Printf("Failed")
          }
          
          t := time.Now()
	        const layout = "2006-01-02 15:04:05"
          current_time := t.Format(layout)
          
          matchSummary := doc.Find(".gameSummaryHead > .head > .title").Text()
          r := regexp.MustCompile(`\w+`)
          section := r.FindAllString(matchSummary, 2)[1]
          fmt.Println(section)
          
          matchTime := doc.Find(".gameSummaryHead > .body > .note > .time > dd").Text()
          fmt.Println(matchTime)
          
          homeTeamName     := doc.Find(".homeTeam > .name > a").Text()
          homeTeamUrl, _   := doc.Find(".homeTeam > .name > a").Attr("href")
          homeTeamUrlSplit := strings.Split(homeTeamUrl, "/")
          homeTeamId       := homeTeamUrlSplit[len(homeTeamUrlSplit)-1]
        
          awayTeamName     := doc.Find(".awayTeam > .name > a").Text()
          awayTeamUrl, _   := doc.Find(".awayTeam > .name > a").Attr("href")
          awayTeamUrlSplit := strings.Split(awayTeamUrl, "/")
          awayTeamId       := awayTeamUrlSplit[len(awayTeamUrlSplit)-1]
          fmt.Println(homeTeamName + "vs" + awayTeamName)
          fmt.Println("HomeTeamId" + homeTeamId)
          fmt.Println("AwayTeamId" + awayTeamId)
          
          homeScore := doc.Find(".home.goal").Text()
          awayScore := doc.Find(".away.goal").Text()
          homeFirstScore := doc.Find(".home.first").Text()
          awayFirstScore := doc.Find(".away.first").Text()
          homeSecondScore := doc.Find(".home.second").Text()
          awaySecondScore := doc.Find(".away.second").Text()
          fmt.Println("前半 " + homeFirstScore + ":" + awayFirstScore)
          fmt.Println("後半 " + homeSecondScore + ":" + awaySecondScore)
          fmt.Println(homeScore + ":" + awayScore)
          
          fmt.Println("Match Progress")
          doc.Find("#gam_stat").Find("tr").Each(func(i int, s *goquery.Selection) {
            
            homeGoal      := s.Find(".home > .goal").Text()
            homeYellow    := s.Find(".home > .yellow").Text()
            homeYellowTwo := s.Find(".home > .yellowTwo").Text()
            homeRed       := s.Find(".home > .red").Text()
            homeChange    := s.Find(".home > .change").Text()
            awayGoal      := s.Find(".away > .goal").Text()
            awayYellow    := s.Find(".away > .yellow").Text()
            awayYellowTwo := s.Find(".away > .yellowTwo").Text()
            awayRed       := s.Find(".away > .red").Text()
            awayChange    := s.Find(".away > .change").Text()
          
            if homeGoal != "" {
              homeGoalPlayer := s.Find(".home > a").Text()
              // 得点付き
              goalTimePoint  := s.Find(".time").Text()
              goalTime := strings.Split(goalTimePoint, "分")[0]
              fmt.Println(goalTime)
              fmt.Println(homeTeamName+":"+homeGoalPlayer)
              // match_progresses(home)
              insert, err := db.Query("INSERT INTO match_progresses(match_id, team_type, progress_type, progress_time, scorer, yellow_card, red_card, from_change_player, to_change_player, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, 0, 0, goalTime, homeGoalPlayer, "",  "", "", "", current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          
            if homeYellow != "" {
              yellowCardPlayer := s.Find(".home > a").Text()
              yellowCardTimeBeforeSplit   :=  s.Find(".time").Text()
              yellowCardTime := strings.Split(yellowCardTimeBeforeSplit, "分")[0]
              fmt.Println(yellowCardTime)
              fmt.Println("イエローカード "+homeTeamName+":"+yellowCardPlayer)
              // match_progresses(home)
              insert, err := db.Query("INSERT INTO match_progresses(match_id, team_type, progress_type, progress_time, scorer, yellow_card, red_card, from_change_player, to_change_player, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, 0, 1, yellowCardTime, "", yellowCardPlayer, "", "", "", current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          
            if homeYellowTwo != "" {
              yellowCardPlayer  := s.Find(".home > a").Text()
              yellowCardTwoTimeBeforeSplit :=  s.Find(".time").Text()
              yellowCardTwoTime := strings.Split(yellowCardTwoTimeBeforeSplit, "分")[0]
              fmt.Println(yellowCardTwoTime)
              fmt.Println("イエローカード2枚目 "+homeTeamName+":"+yellowCardPlayer)
              // match_progresses(home)
              insert, err := db.Query("INSERT INTO match_progresses(match_id, team_type, progress_type, progress_time, scorer, yellow_card, red_card, from_change_player, to_change_player, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, 0, 2, yellowCardTwoTime, "", "", yellowCardPlayer, "", "", current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          
            if homeRed != "" {
              redCardPlayer := s.Find(".home > a").Text()
              redCardTimeBeforeSplit   :=  s.Find(".time").Text()
              redCardTime := strings.Split(redCardTimeBeforeSplit, "分")[0]
              fmt.Println(redCardTime)
              fmt.Println("レッド "+homeTeamName+":"+redCardPlayer)
              // match_progresses(home)
              insert, err := db.Query("INSERT INTO match_progresses(match_id, team_type, progress_type, progress_time, scorer, yellow_card, red_card, from_change_player, to_change_player, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, 0, 3, redCardTime, "", "", redCardPlayer, "", "", current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          
            if homeChange != "" {
              fromChangePlayer := s.Find(".home > a").Eq(0).Text()
              toChangePlayer   := s.Find(".home > a").Eq(1).Text()
              changeTimePoint  :=  s.Find(".time").Text()
              changeTime := strings.Split(changeTimePoint, "分")[0]
              fmt.Println(changeTime)
              fmt.Println(homeTeamName+":"+fromChangePlayer + " -> " + toChangePlayer)
              // match_progresses(home)
              insert, err := db.Query("INSERT INTO match_progresses(match_id, team_type, progress_type, progress_time, scorer, yellow_card, red_card, from_change_player, to_change_player, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, 0, 4, changeTime, "", "", "", fromChangePlayer, toChangePlayer, current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          
            if awayGoal != "" {
              awayGoalPlayer := s.Find(".away > a").Text()
              goalTimePoint  := s.Find(".time").Text()
              goalTime := strings.Split(goalTimePoint, "分")[0]
              fmt.Println(goalTime)
              fmt.Println(awayTeamName+":"+awayGoalPlayer)
              // match_progresses(home)
              insert, err := db.Query("INSERT INTO match_progresses(match_id, team_type, progress_type, progress_time, scorer, yellow_card, red_card, from_change_player, to_change_player, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, 1, 0, goalTime, awayGoalPlayer, "", "", "", "", current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          
            if awayYellow != "" {
              yellowCardPlayer := s.Find(".away > a").Text()
              yellowCardTimeBeforeSplit :=  s.Find(".time").Text()
              yellowCardTime := strings.Split(yellowCardTimeBeforeSplit, "分")[0]
              fmt.Println("イエローカード "+awayTeamName+":"+yellowCardPlayer)
              // match_progresses(home)
              insert, err := db.Query("INSERT INTO match_progresses(match_id, team_type, progress_type, progress_time, scorer, yellow_card, red_card, from_change_player, to_change_player, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, 1, 1, yellowCardTime, "", yellowCardPlayer, "", "", "", current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          
            if awayYellowTwo != "" {
              yellowCardPlayer := s.Find(".away > a").Text()
              yellowCardTwoTimeBeforeSplit :=  s.Find(".time").Text()
              yellowCardTwoTime := strings.Split(yellowCardTwoTimeBeforeSplit, "分")[0]
              fmt.Println("イエローカード2枚目 "+awayTeamName+":"+yellowCardPlayer)
              // match_progresses(home)
              insert, err := db.Query("INSERT INTO match_progresses(match_id, team_type, progress_type, progress_time, scorer, yellow_card, red_card, from_change_player, to_change_player, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, 1, 2, yellowCardTwoTime, "", "", yellowCardPlayer, "", "", current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          
            if awayRed != "" {
              redCardPlayer := s.Find(".away > a").Text()
              redCardTimeBeforeSplit   :=  s.Find(".time").Text()
              redCardTime := strings.Split(redCardTimeBeforeSplit, "分")[0]
              fmt.Println("レッド "+awayTeamName+":"+redCardPlayer)
              // match_progresses(home)
              insert, err := db.Query("INSERT INTO match_progresses(match_id, team_type, progress_type, progress_time, scorer, yellow_card, red_card, from_change_player, to_change_player, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, 1, 3, redCardTime, "", "", redCardPlayer, "", "", current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          
            if awayChange != "" {
              fromChangePlayer := s.Find(".away > a").Eq(0).Text()
              toChangePlayer   := s.Find(".away > a").Eq(1).Text()
              changeTimePoint  :=  s.Find(".time").Text()
              changeTime := strings.Split(changeTimePoint, "分")[0]
              fmt.Println(awayTeamName+":"+fromChangePlayer + " -> " + toChangePlayer)
              // match_progresses(home)
              insert, err := db.Query("INSERT INTO match_progresses(match_id, team_type, progress_type, progress_time, scorer, yellow_card, red_card, from_change_player, to_change_player, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, 1, 4, changeTime, "", "", "", fromChangePlayer, toChangePlayer, current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          })
        
          var homeControleRate    int = 0
          var awayControleRate    int = 0
          var homePassCount       int = 0
          var awayPassCount       int = 0
          var homePassSuccessRate int = 0
          var awayPassSuccessRate int = 0
          var homeMileage         int = 0
          var awayMileage         int = 0
          var homeShootCount      int = 0
          var awayShootCount      int = 0
          var homeFrameCount      int = 0
          var awayFrameCount      int = 0
          var homeSprintCount     int = 0
          var awaySprintCount     int = 0
          var homeOffside         int = 0
          var awayOffside         int = 0
          var homeFreekickCount   int = 0
          var awayFreekickCount   int = 0
          var homeCornerkickCount int = 0
          var awayCornerkickCount int = 0
          var homePenaltyKick     int = 0
          var awayPenaltyKick     int = 0
        
          // チームスタッツ
          doc.Find(".gameSummaryBody table").Eq(1).Find("tr").Each(func(i int, s *goquery.Selection) {
            if s.Find(".time").Text() == "ボール支配率" {
              homeControleRateString :=  r.FindAllString(s.Find(".home").Text(), 1)[0]
              homeControleRate, _ = strconv.Atoi(homeControleRateString)
              awayControleRateString := r.FindAllString(s.Find(".away").Text(), 1)[0]
              awayControleRate, _ =  strconv.Atoi(awayControleRateString)
              fmt.Println(s.Find(".time").Text() + ": " + homeTeamName + " " + r.FindAllString(s.Find(".home").Text(), 1)[0] + " " + awayTeamName + r.FindAllString(s.Find(".away").Text(), 1)[0])
            } else if s.Find(".time").Text() == "パス（成功率）" {
              homePassCountString :=  r.FindAllString(s.Find(".home").Text(), 1)[0]
              homePassCount, _    = strconv.Atoi(homePassCountString)
              awayPassCountString :=  r.FindAllString(s.Find(".away").Text(), 1)[0]
              awayPassCount, _    = strconv.Atoi(awayPassCountString)
              homePassSuccessRateString :=  r.FindAllString(s.Find(".home").Text(), 2)[1]
              homePassSuccessRate, _    = strconv.Atoi(homePassSuccessRateString)
              awayPassSuccessRateString :=  r.FindAllString(s.Find(".away").Text(), 2)[1]
              awayPassSuccessRate, _ =  strconv.Atoi(awayPassSuccessRateString)
              fmt.Println(s.Find(".time").Text() + ": " + homeTeamName + " " + r.FindAllString(s.Find(".home").Text(), 1)[0] + " " + awayTeamName + r.FindAllString(s.Find(".away").Text(), 1)[0])
              fmt.Println(s.Find(".time").Text() + ": " + homeTeamName + " " + r.FindAllString(s.Find(".home").Text(), 2)[1] + " " + awayTeamName + r.FindAllString(s.Find(".away").Text(), 2)[1])
            } else if s.Find(".time").Text() == "警告・退場"{
              fmt.Println(s.Find(".time").Text() + ": " + homeTeamName + " " + r.FindAllString(s.Find(".home").Text(), 1)[0] + " " + awayTeamName + r.FindAllString(s.Find(".away").Text(), 1)[0])
              fmt.Println(s.Find(".time").Text() + ": " + homeTeamName + " " + r.FindAllString(s.Find(".home").Text(), 2)[1] + " " + awayTeamName + r.FindAllString(s.Find(".away").Text(), 2)[1])
            } else if s.Find(".time").Text() == "走行距離" {
              homeMileageString :=  r.FindAllString(s.Find(".home").Text(), 1)[0]
              homeMileage, _ =  strconv.Atoi(homeMileageString)
              awayMileageString :=  r.FindAllString(s.Find(".away").Text(), 1)[0]
              awayMileage, _ =  strconv.Atoi(awayMileageString)
              fmt.Println(s.Find(".time").Text() + ": " + homeTeamName + " " + strings.Replace(s.Find(".home").Text(), "km", "", 1) + " " + awayTeamName + strings.Replace(s.Find(".away").Text(), "km", "", 1))
            } else if s.Find(".time").Text() == "シュート" {
              homeShootCountString :=  s.Find(".home").Text()
              homeShootCount, _ =  strconv.Atoi(homeShootCountString)
              awayShootCountString :=  s.Find(".away").Text()
              awayShootCount, _ =  strconv.Atoi(awayShootCountString)
              fmt.Println(s.Find(".time").Text() + ": " + homeTeamName + " " + s.Find(".home").Text() + " " + awayTeamName + s.Find(".away").Text())
            } else if s.Find(".time").Text() == "枠内シュート" {
              homeFrameCountString :=  s.Find(".home").Text()
              homeFrameCount, _ =  strconv.Atoi(homeFrameCountString)
              awayFrameCountString :=  s.Find(".away").Text()
              homeFrameCount, _ =  strconv.Atoi(awayFrameCountString)
              fmt.Println(s.Find(".time").Text() + ": " + homeTeamName + " " + s.Find(".home").Text() + " " + awayTeamName + s.Find(".away").Text())
            } else if s.Find(".time").Text() == "スプリント" {
              homeSprintCountString :=  s.Find(".home").Text()
              homeSprintCount, _ =  strconv.Atoi(homeSprintCountString)
              awaySprintCountString :=  s.Find(".away").Text()
              awaySprintCount, _ =  strconv.Atoi(awaySprintCountString)
              fmt.Println(s.Find(".time").Text() + ": " + homeTeamName + " " + s.Find(".home").Text() + " " + awayTeamName + s.Find(".away").Text())
            } else if s.Find(".time").Text() == "オフサイド" {
              homeOffsideString :=  s.Find(".home").Text()
              homeOffside, _ =  strconv.Atoi(homeOffsideString)
              awayOffsideString :=  s.Find(".away").Text()
              awayOffside, _ =  strconv.Atoi(awayOffsideString)
              fmt.Println(s.Find(".time").Text() + ": " + homeTeamName + " " + s.Find(".home").Text() + " " + awayTeamName + s.Find(".away").Text())
            } else if s.Find(".time").Text() == "フリーキック" {
              homeFreekickCountString :=  s.Find(".home").Text()
              homeFreekickCount, _ =  strconv.Atoi(homeFreekickCountString)
              awayFreekickCountString :=  s.Find(".away").Text()
              awayFreekickCount, _ =  strconv.Atoi(awayFreekickCountString)
              fmt.Println(s.Find(".time").Text() + ": " + homeTeamName + " " + s.Find(".home").Text() + " " + awayTeamName + s.Find(".away").Text())
            } else if s.Find(".time").Text() == "コーナーキック" {
              homeCornerkickCountString :=  s.Find(".home").Text()
              homeCornerkickCount, _ =  strconv.Atoi(homeCornerkickCountString)
              awayCornerkickCountString :=  s.Find(".away").Text()
              awayCornerkickCount, _ =  strconv.Atoi(awayCornerkickCountString)
              fmt.Println(s.Find(".time").Text() + ": " + homeTeamName + " " + s.Find(".home").Text() + " " + awayTeamName + s.Find(".away").Text())
            } else if s.Find(".time").Text() == "ペナルティキック" {
              homePenaltyKickString :=  s.Find(".home").Text()
              homePenaltyKick, _ =  strconv.Atoi(homePenaltyKickString)
              awayPenaltyKickString :=  s.Find(".away").Text()
              awayPenaltyKick, _ =  strconv.Atoi(awayPenaltyKickString)
              fmt.Println(s.Find(".time").Text() + ": " + homeTeamName + " " + s.Find(".home").Text() + " " + awayTeamName + s.Find(".away").Text())
            }
          })
          
          var position string = ""
          
          // Home Starting member
          fmt.Println("Home Starting Member")
          doc.Find("#1st_mem").Find(".home.partsTable").Find("tr").Each(func(i int, s *goquery.Selection) {
            position_val := s.Find(".position").Text()
            if position_val != "" {
              position = position_val
            }
          
            player := s.Find(".player").Find("a").Text()
            playerUrl, _ := s.Find(".player").Find("a").Attr("href")
            playerUrlSplit    := strings.Split(playerUrl, "/")
            var playerId string = ""
            if len(playerUrlSplit) != 1 {
              playerId          = playerUrlSplit[4]
            }
            var player_change string = ""
            player_change = s.Find(".change").Text()
          
            if player != "" {
              fmt.Println(i)
              fmt.Println(position)
              fmt.Println(player+player_change)
        
              // match_players(home)
              insert, err := db.Query("INSERT INTO match_players(match_id, player_id, team_type, player_type, starting_flg, player_change, position, sort_no, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, playerId, 0, 0, 1, player_change, position, i, current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          })
          
          // Home Substitute member
          fmt.Println("Home Substitute Member")
          doc.Find("#2nd_mem").Find(".home.partsTable").Find("tr").Each(func(i int, s *goquery.Selection) {
            position_val := s.Find(".position").Text()
            if position_val != "" {
              position = position_val
            }
          
            player := s.Find(".player").Find("a").Text()
            playerUrl, _ := s.Find(".player").Find("a").Attr("href")
            playerUrlSplit    := strings.Split(playerUrl, "/")
            var playerId string = ""
            if len(playerUrlSplit) != 1 {
              playerId          = playerUrlSplit[4]
            }
            var player_change string = ""
            player_change = s.Find(".change").Text()
          
            if player != "" {
              fmt.Println(position)
              fmt.Println(player+player_change)
              // match_players(home)
              insert, err := db.Query("INSERT INTO match_players(match_id, player_id, team_type, player_type, player_change, position, sort_no, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, playerId, 0, 0, player_change, position, i, current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          })
          
          // Home Coach
          homeCoach := doc.Find("#3rd_mem .home .player").Text()
          fmt.Println("Home Coach")
          fmt.Println(homeCoach)
          // match_players(home)
          insert, err := db.Query("INSERT INTO match_players(match_id, team_type, player_type, player_name, created_at, updated_at) values (?, ?, ?, ?, ?, ?) ", matchId,  0, 1, homeCoach, current_time, current_time)
          defer insert.Close()
          if err != nil{
            log.Fatal(err)
          }
          
          // Away Starting member
          fmt.Println("Away Starting Member")
          doc.Find("#1st_mem").Find(".away.partsTable").Find("tr").Each(func(i int, s *goquery.Selection) {
            position_val := s.Find(".position").Text()
            if position_val != "" {
              position = position_val
            }
          
            player := s.Find(".player").Find("a").Text()
            playerUrl, _ := s.Find(".player").Find("a").Attr("href")
            playerUrlSplit    := strings.Split(playerUrl, "/")
            var playerId string = ""
            if len(playerUrlSplit) != 1 {
              playerId          = playerUrlSplit[4]
            }
            var player_change string = ""
            player_change = s.Find(".change").Text()
            
            if player != "" {
              fmt.Println(position)
              fmt.Println(player+player_change)
              // match_players(away)
              insert, err := db.Query("INSERT INTO match_players(match_id, player_id, team_type, player_type, starting_flg, player_change, position, sort_no, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, playerId, 1, 0, 1, player_change, position, i, current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          })
          
          // Away Substitute member
          fmt.Println("Away Substitute Member")
          doc.Find("#2nd_mem").Find(".away.partsTable").Find("tr").Each(func(i int, s *goquery.Selection) {
            position_val := s.Find(".position").Text()
            if position_val != "" {
              position = position_val
            }
          
            player := s.Find(".player").Find("a").Text()
            playerUrl, _ := s.Find(".player").Find("a").Attr("href")
            playerUrlSplit    := strings.Split(playerUrl, "/")
            var playerId string = ""
            if len(playerUrlSplit) != 1 {
              playerId          = playerUrlSplit[4]
            }
            var player_change string = ""
            player_change = s.Find(".change").Text()
          
            if player != "" {
              fmt.Println(position)
              fmt.Println(player+player_change)
              // match_players(away)
              insert, err := db.Query("INSERT INTO match_players(match_id, player_id, team_type, player_type, player_change, position, sort_no, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, playerId, 1, 0, player_change, position, i, current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          })
          
          // Away Coach
          awayCoach := doc.Find("#3rd_mem .away .player").Text()
          fmt.Println("Away Coach")
          fmt.Println(awayCoach)
          // match_players(home)
          insert, err = db.Query("INSERT INTO match_players(match_id, team_type, player_type, player_name, created_at, updated_at) values (?, ?, ?, ?, ?, ?) ", matchId, 1, 1, awayCoach, current_time, current_time)
          defer insert.Close()
          if err != nil{
            log.Fatal(err)
          }
          
          // Refree
          doc.Find("#modSoccerGameCondition > table > tbody > tr").Each(func(i int, s *goquery.Selection) {
            if i == 0 {
              chiefReferee := s.Find(".last").Text()
              fmt.Println("主審"+s.Find(".last").Text())
              // match_refrees
              insert, err := db.Query("INSERT INTO match_referees(match_id, referee_name, referee_type, created_at, updated_at) values (?, ?, ?, ?, ?) ", matchId, chiefReferee, 0, current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }else{
              assistantReferee := s.Find(".last").Text()
              fmt.Println("副審"+s.Find(".last").Text())
              // match_refrees
              insert, err := db.Query("INSERT INTO match_referees(match_id, referee_name, referee_type, created_at, updated_at) values (?, ?, ?, ?, ?) ", matchId, assistantReferee, 1, current_time, current_time)
              defer insert.Close()
              if err != nil{
                log.Fatal(err)
              }
            }
          })
          
          //matches [DONE]
          insert, err = db.Query("INSERT INTO matches(match_id, league_type, section, home_team_id, away_team_id, match_date, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?) ", matchId, 1, section, homeTeamId, awayTeamId,  matchTime, current_time, current_time)
          defer insert.Close()
          if err != nil{
            log.Fatal(err)
          }
        
          // match_infos(home) [DONE]
          fmt.Println("away_match_info")
	        insert, err = db.Query("INSERT INTO match_infos(match_id, team_type, team_id, first_point, second_point, control_rate, shoot_count, frame_count, mileage, sprint_count, pass_count, pass_success_rate, offside, freekick_count, cornerkick_count, penalty_kick, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, 0, homeTeamId, homeFirstScore, homeSecondScore, homeControleRate, homeShootCount, homeFrameCount, homeMileage, homeSprintCount, homePassCount, homePassSuccessRate, homeOffside, homeFreekickCount, homeCornerkickCount, homePenaltyKick, current_time, current_time)
          defer insert.Close()
          if err != nil{
            log.Fatal(err)
          }
        
          // match_infos(away) [DONE]
          fmt.Println("home_match_info")
	        insert, err = db.Query("INSERT INTO match_infos(match_id, team_type, team_id, first_point, second_point, control_rate, shoot_count, frame_count, mileage, sprint_count, pass_count, pass_success_rate, offside, freekick_count, cornerkick_count, penalty_kick, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", matchId, 1, awayTeamId, awayFirstScore, awaySecondScore, awayControleRate, awayShootCount, awayFrameCount, awayMileage, awaySprintCount, awayPassCount, awayPassSuccessRate, awayOffside, awayFreekickCount, awayCornerkickCount, awayPenaltyKick, current_time, current_time)
          defer insert.Close()
          if err != nil{
            log.Fatal(err)
          }
        
          if err != nil{
            log.Fatal(err)
          }
  
          fmt.Println("-----------------------------------")
  
        }
      }
      return nil
    })
  })
}


type MyDB struct {
  db *sql.DB
}

var Db MyDB

// Connection
func (m *MyDB) Connection() error {
  var err error
  //m.db, err = sql.Open("mysql", "root:@/soccor_scoring_development")
  m.db, err = sql.Open("mysql", "hhvuafdv51p2lm3d:zpzmno3c34ywxu76@tcp(pfw0ltdr46khxib3.cbetxkdyhwsb.us-east-1.rds.amazonaws.com:3306)/l1crt4uh0916tivw?")
  if err != nil {
      return err
  }
  return nil
}

// Close
func (m *MyDB) Close() {
  if m.db != nil {
      m.db.Close()
  }
}

// Fetch
func (m *MyDB) Fetch(query string, tx *sql.Tx) ([][]interface{}, error) {
  rows, err := tx.Query(query)
  if err != nil {
      return nil, err
  }
  defer rows.Close()

  columns, _ := rows.Columns()
  count := len(columns)
  valuePtrs := make([]interface{}, count)

  ret := make([][]interface{}, 0)
  for rows.Next() {

      values := make([]interface{}, count)
      for i, _ := range columns {
          valuePtrs[i] = &values[i]
      }
      rows.Scan(valuePtrs...)

      for i, _ := range columns {
          var v interface{}
          val := values[i]
          b, ok := val.([]byte)
          if ok {
              v = string(b)
          } else {
              v = val
          }
          values[i] = v
      }
      ret = append(ret, values)
  }

  return ret, nil
}

//　Transaction
func (m *MyDB) Transaction(txFunc func(*sql.Tx) error) error {
  tx, err := m.db.Begin()
  if err != nil {
      return err
  }

  defer func() {
      if p := recover(); p != nil {
          fmt.Println("recover")
          tx.Rollback()
          panic(p)
      } else if err != nil {
          fmt.Println("rollback")
          tx.Rollback()
      } else {
          fmt.Println("commit")
          err = tx.Commit()
      }
  }()
  err = txFunc(tx)
  return err
}
