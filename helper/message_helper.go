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
	} else if msg.IsSticker() {
		text = handleSticker(msg)
	} else if msg.IsVideo() {
		text = handleVideo(msg)
	} else if msg.IsSystem() {
		text = handleText(msg)
	}
	return text
}

func handleText(msg *client.Message) string {
	return msg.Content
}

func handlePicture(msg *client.Message) string {
	text := "[Photo]"
	var err error
	resp, err := msg.GetPicture()
	if err != nil {
		return text
	}
	length := resp.ContentLength
	if length == 0 {
		return text
	}
	buf := make([]byte, length)
	buf, err = ioutil.ReadAll(resp.Body)
	if err == nil {
		filename := fmt.Sprintf("%v.jpg", time.Now().UnixNano())
		err := ioutil.WriteFile(filename, buf, 0666)
		if err != nil {
			fmt.Println("write file err:", err)
		}
	}
	return text
}

func handleSticker(msg *client.Message) string {
	text := "[Sticker]"
	var err error
	resp, err := msg.GetPicture()
	if err != nil {
		return text
	}
	length := resp.ContentLength
	if length == 0 {
		return text
	}
	buf := make([]byte, length)
	buf, err = ioutil.ReadAll(resp.Body)
	if err == nil {
		filename := fmt.Sprintf("%v.gif", time.Now().UnixNano())
		err := ioutil.WriteFile(filename, buf, 0666)
		if err != nil {
			fmt.Println("write file err:", err)
		}
	}
	return text
}

func handleVideo(msg *client.Message) string {
	text := "[Video]"
	var err error
	resp, err := msg.GetVideo()
	if err != nil {
		return text
	}
	length := resp.ContentLength
	if length == 0 {
		return text
	}
	buf := make([]byte, length)
	buf, err = ioutil.ReadAll(resp.Body)
	if err == nil {
		filename := fmt.Sprintf("%v.gif", time.Now().UnixNano())
		err := ioutil.WriteFile(filename, buf, 0666)
		if err != nil {
			fmt.Println("write file err:", err)
		}
	}
	return text
}

func handleArticle(msg *client.Message) string {
	return fmt.Sprintf("[%s]", msg.Url)
}
