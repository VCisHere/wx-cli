package helper

import (
	"errors"
	"fmt"
	"wx-cli/client"
	"wx-cli/storage"
	"wx-cli/util"
)

type Helper struct {
	bot   *client.Bot
	self  *client.Self
	cfg   *Config
	to    *client.User
	cache *storage.MsgCache
}

type Config struct {
	StorageFileName string
}

func NewHelper(cfg *Config) *Helper {
	return &Helper{
		bot:   client.NewBot(client.Desktop),
		cfg:   cfg,
		cache: storage.NewMsgCache(),
	}
}

func (h *Helper) BindUUIDCallback(f func(uuid string)) {
	h.bot.UUIDCallback = f
}

func (h *Helper) BindScanCallBack(f func(body []byte)) {
	h.bot.ScanCallBack = f
}

func (h *Helper) BindLoginCallBack(f func(body []byte)) {
	h.bot.LoginCallBack = f
}

func (h *Helper) BindSyncCheckCallback(f func(resp client.SyncCheckResponse)) {
	h.bot.SyncCheckCallback = f
}

func (h *Helper) BindMessageHandler(f func(msg *client.Message)) {
	h.bot.MessageHandler = f
}

func (h *Helper) HotLogin() error {
	reloadStorage := client.NewJsonFileHotReloadStorage(h.cfg.StorageFileName)
	err := h.bot.HotLogin(reloadStorage)
	if err == nil {
		h.self, _ = h.bot.GetCurrentUser()
	}
	return err
}

func (h *Helper) GetCurrentUserName() string {
	return h.self.NickName
}

func (h *Helper) GetName(user *client.User) string {
	if user == nil {
		return "Unknown"
	}
	name := user.RemarkName
	if len(name) == 0 {
		name = user.NickName
	}
	return name
}

func (h *Helper) GetFriendsName() ([]string, error) {
	friends, err := h.self.Friends()
	if err != nil {
		return []string{}, err
	}
	names := make([]string, len(friends))
	for i := range names {
		names[i] = h.GetName(friends[i].User)
	}
	return names, nil
}

func (h *Helper) Block() error {
	return h.bot.Block()
}

func (h *Helper) Done() <-chan struct{} {
	return h.bot.Done()
}

func (h *Helper) StoreMessage(msg *client.Message) {
	h.cache.StoreMessage(msg)
}

func (h *Helper) Messages() storage.Messages {
	return h.cache.Messages()
}

func (h *Helper) MessageToString(msg *client.Message) string {
	var msgType string
	var senderText string
	var receiverText string
	var err error

	sender, err := msg.Sender()
	if err != nil {
		senderText = "[Unknown]"
		fmt.Println(err)
	} else {
		fmt.Println("Sender:", sender.NickName)
	}
	receiver, err := msg.Receiver()
	if err != nil {
		receiverText = "[Unknown]"
		fmt.Println(err)
	} else {
		fmt.Println("Receiver:", receiver.NickName)
	}

	switch msg.Category {
	case client.CategoryUnknown:
		senderText = "[Unknown]"
	case client.CategorySystem:
		msgType = "S"
		senderText = "[System]"
	case client.CategoryFriend:
		msgType = "F"
		senderText = fmt.Sprintf("[%s]->", h.GetName(sender))
		receiverText = fmt.Sprintf("[%s]", h.GetName(receiver))
	case client.CategoryGroup:
		msgType = "G"
		senderInGroup, err := msg.SenderInGroup()
		if err != nil {
			if errors.Is(err, client.ErrMsgIsFromSys) {
				senderText = fmt.Sprintf("[%s][System]", h.GetName(sender))
			} else {
				senderText = fmt.Sprintf("[Unknown][System]")
			}
		} else {
			senderText = fmt.Sprintf("[%s][%s]", h.GetName(receiver), h.GetName(senderInGroup))
		}
	case client.CategoryMP:
		msgType = "P"
		senderText = fmt.Sprintf("[%s]->", h.GetName(sender))
		receiverText = fmt.Sprintf("[%s]", h.GetName(receiver))
	}

	createTime := msg.CreateTime
	timeStr := util.Int64ToTimeString(createTime)
	messageStr := HandleMessage(msg)
	result := fmt.Sprintf("[%s][%s]%s%s:%s", timeStr, msgType, senderText, receiverText, messageStr)
	return fmt.Sprintf("%s", result)
}
