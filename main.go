package main

import (
	"fmt"
	"github.com/skip2/go-qrcode"
	"io/ioutil"
	"time"
	"wx-cli/client"
)

func ConsoleQrCode(uuid string) {
	q, _ := qrcode.New("https://login.weixin.qq.com/l/"+uuid, qrcode.Low)
	fmt.Println(q.ToString(true))
}

func MessageHandler(msg *client.Message) {
	if len(msg.Content) == 0 {
		return
	}
	sender, _ := msg.Sender()
	senderName := sender.NickName
	prefix := fmt.Sprintf("[%s]:", senderName)
	var text string
	if msg.IsText() {
		text = msg.Content
	} else if msg.IsPicture() {
		resp, err := msg.GetPicture()
		text = "[Picture]"
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
	}
	fmt.Println(fmt.Sprintf("%s%s", prefix, text))
}

func SyncCheckCallback(resp client.SyncCheckResponse) {
	if !resp.Success() {
		fmt.Println(resp.Error())
	}
}

func main() {
	bot := client.DefaultBot(client.Desktop)

	bot.MessageHandler = MessageHandler
	bot.UUIDCallback = ConsoleQrCode
	bot.SyncCheckCallback = SyncCheckCallback

	reloadStorage := client.NewJsonFileHotReloadStorage("storage.json")

	if err := bot.HotLogin(reloadStorage); err != nil {
		fmt.Println(err)
		return
	}

	self, err := bot.GetCurrentUser()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Welcome,", self.NickName)
	fmt.Println(self.Alias)
	fmt.Println(self.DisplayName)
	_ = bot.Block()
}
