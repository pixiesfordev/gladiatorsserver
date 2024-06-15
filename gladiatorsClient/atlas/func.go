package atlas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gladiatorsClient/setting"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func CallAtlasFunc(funcName string, keyValues map[string]string) {
	url := fmt.Sprintf("https://realm.mongodb.com/api/client/v2.0/app/%s/functions/%s", setting.AppID, funcName)
	jsonBytes, err := json.Marshal(keyValues)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %s", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		fmt.Println("Error creating request: ", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", setting.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request: ", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response: ", err)
		return
	}

	log.Infof("Response: %v", string(body))
}
