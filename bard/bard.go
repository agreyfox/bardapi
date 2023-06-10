package bard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func NewBard(sessionID string) *BardApi {
	baseURL := BardBaseURL
	bardOption := Options{
		ConversationID: "",
		ResponseID:     "",
		ChoiceID:       "",
	}
	return &BardApi{
		BaseURL: baseURL,
		Headers: http.Header{
			"Host":          []string{BardHost},
			"X-Same-Domain": []string{"1"},
			"User-Agent":    []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36"},
			"Content-Type":  []string{"application/x-www-form-urlencoded;charset=UTF-8"},
			"Origin":        []string{baseURL},
			"Referer":       []string{baseURL + "/"},
			"Cookie":        []string{"__Secure-1PSID=" + sessionID},
		},
		RequestID: rand.Int63n(19999),
		Options:   &bardOption,
	}
}

func (b *BardApi) getSNlM0e() (string, error) {
	request, _ := http.NewRequest("GET", b.BaseURL, nil)

	request.Header = b.Headers

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	snlm0e := regexp.MustCompile(`"SNlM0e":"(.*?)"`).FindStringSubmatch(string(bodyBytes))[1]

	return snlm0e, nil
}

func (b *BardApi) generateRequestBody(message string, options *Options) (url.Values, error) {
	messageBytes, err := json.Marshal([][]string{
		{message},
		nil,
		{options.ConversationID, options.ResponseID, options.ChoiceID},
	})
	if err != nil {
		return nil, err
	}

	messageString := string(messageBytes)
	fReqBytes, err := json.Marshal([]*string{
		nil,
		&messageString,
	})
	if err != nil {
		return nil, err
	}

	snlm0e, err := b.getSNlM0e()
	if err != nil {
		return nil, err
	}

	requestBody := url.Values{
		"f.req": []string{string(fReqBytes)},
		"at":    []string{snlm0e},
	}

	return requestBody, nil
}

func (b *BardApi) sendRequest(params string, requestBody url.Values) (*http.Response, error) {
	request, _ := http.NewRequest("POST", b.BaseURL+"/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate"+params, bytes.NewBufferString(requestBody.Encode()))

	request.Header = b.Headers
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	client := &http.Client{}

	return client.Do(request)
}

func (b *BardApi) findImage(repondid string, msg []interface{}) (string, string, error) {
	retImageurl := ""
	retfulImageurl := ""
	for _, item := range msg {
		iteminfo := item.([]interface{})
		if len(iteminfo) == 0 {
			continue
		}
		fullimgurl := iteminfo[0].([]interface{})
		shortimgurl := iteminfo[3].([]interface{})
		fullname := iteminfo[2].(string)
		//fmt.Printf("%s--->%v\n", fullname, fullimgurl,shortimgurl)
		if repondid == fullname && len(shortimgurl) > 0 {

			retImageurl = shortimgurl[0].([]interface{})[0].(string)
			retfulImageurl = fullimgurl[0].([]interface{})[0].(string)
			break
		}
	}
	return retImageurl, retfulImageurl, nil
}

func (b *BardApi) replaceImageUrls(input string, msg []interface{}) string {
	re := regexp.MustCompile(Match_Image)
	matches := re.FindAllStringSubmatch(input, -1)
	if len(msg) >= 5 {
		metainfo := msg[4] //there is how many respond

		for _, match := range matches {
			if len(match) == 0 {
				continue
			}
			url, fullurl, err := b.findImage(match[0], metainfo.([]interface{}))
			if err == nil {
				replacement := fmt.Sprintf(`<img alt=%q src=%q width=%q>`, match[1], url, "50%")
				input = strings.Replace(input, match[0], replacement, -1)
				fmt.Printf("image:%s,thumbnail:%s,full:%s\n", match[1], url, fullurl)

			} else {
				fmt.Println("image url not found!")
			}

		}
	}

	return input
}

func (b *BardApi) replaceVideoUrls(input string, msg []interface{}) string {
	re := regexp.MustCompile(Match_Video)
	matches := re.FindAllStringSubmatch(input, -1)

	if len(msg) >= 5 {
		metainfo := msg[4] //there is how many respond

		for _, match := range matches {
			if len(match) == 0 {
				continue
			}
			url, fullurl, err := b.findImage(match[0], metainfo.([]interface{}))
			//thumbpic, err := b.findImage(match[0], metainfo.([]interface{}))
			if err == nil {
				replacement := fmt.Sprintf("[![%s](%s)](width=50) ]", match[1], url)
				input = strings.Replace(input, match[0], replacement, -1)
				fmt.Printf("video:%s,thumbnail:%s,full:%s\n", match[1], url, fullurl)
			} else {
				fmt.Println("image url not found!")
			}

		}
	}

	return input
}
func (b *BardApi) handleResponse(response *http.Response) (*ResponseBody, error) {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	bodySplit := strings.Split(string(bodyBytes), "\n")

	if len(bodySplit) < 8 {
		return nil, fmt.Errorf("invalid response body: %s", string(bodyBytes))
	}

	var responseBody [][]string
	err = json.Unmarshal([]byte(bodySplit[3]), &responseBody)
	if err != nil {
		return nil, err
	}

	if len(responseBody) < 1 || len(responseBody[0]) < 3 {
		return nil, fmt.Errorf("invalid response body: %s", responseBody)
	}

	var responseMessage [][]interface{}
	fmt.Println(responseBody[0][2])
	json.Unmarshal([]byte(responseBody[0][2]), &responseMessage)
	/* if err != nil {
		fmt.Println(err.Error())
		return nil, err
	} */
	if responseMessage == nil {
		return nil, fmt.Errorf("invalid response body: %s", responseBody)
	}
	responseID, ok := responseMessage[1][1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid responseID: %s", responseMessage[1][1])
	}
	conversationID, ok := responseMessage[1][0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid conversationID: %s", responseMessage[1][0])
	}
	resp2 := responseMessage[2]
	resp1 := responseMessage[0][0]
	if resp2 == nil {

		responseStruct := &ResponseBody{
			ResponseID:     responseID,
			ConversationID: conversationID,
			Question:       resp1.(string),
			Choices: []Choice{
				{
					ChoiceID: "xxx",
					Answer:   resp1.(string),
				},
			},
		}
		return responseStruct, nil
	}

	question, ok := resp2[0].([]interface{})[0].(string)

	if !ok {
		return nil, fmt.Errorf("invalid question: %s", responseMessage[2][0])
	}

	var choices []Choice
	for _, c := range responseMessage[4] {
		choiceID, ok := c.([]interface{})[0].(string)
		if !ok {
			continue
		}

		answer, ok := c.([]interface{})[1].([]interface{})[0].(string)
		if !ok {
			continue
		}
		//meta := c.([]interface{})

		meta := c.([]interface{})
		if len(meta) >= 5 {
			answer = b.replaceImageUrls(answer, meta)
			answer = b.replaceVideoUrls(answer, meta)
		}
		choice := Choice{
			ChoiceID: choiceID,
			Answer:   answer,
		}

		choices = append(choices, choice)
	}

	responseStruct := &ResponseBody{
		ResponseID:     responseID,
		ConversationID: conversationID,
		Question:       question,
		Choices:        choices,
	}

	b.RequestID += 100000

	return responseStruct, nil
}

func (b *BardApi) SendMessage(message string) (*ResponseBody, error) {

	params := fmt.Sprintf("?bl=%s&_reqid=%d&rt=%s", "boq_assistant-bard-web-server_20230402.21_p0", b.RequestID, "c")
	requestBody, err := b.generateRequestBody(message, b.Options)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	response, err := b.sendRequest(params, requestBody)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resp, err := b.handleResponse(response)
	if err != nil {
		return nil, err
	} else {
		b.Options.ConversationID = resp.ConversationID
		b.Options.ResponseID = resp.ResponseID
		b.Options.ChoiceID = resp.Choices[0].ChoiceID
		return resp, nil
	}
}
