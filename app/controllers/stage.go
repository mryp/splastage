package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mryp/splastage/app/models"
	"github.com/revel/revel"
)

//Stage 構造体
type Stage struct {
	GorpController
}

//変数
var DefUnknownTime = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

//Latest 最新のステージ情報を取得する
func (c Stage) Latest() revel.Result {
	stageList := models.StageSelectLast(DbMap)
	return c.RenderJson(stageList)
}

//SelectAll 保存されているステージ情報をすべて取得し表示する
func (c Stage) SelectAll() revel.Result {
	stageList := models.StageSelectAll(DbMap)
	return c.RenderJson(stageList)
}

//ステージ情報を取得してDBに保存する
//return bool 更新を行ったときはtrue
func UpdateStageInfo() bool {
	revel.INFO.Println("call UpdateStageInfo")
	itemList := getNowStageInfo()
	if itemList == nil {
		revel.INFO.Println("getNowStageInfo error")
		return false
	}

	isUpdate := false
	for _, item := range itemList {
		ret, err := models.StageInsertIfNotExists(DbMap, item)
		if err != nil {
			revel.INFO.Println("StageInsertIfNotExists error", err)
		}
		if ret {
			isUpdate = true
		}
	}

	return isUpdate
}

//現在の最新データをダウンロードして返す
func getNowStageInfo() []models.Stage {
	unixTime := fmt.Sprintf("%d", time.Now().Unix())
	url := "http://s3-ap-northeast-1.amazonaws.com/splatoon-data.nintendo.net/stages_info.json?" + unixTime

	resp, err := http.Get(url)
	if err != nil {
		revel.ERROR.Println("ダウンロードエラー", err)
		return nil
	}
	defer resp.Body.Close()
	byteArray, _ := ioutil.ReadAll(resp.Body)

	output, err2 := jsonUnmarshal(byteArray)
	if err2 != nil {
		revel.ERROR.Println("変換エラー", err2)
		return nil
	}

	return jsonParseRoot(output)
}

//JSONダウンロードデータをデータオブジェクト（interface{}）に変換して返す
func jsonUnmarshal(data []byte) (interface{}, error) {
	var outputData interface{}
	err := json.Unmarshal(data, &outputData)
	if err != nil {
		return nil, err
	}

	return outputData, nil
}

//ステージ情報のJSONオブジェクトからステージ情報リストを作成して返す
func jsonParseRoot(data interface{}) []models.Stage {
	stageList := []models.Stage{}
	for _, item := range data.([]interface{}) {
		startTime := DefUnknownTime
		endTime := DefUnknownTime
		nameList := []string{}
		for key, v := range item.(map[string]interface{}) {
			switch key {
			case "datetime_term_begin":
				startTime = convertTermTimeStr(v.(string))
			case "datetime_term_end":
				endTime = convertTermTimeStr(v.(string))
			case "stages":
				nameList = jsonParseStage(v)
			}
		}

		for _, name := range nameList {
			stage := models.Stage{Rule: "ナワバリバトル", MatchType: "レギュラーマッチ", Name: name, StartTime: startTime, EndTime: endTime}
			stageList = append(stageList, stage)
		}
	}

	return stageList
}

//ステージ情報時刻文字列をtime.Time型に変換して返す
func convertTermTimeStr(strTime string) time.Time {
	var timeFormat = "2006-01-02 15:04"
	result, err := time.Parse(timeFormat, strTime)
	if err != nil {
		result = DefUnknownTime
	}
	return result
}

//ステージ情報のステージ名オブジェクトからステージ名リストを生成して返す
func jsonParseStage(data interface{}) []string {
	result := []string{}
	for _, stage := range data.([]interface{}) {
		for key, value := range stage.(map[string]interface{}) {
			switch key {
			case "id":
				//何もしない
			case "name":
				result = append(result, value.(string))
			}
		}
	}

	return result
}

func (c Stage) Test() revel.Result {
	//認証用の情報を取得
	userid, _ := revel.Config.String("my.nintendo_userid")
	password, _ := revel.Config.String("my.nintendo_password")
	authUrl := "https://splatoon.nintendo.net/users/auth/nintendo"
	authData := getNintendoAuthorize(authUrl, userid, password)
	if authData == nil {
		return c.RenderJson("ERROR getNintendoAuthorize")
	}

	//ログインを行いクッキーをセットした接続クライアントを取得する
	loginPostUrl := "https://id.nintendo.net/oauth/authorize"
	loginClient := getNintendoLoginClient(loginPostUrl, authData)
	if loginClient == nil {
		return c.RenderJson("ERROR2 getNintendoLoginClient")
	}

	//クッキーをセットしたクライアントからHTMLを取得する
	scheduleUrl := "https://splatoon.nintendo.net/schedule"
	scheduleHtml := getStageScheduleHtml(scheduleUrl, loginClient)
	if scheduleHtml == "" {
		return c.RenderJson("ERROR3 getStageScheduleHtml")
	}

	//HTMLからステージ情報を取得
	output := convertIkaringHtml(scheduleHtml)
	if output == nil {
		return c.RenderJson("ERROR4 convertIkaringHtml")
	}

	return c.RenderJson(output)
}

func getNintendoAuthorize(authUrl string, userid string, password string) url.Values {
	doc, err := goquery.NewDocument(authUrl)
	if err != nil {
		return nil
	}

	values := url.Values{}
	doc.Find("input").Each(func(_ int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		switch name {
		case "client_id":
			v, _ := s.Attr("value")
			values.Add(name, v)
		case "response_type":
			v, _ := s.Attr("value")
			values.Add(name, v)
		case "state":
			v, _ := s.Attr("value")
			values.Add(name, v)
		case "redirect_uri":
			v, _ := s.Attr("value")
			values.Add(name, v)
		}
	})
	values.Add("lang", "ja-JP")
	values.Add("nintendo_authenticate", "")
	values.Add("nintendo_authorize", "")
	values.Add("scope", "")
	values.Add("username", userid)
	values.Add("password", password)

	return values
}

func getNintendoLoginClient(postUrl string, authData url.Values) *http.Client {
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}

	resp, err := client.PostForm(postUrl, authData)
	if err != nil {
		revel.INFO.Println("getNintendoLoginClient", err)
		return nil
	}
	defer resp.Body.Close()

	return client
}

func getStageScheduleHtml(getUrl string, client *http.Client) string {
	getdata, err := client.Get(getUrl)
	if err != nil {
		revel.INFO.Println("getStageScheduleHtml", err)
		return ""
	}
	defer getdata.Body.Close()

	body, err2 := ioutil.ReadAll(getdata.Body)
	if err2 != nil {
		revel.INFO.Println("getStageScheduleHtml", err2)
		return ""
	}

	return string(body)
}

func convertIkaringHtml(html string) []models.Stage {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}

	stageList := []models.Stage{}
	startTime := DefUnknownTime
	endTime := DefUnknownTime
	matchType := ""
	rule := ""
	name := ""
	doc.Find("span").Each(func(_ int, s *goquery.Selection) {
		if s.HasClass("stage-schedule") {
			timeSplit := strings.Split(s.Text(), " ~ ")
			if len(timeSplit) >= 2 {
				startTime = convertStageScheduleTimeStr(timeSplit[0])
				endTime = convertStageScheduleTimeStr(timeSplit[1])
			}
		} else if s.HasClass("icon-regular-match") {
			matchType = "レギュラーマッチ"
			rule = "ナワバリバトル"
		} else if s.HasClass("icon-earnest-match") {
			matchType = "ガチマッチ"
		} else if s.HasClass("rule-description") {
			rule = s.Text()
		} else if s.HasClass("map-name") {
			name = s.Text()
			stage := models.Stage{Rule: rule, MatchType: matchType, Name: name, StartTime: startTime, EndTime: endTime}
			stageList = append(stageList, stage)
		}
	})

	return stageList
}

func convertStageScheduleTimeStr(strTime string) time.Time {

	var timeFormat = "2006/1/2 15:04"
	result, err := time.Parse(timeFormat, strconv.Itoa(time.Now().Year())+"/"+strTime)
	if err != nil {
		result = DefUnknownTime
		revel.INFO.Println("convertStageScheduleTimeStr strTime=" + strTime)
	}
	return result
}
