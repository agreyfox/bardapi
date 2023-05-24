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

	return &BardApi{
		BaseURL: baseURL,
		Headers: http.Header{
			"Host":          []string{"bard.google.com"},
			"X-Same-Domain": []string{"1"},
			"User-Agent":    []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36"},
			"Content-Type":  []string{"application/x-www-form-urlencoded;charset=UTF-8"},
			"Origin":        []string{baseURL},
			"Referer":       []string{baseURL + "/"},
			"Cookie":        []string{"__Secure-1PSID=" + sessionID},
		},
		RequestID: rand.Int63n(9999),
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

func (b *BardApi) generateRequestBody(message string, options Options) (url.Values, error) {
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

func (b *BardApi) findImage(repondid string, msg [][]interface{}) (string, error) {
	retImageurl := ""

	return retImageurl, nil
}

func (b *BardApi) replaceImageUrls(input string, msg [][]interface{}) string {
	re := regexp.MustCompile(Match_Image)
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		url, err := b.findImage(match[0], msg)
		if err == nil {
			replacement := fmt.Sprintf("'%s(%s)'", match[0], url)
			input = re.ReplaceAllString(input, replacement)
		} else {
			fmt.Println("image url not found!")
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
	err = json.Unmarshal([]byte(responseBody[0][2]), &responseMessage)
	if err != nil {
		return nil, err
	}

	responseID, ok := responseMessage[1][1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid responseID: %s", responseMessage[1][1])
	}
	conversationID, ok := responseMessage[1][0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid conversationID: %s", responseMessage[1][0])
	}
	question, ok := responseMessage[2][0].([]interface{})[0].(string)
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

		choice := Choice{
			ChoiceID: choiceID,
			Answer:   b.replaceImageUrls(answer, responseMessage),
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

func (b *BardApi) SendMessage(message string, options Options) (*ResponseBody, error) {
	params := fmt.Sprintf("?bl=%s&_reqid=%d&rt=%s", "boq_assistant-bard-web-server_20230402.21_p0", b.RequestID, "c")
	requestBody, err := b.generateRequestBody(message, options)
	if err != nil {

		return nil, nil
	}

	response, err := b.sendRequest(params, requestBody)
	if err != nil {
		return nil, nil
	}

	return b.handleResponse(response)
}
