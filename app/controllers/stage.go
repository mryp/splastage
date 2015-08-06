package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/revel/revel"
)

//Stage 構造体
type Stage struct {
	*revel.Controller
}

//StageItem 構造体
type StageItem struct {
	Kind      string   `json:"kind"`
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	StageName []string `json:"stage_name"`
}

//Now 現在のステージ情報を取得する
func (c Stage) Now() revel.Result {
	itemList := getNowStageInfo()
	//	item := StageItem{StartTime: "2015/8/6 11:00", EndTime: "2015/8/6 15:00", StageName: "ホッケ埠頭", Kind: "ナワバリバトル"}
	//	itemList := [...]StageItem{item, item}
	return c.RenderJson(itemList)
}

//現在の最新データをダウンロードして返す
func getNowStageInfo() []StageItem {
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
func jsonParseRoot(data interface{}) []StageItem {
	stageList := []StageItem{}
	for _, item := range data.([]interface{}) {
		stage := StageItem{Kind: "ナワバリバトル"}
		for key, v := range item.(map[string]interface{}) {
			switch key {
			case "datetime_term_begin":
				stage.StartTime = v.(string)
			case "datetime_term_end":
				stage.EndTime = v.(string)
			case "stages":
				stage.StageName = jsonParseStage(v)
			}
		}

		stageList = append(stageList, stage)
	}

	return stageList
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
