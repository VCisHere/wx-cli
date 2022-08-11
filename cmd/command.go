package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"reflect"
	"strings"
	"wx-cli/helper"
)

const prefix = "Cmd"

var h *helper.Helper
var CliCommands []*cli.Command

type cmdFactory struct{}

func Init(hh *helper.Helper) {
	h = hh
	initCommands()
}

func initCommands() {
	cmdFactoryType := reflect.TypeOf(cmdFactory{})
	cmdFactoryValue := reflect.ValueOf(cmdFactory{})
	for i := 0; i < cmdFactoryType.NumMethod(); i++ {
		method := cmdFactoryType.Method(i)
		methodName := method.Name
		if !strings.HasPrefix(methodName, prefix) {
			continue
		}
		args := []reflect.Value{
			cmdFactoryValue,
		}
		values := method.Func.Call(args)
		c := values[0].Interface().(*cli.Command)
		c.Name = strings.TrimPrefix(methodName, prefix)
		c.Name = strings.ToLower(c.Name)
		CliCommands = append(CliCommands, c)
	}
}

func FallbackFunc(ctx *cli.Context, str string) {
	fmt.Println("error input")
}

func (c cmdFactory) CmdFriends() *cli.Command {
	return &cli.Command{
		Usage:       "Friends",
		Description: "Show friends",
		Action: func(ctx *cli.Context) error {
			fmt.Println(h.GetFriendsName())
			return nil
		},
	}
}

func (c cmdFactory) CmdFetchMembers() *cli.Command {
	return &cli.Command{
		Aliases: []string{
			"fetch",
		},
		Usage:       "FetchMembers",
		Description: "Fetch All Members",
		Action: func(ctx *cli.Context) error {
			fmt.Println("Fetching Members...")
			err := h.FetchMembers()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Fetched", h.MemberCount())
			}
			return nil
		},
	}
}

func (c cmdFactory) CmdAllMessages() *cli.Command {
	return &cli.Command{
		Aliases: []string{
			"am",
		},
		Usage:       "AllMessages",
		Description: "Show All Messages",
		Action: func(ctx *cli.Context) error {
			messages := h.AllMessages()
			for _, msg := range messages {
				text := h.MessageToString(msg)
				fmt.Println(text)
			}
			return nil
		},
	}
}

func (c cmdFactory) CmdMessages() *cli.Command {
	return &cli.Command{
		Aliases: []string{
			"m",
		},
		Usage:       "Messages",
		Description: "Show Unread Messages",
		Action: func(ctx *cli.Context) error {
			messages := h.UnreadMessages()
			for _, msg := range messages {
				text := h.MessageToString(msg)
				fmt.Println(text)
			}
			return nil
		},
	}
}
