package main

import (
	"bufio"
	"fmt"
	"github.com/skip2/go-qrcode"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strings"
	"wx-cli/client"
	"wx-cli/cmd"
	"wx-cli/helper"
)

func ConsoleQrCode(uuid string) {
	q, _ := qrcode.New("https://login.weixin.qq.com/l/"+uuid, qrcode.Low)
	fmt.Println(q.ToString(true))
}

func ScanCallback(body []byte) {
	log.Println("Waiting Confirm...")
}

func LoginCallback(body []byte) {
	log.Println("Login Succeeded")
}

func MessageHandler(msg *client.Message) {
	if len(msg.Content) == 0 {
		return
	}
	h.StoreMessage(msg)
}

func SyncCheckCallback(resp client.SyncCheckResponse) {
	if !resp.Success() {
		fmt.Println(resp.Error())
	}
}

var app *cli.App
var h *helper.Helper

func mainLoop() {
	for {
		select {
		case <-h.Done():
		default:
			fmt.Print("> ")
			command, err := bufio.NewReader(os.Stdin).ReadString(';')
			if err != nil {
				panic(err)
			}
			command = fmt.Sprintf("# %s", strings.TrimRight(command, ";"))
			execute(command)
		}
	}
}

func execute(command string) {
	args := strings.Split(command, " ")
	if err := app.Run(args); err != nil {
		fmt.Println(err.Error())
	}
}

func main() {
	cfg := &helper.Config{
		StorageFileName: "storage.json",
	}
	h = helper.NewHelper(cfg)
	h.BindSyncCheckCallback(SyncCheckCallback)
	h.BindUUIDCallback(ConsoleQrCode)
	h.BindScanCallBack(ScanCallback)
	h.BindLoginCallBack(LoginCallback)
	h.BindMessageHandler(MessageHandler)

	if err := h.HotLogin(); err != nil {
		fmt.Println(err)
		return
	}

	app = &cli.App{
		Name:            "wx-cli",
		CommandNotFound: cmd.FallbackFunc,
	}
	cmd.Init(h)
	app.Commands = cmd.CliCommands

	username := h.GetCurrentUserName()

	fmt.Println("Welcome,", username)

	mainLoop()
}
