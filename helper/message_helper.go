package helper

import (
	"fmt"
	"io/ioutil"
	"time"
	"wx-cli/client"
)

func HandleMessage(msg *client.Message) string {
	var text string
	if msg.IsText() {
		text = handleText(msg)
	} else if msg.IsPicture() {
		text = handlePicture(msg)
	} else if msg.IsArticle() {
		text = handleArticle(msg)
	} else if msg.IsTickled() {

	} else if msg.IsSystem() {
		text = handleText(msg)
	}
	sender, _ := msg.Sender()
	sender.IsFriend()
	return text
}

func handleText(msg *client.Message) string {
	return msg.Content
}

func handlePicture(msg *client.Message) string {
	resp, err := msg.GetPicture()
	text := "[Picture]"
	if err == nil {
		length := resp.ContentLength
		buf := make([]byte, length)
		buf, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			filename := fmt.Sprintf("%v.jpg", time.Now().UnixNano())
			err := ioutil.WriteFile(filename, buf, 0666)
			if err != nil {
				fmt.Println("write file err:", err)
			}
		}
	}
	return text
}

func handleArticle(msg *client.Message) string {
	return fmt.Sprintf("[%s]", msg.Url)
}
