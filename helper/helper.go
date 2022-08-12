package helper

import (
	"errors"
	"fmt"
	"strconv"
	"wx-cli/client"
	"wx-cli/storage"
	"wx-cli/util"
)

type Helper struct {
	bot   *client.Bot
	self  *client.Self
	cfg   *Config
	to    *client.User
	cache *storage.Cache
}

type Config struct {
	StorageFileName string
}

func NewHelper(cfg *Config) *Helper {
	return &Helper{
		bot: client.NewBot(client.Desktop),
		cfg: cfg,
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
	var err error
	reloadStorage := client.NewJsonFileHotReloadStorage(h.cfg.StorageFileName)
	err = h.bot.HotLogin(reloadStorage)
	if err != nil {
		return err
	}

	h.self, _ = h.bot.GetCurrentUser()
	uin := h.bot.Storage.Response.User.Uin
	filePath := fmt.Sprintf("%s/%s", util.GetCurrentPath(), strconv.FormatInt(uin, 10))
	h.cache, err = storage.NewCacheFromFile(filePath)
	if err == nil {
		return nil
	}
	h.cache = storage.NewCache(filePath)
	return nil
}

func (h *Helper) FetchMembers() error {
	_, err := h.self.Members(true)
	return err
}

func (h *Helper) MemberCount() int {
	members, err := h.self.Members(false)
	if err != nil {
		return 0
	}
	return len(members)
}

func (h *Helper) GetCurrentUserName() string {
	return h.self.NickName
}

func (h *Helper) GetName(user *client.User) string {
	if user == nil {
		return "Unknown"
	}
	if len(user.DisplayName) > 0 {
		return user.DisplayName
	}
	if len(user.RemarkName) > 0 {
		return user.RemarkName
	}
	return user.NickName
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

func (h *Helper) AllMessages() storage.Messages {
	return h.cache.AllMessages()
}

func (h *Helper) UnreadMessages() storage.Messages {
	return h.cache.UnreadMessages()
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
	}
	receiver, err := msg.Receiver()
	if err != nil {
		receiverText = "[Unknown]"
		fmt.Println(err)
	}

	switch msg.Category {
	case client.CategoryUnknown:
		msgType = "[U]"
	case client.CategorySystem:
		msgType = "[S]"
		senderText = "[System]"
	case client.CategoryFriend:
		msgType = "[F]"
		senderText = fmt.Sprintf("[%s]->", h.GetName(sender))
		receiverText = fmt.Sprintf("[%s]", h.GetName(receiver))
	case client.CategoryGroup:
		msgType = "[G]"
		senderInGroup, err := msg.SenderInGroup()
		if err != nil {
			if errors.Is(err, client.ErrMsgIsFromSys) {
				senderText = fmt.Sprintf("[%s][System]", h.GetName(sender))
			} else {
				senderText = fmt.Sprintf("[Unknown][System]")
			}
		} else {
			if msg.IsSendBySelf() {
				senderText = fmt.Sprintf("[%s][%s]", h.GetName(receiver), h.GetName(senderInGroup))
			} else {
				senderText = fmt.Sprintf("[%s][%s]", h.GetName(sender), h.GetName(senderInGroup))
			}
		}
	case client.CategoryMP:
		msgType = "[P]"
		senderText = fmt.Sprintf("[%s]->", h.GetName(sender))
		receiverText = fmt.Sprintf("[%s]", h.GetName(receiver))
	}

	createTime := msg.CreateTime
	timeStr := util.Int64ToTimeString(createTime)
	messageStr := HandleMessage(msg)
	result := fmt.Sprintf("[%s]%s%s%s:%s", timeStr, msgType, senderText, receiverText, messageStr)
	return fmt.Sprintf("%s", result)
}
