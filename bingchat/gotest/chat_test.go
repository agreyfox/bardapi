package gotest

import (
	"os"
	"strings"
	"testing"
	"time"

	bapi "github.com/agreyfox/bardapi/bingchat"
)

const ENV_Cookie = "BingChatCookie"

func Test_ConversationStream(t *testing.T) {
	os.Setenv(ENV_Cookie, "/home/lq/ftrader/config/bing.json")
	chat, err := bapi.NewBingChat(os.Getenv(ENV_Cookie), bapi.ConversationBalanceStyle, 2*time.Minute)
	if err != nil {
		panic(err)
	}
	t.Log(chat)

	message, err := chat.SendMessage("如果需要利用golang来编写wasm的code，你建议用什么编程环境 ?")
	if err != nil {
		panic(err)
	}
	//t.Logf("%v\n", message)
	var respBuilder strings.Builder
	for {
		msg, ok := <-message.Notify
		if !ok {
			break
		}
		t.Log(msg)
		respBuilder.WriteString(msg)
	}

	t.Logf("%+v", message.Suggest)

	t.Logf("%s", respBuilder.String())

}
