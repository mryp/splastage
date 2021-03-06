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

//定数
const (
	urlNawabariStageJSON    = "http://s3-ap-northeast-1.amazonaws.com/splatoon-data.nintendo.net/stages_info.json"
	urlFesStageJSON         = "http://s3-ap-northeast-1.amazonaws.com/splatoon-data.nintendo.net/fes_info.json"
	urlIkaringAuth          = "https://splatoon.nintendo.net/users/auth/nintendo"
	urlNintendoLoginPost    = "https://id.nintendo.net/oauth/authorize"
	urlIkaringSchedule      = "https://splatoon.nintendo.net/schedule"
	urlNawabariImageBaseURL = "http://www.nintendo.co.jp/wiiu/agmj/stage/images/stage/"
)

//変数
var (
	defUnknownTime = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
)

//Now 現在開催しているステージ情報を取得する
func (c Stage) Now(id string) revel.Result {
	if id == "" {
		return c.Forbidden("パラメーターエラー")
	}
	revel.INFO.Println("call stage/now", c.Request.UserAgent(), c.Request.Host, id)
	models.AccessLogInsert(DbMap, models.AccessLogCreate("stage/now", id, c.Request))

	stageList := models.StageSelectNow(DbMap)
	return c.RenderJson(stageList)
}

//CurrentLater 現在時刻以降開催しているステージ情報を取得する（現在時刻を含む）
func (c Stage) CurrentLater(id string) revel.Result {
	if id == "" {
		return c.Forbidden("パラメーターエラー")
	}
	revel.INFO.Println("call stage/carrentlater", c.Request.UserAgent(), c.Request.Host, id)
	models.AccessLogInsert(DbMap, models.AccessLogCreate("stage/carrentlater", id, c.Request))

	stageList := models.StageSelectCurrentLater(DbMap)
	return c.RenderJson(stageList)
}

//UpdateStageFromNawabari ステージ情報を取得してDBに保存する
//return bool 更新を行ったときはtrue
func UpdateStageFromNawabari() bool {
	revel.INFO.Println("call UpdateStageFromNawabari")
	//itemList := getNawabariStageInfo()//ナワバリ情報
	itemList := getFesStageInfo() //フェス情報
	if itemList == nil {
		revel.WARN.Println("データなし")
		return false
	}

	isUpdate := insertStageList(itemList)
	return isUpdate
}

//UpdateStageFromIkaring イカリングからステージ情報を取得してDBに保存する
//return bool 更新を行ったときはtrue
func UpdateStageFromIkaring() bool {
	revel.INFO.Println("call UpdateStageFromIkaring")
	itemList := getIkaringStageInfo()
	if itemList == nil {
		revel.WARN.Println("データ取得失敗")
		return false
	}

	isUpdate := insertStageList(itemList)
	return isUpdate
}

//ステージ情報リストをDBに保存する
//同じデータがすでに存在するときは追加しない
//return bool 追加処理が1件以上行われたときはtrue
func insertStageList(itemList []models.Stage) bool {
	isUpdate := false
	for _, item := range itemList {
		ret, err := models.StageInsertIfNotExists(DbMap, item)
		if err != nil {
			revel.WARN.Println("データの追加に失敗", err)
		}
		if ret {
			isUpdate = true
		}
	}

	return isUpdate
}

//現在の最新データをダウンロードして返す
func getNawabariStageInfo() []models.Stage {
	unixTime := fmt.Sprintf("%d", time.Now().Unix())
	url := urlNawabariStageJSON + "?" + unixTime

	resp, err := http.Get(url)
	if err != nil {
		revel.ERROR.Println("ダウンロード失敗", url, err)
		return nil
	}
	defer resp.Body.Close()
	byteArray, _ := ioutil.ReadAll(resp.Body)

	var output interface{}
	output, err = jsonUnmarshal(byteArray)
	if err != nil {
		revel.ERROR.Println("Jsonオブジェクト変換失敗", err)
		return nil
	}

	return jsonParseRoot(output)
}

func getFesStageInfo() []models.Stage {
	unixTime := fmt.Sprintf("%d", time.Now().Unix())
	url := urlFesStageJSON + "?" + unixTime

	resp, err := http.Get(url)
	if err != nil {
		revel.ERROR.Println("ダウンロード失敗", url, err)
		return nil
	}
	defer resp.Body.Close()
	byteArray, _ := ioutil.ReadAll(resp.Body)

	var output interface{}
	output, err = jsonUnmarshal(byteArray)
	if err != nil {
		revel.ERROR.Println("Jsonオブジェクト変換失敗", err)
		return nil
	}

	return jsonParseFesRoot(output)
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
		startTime := defUnknownTime
		endTime := defUnknownTime
		nameList := []string{}
		imageList := []string{}
		for key, v := range item.(map[string]interface{}) {
			switch key {
			case "datetime_term_begin":
				startTime = convertTermTimeStr(v.(string))
			case "datetime_term_end":
				endTime = convertTermTimeStr(v.(string))
			case "stages":
				nameList, imageList = jsonParseStage(v)
			}
		}

		for i, name := range nameList {
			imageURL := imageList[i]
			stage := models.Stage{Rule: "ナワバリバトル", MatchType: "レギュラーマッチ", Name: name, StartTime: startTime, EndTime: endTime, ImageURL: imageURL}
			stageList = append(stageList, stage)
		}
	}

	return stageList
}

func jsonParseFesRoot(data interface{}) []models.Stage {
	stageList := []models.Stage{}
	startTime := defUnknownTime
	endTime := defUnknownTime
	nameList := []string{}
	imageList := []string{}
	for key, v := range data.(map[string]interface{}) {
		switch key {
		case "datetime_fes_begin":
			startTime = convertTermTimeStr(v.(string))
		case "datetime_fes_end":
			endTime = convertTermTimeStr(v.(string))
		case "fes_stages":
			nameList, imageList = jsonParseStage(v)
		}
	}

	for i, name := range nameList {
		imageURL := imageList[i]
		stage := models.Stage{Rule: "ナワバリバトル", MatchType: "フェスマッチ", Name: name, StartTime: startTime, EndTime: endTime, ImageURL: imageURL}
		stageList = append(stageList, stage)
	}

	return stageList
}

//ステージ情報時刻文字列をtime.Time型に変換して返す
func convertTermTimeStr(strTime string) time.Time {
	var timeFormat = "2006-01-02 15:04"
	result, err := time.Parse(timeFormat, strTime)
	if err != nil {
		result = defUnknownTime
	}
	return result
}

//ステージ情報のステージ名オブジェクトからステージ名リストを生成して返す
func jsonParseStage(data interface{}) ([]string, []string) {
	nameList := []string{}
	imageURLList := []string{}
	for _, stage := range data.([]interface{}) {
		for key, value := range stage.(map[string]interface{}) {
			switch key {
			case "id":
				imageURL := "http://www.nintendo.co.jp/wiiu/agmj/stage/images/stage/" + value.(string) + ".png"
				imageURLList = append(imageURLList, imageURL)
			case "name":
				nameList = append(nameList, value.(string))
			}
		}
	}

	return nameList, imageURLList
}

//イカリングからステージ情報を取得して返す
func getIkaringStageInfo() []models.Stage {
	//認証用の情報を取得
	userid, _ := revel.Config.String("my.nintendo_userid")
	password, _ := revel.Config.String("my.nintendo_password")
	authData := getNintendoAuthorize(urlIkaringAuth, userid, password)
	if authData == nil {
		revel.WARN.Println("ニンテンドーネットワーク認証データ取得失敗")
		return nil
	}

	//ログインを行いクッキーをセットした接続クライアントを取得する
	loginClient := getNintendoLoginClient(urlNintendoLoginPost, authData)
	if loginClient == nil {
		revel.WARN.Println("ニンテンドーネットワークログイン失敗")
		return nil
	}

	//クッキーをセットしたクライアントからHTMLを取得する
	scheduleHTML := getIkaringScheduleHTML(urlIkaringSchedule, loginClient)
	if scheduleHTML == "" {
		revel.WARN.Println("イカリングステージ情報データ取得失敗")
		return nil
	}

	//HTMLからステージ情報を取得
	output := convertIkaringHTML(scheduleHTML)
	if output == nil {
		revel.WARN.Println("イカリングステージ情報データ変換失敗")
		return nil
	}

	return output
}

//ニンテンドーネットワークID認証用データを取得する
func getNintendoAuthorize(authURL string, userid string, password string) url.Values {
	doc, err := goquery.NewDocument(authURL)
	if err != nil {
		revel.WARN.Println("goquery.NewDocument", err)
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

//イカリングログインを行った認証クライアントを作成する
func getNintendoLoginClient(postURL string, authData url.Values) *http.Client {
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}

	resp, err := client.PostForm(postURL, authData)
	if err != nil {
		revel.WARN.Println("client.PostForm", err)
		return nil
	}
	defer resp.Body.Close()

	return client
}

//イカリングスケジュールHTMLをダウンロードする
func getIkaringScheduleHTML(getURL string, client *http.Client) string {
	getdata, err := client.Get(getURL)
	if err != nil {
		revel.WARN.Println("client.Get", err)
		return ""
	}
	defer getdata.Body.Close()

	var body []byte
	body, err = ioutil.ReadAll(getdata.Body)
	if err != nil {
		revel.WARN.Println("ioutil.ReadAll", err)
		return ""
	}

	return string(body)
}

//イカリングスケジュールHTMLからステージ情報リストを取得する
func convertIkaringHTML(html string) []models.Stage {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		revel.WARN.Println("goquery.NewDocumentFromReader", err)
		return nil
	}

	stageList := []models.Stage{}
	startTime := defUnknownTime
	endTime := defUnknownTime
	matchType := ""
	rule := ""
	imageURL := ""
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
		} else if s.HasClass("map-image retina-support") {
			imageURL, _ = s.Attr("data-retina-image")
			imageURL = "https://splatoon.nintendo.net" + imageURL
		} else if s.HasClass("map-name") {
			name = s.Text()
			stage := models.Stage{Rule: rule, MatchType: matchType, Name: name, ImageURL: imageURL, StartTime: startTime, EndTime: endTime}
			stageList = append(stageList, stage)
		}
	})

	return stageList
}

//イカリングのステージ時刻情報を変換する
//strTime "8/7 11:00"のような文字列
func convertStageScheduleTimeStr(strTime string) time.Time {
	var timeFormat = "2006/1/2 15:04"
	result, err := time.Parse(timeFormat, strconv.Itoa(time.Now().Year())+"/"+strTime)
	if err != nil {
		result = defUnknownTime
		revel.WARN.Println("time.Parse", strTime, err)
	}

	revel.INFO.Println("convert time", result, result.UTC())
	return result
}
