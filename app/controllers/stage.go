package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

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
