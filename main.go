package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/agreyfox/bardapi/bard"
	"github.com/agreyfox/bardapi/translate"
)

func main() {
	googleTranslate := translate.NewGoogle("zh-CN", "en")

	sessionID := ""

	b := bard.NewBard(sessionID)
	bardOptions := bard.Options{
		ConversationID: "",
		ResponseID:     "",
		ChoiceID:       "",
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("You: ")
		scanner.Scan()
		message := scanner.Text()

		translateMessage, err := googleTranslate.Translate(message)
		if err != nil {
			log.Fatalln(err)
		}

		response, err := b.SendMessage(translateMessage, bardOptions)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Printf("Bard: %s\n\n", response.Choices[0].Answer)

		bardOptions.ConversationID = response.ConversationID
		bardOptions.ResponseID = response.ResponseID
		bardOptions.ChoiceID = response.Choices[0].ChoiceID
	}
}
