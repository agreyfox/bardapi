package bard

import "net/http"

const Prefix_Image = "[Image of "
const Match_Image = `\[Image of (.+?)\]`
const Match_Video = `\[Video of (.+?)\]`
const BardHost = "bard.google.com"
const BardBaseURL = "https://bard.google.com"

type BardApi struct {
	BaseURL   string
	Headers   http.Header
	RequestID int64
	Options   *Options
}

type Options struct {
	ConversationID string
	ResponseID     string
	ChoiceID       string
}

type RequestBody struct {
	FReq string `json:"f.req"`
	At   string `json:"at"`
}

type Choice struct {
	ChoiceID string `json:"choice_id"`
	Answer   string `json:"answer"`
}

type ResponseBody struct {
	ResponseID     string `json:"response_id"`
	ConversationID string `json:"conversation_id"`
	Question       string `json:"question"`
	Choices        []Choice
}
