package client

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// todo: 更灵活的 error
var (
	ErrMsgIsFromSys = errors.New("can not found sender from system message")
)

type Message struct {
	isAt    bool
	AppInfo struct {
		Type  int
		AppID string
	}
	AppMsgType            AppMessageType
	HasProductId          int
	ImgHeight             int
	ImgStatus             int
	ImgWidth              int
	ForwardFlag           int
	MsgType               MessageType
	Status                int
	StatusNotifyCode      int
	SubMsgType            int
	VoiceLength           int
	CreateTime            int64
	NewMsgId              int64
	PlayLength            int64
	MediaId               string
	MsgId                 string
	EncryFileName         string
	FileName              string
	FileSize              string
	Content               string
	FromUserName          string
	OriContent            string
	StatusNotifyUserName  string
	Ticket                string
	ToUserName            string
	Url                   string
	senderInGroupUserName string
	RecommendInfo         RecommendInfo
	Bot                   *Bot `json:"-"`
	mu                    sync.RWMutex
	Context               context.Context `json:"-"`
	item                  map[string]interface{}
	Raw                   []byte `json:"-"`
	RawContent            string `json:"-"` // 消息原始内容

	MessagePersistence
}

type MessagePersistence struct {
	Category MessageCategory
	FromUin  int64
	ToUin    int64
}

// Sender 获取消息的发送者
func (m *Message) Sender() (*User, error) {
	if m.FromUserName == m.Bot.self.User.UserName {
		return m.Bot.self.User, nil
	}
	user := &User{UserName: m.FromUserName}
	err := user.Detail(m.Bot.self)
	return user, err
}

// SenderInGroup 获取消息在群里面的发送者
func (m *Message) SenderInGroup() (*User, error) {
	if m.IsSystem() {
		// 判断是否有自己发送
		if m.FromUserName == m.Bot.self.User.UserName {
			return m.Bot.self.User, nil
		}
		return nil, ErrMsgIsFromSys
	}
	group, err := m.Sender()
	if err != nil {
		return nil, err
	}
	if err := group.Detail(m.Bot.self); err != nil {
		return nil, err
	}
	if group.IsFriend() {
		return group, nil
	}
	users := group.MemberList.SearchByUserName(1, m.senderInGroupUserName)
	if users == nil {
		return nil, ErrNoSuchUserFoundError
	}
	users.init()
	return users.First(), nil
}

// Receiver 获取消息的接收者
// 如果消息是群组消息，则返回群组
// 如果消息是好友消息，则返回好友
// 如果消息是系统消息，则返回当前用户
func (m *Message) Receiver() (*User, error) {
	if m.ToUserName == m.Bot.self.UserName {
		return m.Bot.self.User, nil
	}
	if m.Category == CategorySystem {
		return m.Bot.self.User, nil
	}
	username := m.ToUserName
	user, ok := m.Bot.self.FindContactByUserName(username)
	if ok {
		return user, nil
	}

	if m.Category == CategoryGroup {
		groups, err := m.Bot.self.Groups()
		if err != nil {
			return nil, err
		}
		users := groups.SearchByUserName(1, username)
		if users.Count() > 0 {
			return users.First().User, nil
		}
	}

	user, exist := m.Bot.self.MemberList.GetByUserName(m.ToUserName)
	if exist {
		return user, nil
	}

	user = &User{UserName: m.ToUserName}
	err := user.Detail(m.Bot.self)
	return user, err
}

// IsSendBySelf 判断消息是否由自己发送
func (m *Message) IsSendBySelf() bool {
	return m.FromUserName == m.Bot.self.User.UserName
}

// ReplyText 回复文本消息
func (m *Message) ReplyText(content string) (*SentMessage, error) {
	msg := NewSendMessage(MsgTypeText, content, m.Bot.self.User.UserName, m.FromUserName, "")
	info := m.Bot.Storage.LoginInfo
	request := m.Bot.Storage.Request
	sentMessage, err := m.Bot.Caller.WebWxSendMsg(msg, info, request)
	return m.Bot.self.sendMessageWrapper(sentMessage, err)
}

// ReplyImage 回复图片消息
func (m *Message) ReplyImage(file *os.File) (*SentMessage, error) {
	info := m.Bot.Storage.LoginInfo
	request := m.Bot.Storage.Request
	sentMessage, err := m.Bot.Caller.WebWxSendImageMsg(file, request, info, m.Bot.self.UserName, m.FromUserName)
	return m.Bot.self.sendMessageWrapper(sentMessage, err)
}

// ReplyVideo 回复视频消息
func (m *Message) ReplyVideo(file *os.File) (*SentMessage, error) {
	info := m.Bot.Storage.LoginInfo
	request := m.Bot.Storage.Request
	sentMessage, err := m.Bot.Caller.WebWxSendVideoMsg(file, request, info, m.Bot.self.UserName, m.FromUserName)
	return m.Bot.self.sendMessageWrapper(sentMessage, err)
}

// ReplyFile 回复文件消息
func (m *Message) ReplyFile(file *os.File) (*SentMessage, error) {
	info := m.Bot.Storage.LoginInfo
	request := m.Bot.Storage.Request
	sentMessage, err := m.Bot.Caller.WebWxSendFile(file, request, info, m.Bot.self.UserName, m.FromUserName)
	return m.Bot.self.sendMessageWrapper(sentMessage, err)
}

func (m *Message) IsText() bool {
	return m.MsgType == MsgTypeText && m.Url == ""
}

func (m *Message) IsMap() bool {
	return m.MsgType == MsgTypeText && m.Url != ""
}

func (m *Message) IsPicture() bool {
	return m.MsgType == MsgTypeImage
}

func (m *Message) IsSticker() bool {
	return m.MsgType == MsgTypeSticker
}

func (m *Message) IsVoice() bool {
	return m.MsgType == MsgTypeVoice
}

func (m *Message) IsFriendAdd() bool {
	return m.MsgType == MsgTypeVerify && m.FromUserName == "fmessage"
}

func (m *Message) IsCard() bool {
	return m.MsgType == MsgTypeShareCard
}

func (m *Message) IsVideo() bool {
	return m.MsgType == MsgTypeVideo || m.MsgType == MsgTypeMicroVideo
}

func (m *Message) IsMedia() bool {
	return m.MsgType == MsgTypeApp
}

// IsRecalled 判断是否撤回
func (m *Message) IsRecalled() bool {
	return m.MsgType == MsgTypeRecalled
}

func (m *Message) IsSystem() bool {
	return m.MsgType == MsgTypeSys
}

func (m *Message) IsNotify() bool {
	return m.MsgType == 51 && m.StatusNotifyCode != 0
}

// IsTransferAccounts 判断当前的消息是不是微信转账
func (m *Message) IsTransferAccounts() bool {
	return m.IsMedia() && m.FileName == "微信转账"
}

// IsSendRedPacket 否发出红包判断当前是
func (m *Message) IsSendRedPacket() bool {
	return m.IsSystem() && m.Content == "发出红包，请在手机上查看"
}

// IsReceiveRedPacket 判断当前是否收到红包
func (m *Message) IsReceiveRedPacket() bool {
	return m.IsSystem() && m.Content == "收到红包，请在手机上查看"
}

func (m *Message) IsSysNotice() bool {
	return m.MsgType == 9999
}

// IsStatusNotify 判断是否为操作通知消息
func (m *Message) IsStatusNotify() bool {
	return m.MsgType == MsgTypeWxInit
}

// HasFile 判断消息是否为文件类型的消息
func (m *Message) HasFile() bool {
	return m.IsPicture() || m.IsVoice() || m.IsVideo() || (m.IsMedia() && m.AppMsgType == AppMsgTypeAttach) || m.IsSticker()
}

// GetFile 获取文件消息的文件
func (m *Message) GetFile() (*http.Response, error) {
	if !m.HasFile() {
		return nil, errors.New("invalid message type")
	}
	if m.IsPicture() || m.IsSticker() {
		return m.Bot.Caller.Client.WebWxGetMsgImg(m, m.Bot.Storage.LoginInfo)
	}
	if m.IsVoice() {
		return m.Bot.Caller.Client.WebWxGetVoice(m, m.Bot.Storage.LoginInfo)
	}
	if m.IsVideo() {
		return m.Bot.Caller.Client.WebWxGetVideo(m, m.Bot.Storage.LoginInfo)
	}
	if m.IsMedia() {
		return m.Bot.Caller.Client.WebWxGetMedia(m, m.Bot.Storage.LoginInfo)
	}
	return nil, errors.New("unsupported type")
}

// GetPicture 获取图片消息的响应
func (m *Message) GetPicture() (*http.Response, error) {
	if !(m.IsPicture() || m.IsSticker()) {
		return nil, errors.New("picture message required")
	}
	return m.Bot.Caller.Client.WebWxGetMsgImg(m, m.Bot.Storage.LoginInfo)
}

// GetVoice 获取录音消息的响应
func (m *Message) GetVoice() (*http.Response, error) {
	if !m.IsVoice() {
		return nil, errors.New("voice message required")
	}
	return m.Bot.Caller.Client.WebWxGetVoice(m, m.Bot.Storage.LoginInfo)
}

// GetVideo 获取视频消息的响应
func (m *Message) GetVideo() (*http.Response, error) {
	if !m.IsVideo() {
		return nil, errors.New("video message required")
	}
	return m.Bot.Caller.Client.WebWxGetVideo(m, m.Bot.Storage.LoginInfo)
}

// GetMedia 获取媒体消息的响应
func (m *Message) GetMedia() (*http.Response, error) {
	if !m.IsMedia() {
		return nil, errors.New("media message required")
	}
	return m.Bot.Caller.Client.WebWxGetMedia(m, m.Bot.Storage.LoginInfo)
}

// Card 获取card类型
func (m *Message) Card() (*Card, error) {
	if !m.IsCard() {
		return nil, errors.New("card message required")
	}
	var card Card
	err := xml.Unmarshal(stringToByte(m.Content), &card)
	return &card, err
}

// FriendAddMessageContent 获取FriendAddMessageContent内容
func (m *Message) FriendAddMessageContent() (*FriendAddMessage, error) {
	if !m.IsFriendAdd() {
		return nil, errors.New("friend add message required")
	}
	var f FriendAddMessage
	err := xml.Unmarshal(stringToByte(m.Content), &f)
	return &f, err
}

// RevokeMsg 获取撤回消息的内容
func (m *Message) RevokeMsg() (*RevokeMsg, error) {
	if !m.IsRecalled() {
		return nil, errors.New("recalled message required")
	}
	var r RevokeMsg
	err := xml.Unmarshal(stringToByte(m.Content), &r)
	return &r, err
}

// Agree 同意好友的请求
func (m *Message) Agree(verifyContents ...string) error {
	if !m.IsFriendAdd() {
		return fmt.Errorf("friend add message required")
	}
	return m.Bot.Caller.WebWxVerifyUser(m.Bot.Storage, m.RecommendInfo, strings.Join(verifyContents, ""))
}

// AsRead 将消息设置为已读
func (m *Message) AsRead() error {
	return m.Bot.Caller.WebWxStatusAsRead(m.Bot.Storage.Request, m.Bot.Storage.LoginInfo, m)
}

// IsArticle 判断当前的消息类型是否为文章
func (m *Message) IsArticle() bool {
	return m.AppMsgType == AppMsgTypeUrl
}

// MediaData 获取当前App Message的具体内容
func (m *Message) MediaData() (*AppMessageData, error) {
	if !m.IsMedia() {
		return nil, errors.New("media message required")
	}
	var data AppMessageData
	if err := xml.Unmarshal(stringToByte(m.Content), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// Set 往消息上下文中设置值
// goroutine safe
func (m *Message) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.item == nil {
		m.item = make(map[string]interface{})
	}
	m.item[key] = value
}

// Get 从消息上下文中获取值
// goroutine safe
func (m *Message) Get(key string) (value interface{}, exist bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exist = m.item[key]
	return
}

func (m *Message) initPersistence() {
	sender, err := m.Sender()
	if err == nil {
		m.FromUin = sender.Uin
	}
	receiver, err := m.Receiver()
	if err == nil {
		m.ToUin = receiver.Uin
	}

	m.initMessageCategory(sender, receiver)

}

func (m *Message) initMessageCategory(sender *User, receiver *User) {
	if strings.HasPrefix(m.FromUserName, "@@") || strings.HasPrefix(m.ToUserName, "@@") {
		m.Category = CategoryGroup
		return
	}
	if m.IsSystem() {
		m.Category = CategorySystem
		return
	}
	if sender == nil || receiver == nil {
		m.Category = CategoryUnknown
		return
	}
	if sender.IsMP() || receiver.IsMP() {
		m.Category = CategoryMP
		return
	}
	m.Category = CategoryFriend
}

// 消息初始化,根据不同的消息作出不同的处理
func (m *Message) init(bot *Bot) {
	m.Bot = bot
	raw, _ := json.Marshal(m)
	m.Raw = raw
	m.RawContent = m.Content
	// 如果是群消息
	if strings.HasPrefix(m.FromUserName, "@@") || strings.HasPrefix(m.ToUserName, "@@") {
		if !m.IsSystem() {
			// 将Username和正文分开
			if !m.IsSendBySelf() {
				data := strings.Split(m.Content, ":<br/>")
				m.Content = strings.Join(data[1:], "")
				m.senderInGroupUserName = data[0]
				if strings.Contains(m.Content, "@") {
					sender, err := m.Sender()
					if err == nil {
						receiver := sender.MemberList.SearchByUserName(1, m.ToUserName)
						if receiver != nil {
							displayName := receiver.First().DisplayName
							if displayName == "" {
								displayName = receiver.First().NickName
							}
							var atFlag string
							if strings.Contains(m.Content, "\u2005") {
								atFlag = "@" + displayName + "\u2005"
							} else {
								atFlag = "@" + displayName + " "
							}
							m.isAt = strings.Contains(m.Content, atFlag) || strings.HasSuffix(m.Content, atFlag)
						}
					}
				}
			} else {
				// 这块不严谨，但是只能这么干了
				m.isAt = strings.Contains(m.Content, "@") || strings.Contains(m.Content, "\u2005")
			}
		}
	}
	// 处理消息中的换行
	m.Content = strings.Replace(m.Content, `<br/>`, "\n", -1)
	// 处理html转义字符
	m.Content = html.UnescapeString(m.Content)
	// 处理消息中的emoji表情
	m.Content = FormatEmoji(m.Content)

	m.initPersistence()
}

// SendMessage 发送消息的结构体
type SendMessage struct {
	Type         MessageType
	Content      string
	FromUserName string
	ToUserName   string
	LocalID      string
	ClientMsgId  string
	MediaId      string `json:"MediaId,omitempty"`
}

// NewSendMessage SendMessage的构造方法
func NewSendMessage(msgType MessageType, content, fromUserName, toUserName, mediaId string) *SendMessage {
	id := strconv.FormatInt(time.Now().UnixNano()/1e2, 10)
	return &SendMessage{
		Type:         msgType,
		Content:      content,
		FromUserName: fromUserName,
		ToUserName:   toUserName,
		LocalID:      id,
		ClientMsgId:  id,
		MediaId:      mediaId,
	}
}

// NewTextSendMessage 文本消息的构造方法
func NewTextSendMessage(content, fromUserName, toUserName string) *SendMessage {
	return NewSendMessage(MsgTypeText, content, fromUserName, toUserName, "")
}

// NewMediaSendMessage 媒体消息的构造方法
func NewMediaSendMessage(msgType MessageType, fromUserName, toUserName, mediaId string) *SendMessage {
	return NewSendMessage(msgType, "", fromUserName, toUserName, mediaId)
}

// RecommendInfo 一些特殊类型的消息会携带该结构体信息
type RecommendInfo struct {
	OpCode     int
	Scene      int
	Sex        int
	VerifyFlag int
	AttrStatus int64
	QQNum      int64
	Alias      string
	City       string
	Content    string
	NickName   string
	Province   string
	Signature  string
	Ticket     string
	UserName   string
}

// Card 名片消息内容
type Card struct {
	XMLName                 xml.Name `xml:"msg"`
	ImageStatus             int      `xml:"imagestatus,attr"`
	Scene                   int      `xml:"scene,attr"`
	Sex                     int      `xml:"sex,attr"`
	Certflag                int      `xml:"certflag,attr"`
	BigHeadImgUrl           string   `xml:"bigheadimgurl,attr"`
	SmallHeadImgUrl         string   `xml:"smallheadimgurl,attr"`
	UserName                string   `xml:"username,attr"`
	NickName                string   `xml:"nickname,attr"`
	ShortPy                 string   `xml:"shortpy,attr"`
	Alias                   string   `xml:"alias,attr"` // Note: 这个是名片用户的微信号
	Province                string   `xml:"province,attr"`
	City                    string   `xml:"city,attr"`
	Sign                    string   `xml:"sign,attr"`
	Certinfo                string   `xml:"certinfo,attr"`
	BrandIconUrl            string   `xml:"brandIconUrl,attr"`
	BrandHomeUr             string   `xml:"brandHomeUr,attr"`
	BrandSubscriptConfigUrl string   `xml:"brandSubscriptConfigUrl,attr"`
	BrandFlags              string   `xml:"brandFlags,attr"`
	RegionCode              string   `xml:"regionCode,attr"`
}

// FriendAddMessage 好友添加消息信息内容
type FriendAddMessage struct {
	XMLName           xml.Name `xml:"msg"`
	Shortpy           int      `xml:"shortpy,attr"`
	ImageStatus       int      `xml:"imagestatus,attr"`
	Scene             int      `xml:"scene,attr"`
	PerCard           int      `xml:"percard,attr"`
	Sex               int      `xml:"sex,attr"`
	AlbumFlag         int      `xml:"albumflag,attr"`
	AlbumStyle        int      `xml:"albumstyle,attr"`
	SnsFlag           int      `xml:"snsflag,attr"`
	Opcode            int      `xml:"opcode,attr"`
	FromUserName      string   `xml:"fromusername,attr"`
	EncryptUserName   string   `xml:"encryptusername,attr"`
	FromNickName      string   `xml:"fromnickname,attr"`
	Content           string   `xml:"content,attr"`
	Country           string   `xml:"country,attr"`
	Province          string   `xml:"province,attr"`
	City              string   `xml:"city,attr"`
	Sign              string   `xml:"sign,attr"`
	Alias             string   `xml:"alias,attr"`
	WeiBo             string   `xml:"weibo,attr"`
	AlbumBgImgId      string   `xml:"albumbgimgid,attr"`
	SnsBgImgId        string   `xml:"snsbgimgid,attr"`
	SnsBgObjectId     string   `xml:"snsbgobjectid,attr"`
	MHash             string   `xml:"mhash,attr"`
	MFullHash         string   `xml:"mfullhash,attr"`
	BigHeadImgUrl     string   `xml:"bigheadimgurl,attr"`
	SmallHeadImgUrl   string   `xml:"smallheadimgurl,attr"`
	Ticket            string   `xml:"ticket,attr"`
	GoogleContact     string   `xml:"googlecontact,attr"`
	QrTicket          string   `xml:"qrticket,attr"`
	ChatRoomUserName  string   `xml:"chatroomusername,attr"`
	SourceUserName    string   `xml:"sourceusername,attr"`
	ShareCardUserName string   `xml:"sharecardusername,attr"`
	ShareCardNickName string   `xml:"sharecardnickname,attr"`
	CardVersion       string   `xml:"cardversion,attr"`
	BrandList         struct {
		Count int   `xml:"count,attr"`
		Ver   int64 `xml:"ver,attr"`
	} `xml:"brandlist"`
}

// RevokeMsg 撤回消息Content
type RevokeMsg struct {
	SysMsg    xml.Name `xml:"sysmsg"`
	Type      string   `xml:"type,attr"`
	RevokeMsg struct {
		OldMsgId   int64  `xml:"oldmsgid"`
		MsgId      int64  `xml:"msgid"`
		Session    string `xml:"session"`
		ReplaceMsg string `xml:"replacemsg"`
	} `xml:"revokemsg"`
}

// SentMessage 已发送的信息
type SentMessage struct {
	*SendMessage
	Self  *Self
	MsgId string
}

// Revoke 撤回该消息
func (s *SentMessage) Revoke() error {
	return s.Self.RevokeMessage(s)
}

// CanRevoke 是否可以撤回该消息
func (s *SentMessage) CanRevoke() bool {
	i, err := strconv.ParseInt(s.ClientMsgId, 10, 64)
	if err != nil {
		return false
	}
	start := time.Unix(i/10000000, 0)
	return time.Now().Sub(start) < time.Minute*2
}

// ForwardToFriends 转发该消息给好友
func (s *SentMessage) ForwardToFriends(friends ...*Friend) error {
	return s.Self.ForwardMessageToFriends(s, friends...)
}

// ForwardToGroups 转发该消息给群组
func (s *SentMessage) ForwardToGroups(groups ...*Group) error {
	return s.Self.ForwardMessageToGroups(s, groups...)
}

type appmsg struct {
	Type      int    `xml:"type"`
	AppId     string `xml:"appid,attr"` // wxeb7ec651dd0aefa9
	SdkVer    string `xml:"sdkver,attr"`
	Title     string `xml:"title"`
	Des       string `xml:"des"`
	Action    string `xml:"action"`
	Content   string `xml:"content"`
	Url       string `xml:"url"`
	LowUrl    string `xml:"lowurl"`
	ExtInfo   string `xml:"extinfo"`
	AppAttach struct {
		TotalLen int64  `xml:"totallen"`
		AttachId string `xml:"attachid"`
		FileExt  string `xml:"fileext"`
	} `xml:"appattach"`
}

func (f appmsg) XmlByte() ([]byte, error) {
	return xml.Marshal(f)
}

func NewFileAppMessage(stat os.FileInfo, attachId string) *appmsg {
	m := &appmsg{AppId: appMessageAppId, Title: stat.Name()}
	m.AppAttach.AttachId = attachId
	m.AppAttach.TotalLen = stat.Size()
	m.Type = 6
	m.AppAttach.FileExt = getFileExt(stat.Name())
	return m
}

// AppMessageData 获取APP消息的正文
// See https://github.com/eatmoreapple/openwechat/issues/62
type AppMessageData struct {
	XMLName xml.Name `xml:"msg"`
	AppMsg  struct {
		Appid             string         `xml:"appid,attr"`
		SdkVer            string         `xml:"sdkver,attr"`
		Title             string         `xml:"title"`
		Des               string         `xml:"des"`
		Action            string         `xml:"action"`
		Type              AppMessageType `xml:"type"`
		ShowType          string         `xml:"showtype"`
		Content           string         `xml:"content"`
		URL               string         `xml:"url"`
		DataUrl           string         `xml:"dataurl"`
		LowUrl            string         `xml:"lowurl"`
		LowDataUrl        string         `xml:"lowdataurl"`
		RecordItem        string         `xml:"recorditem"`
		ThumbUrl          string         `xml:"thumburl"`
		MessageAction     string         `xml:"messageaction"`
		Md5               string         `xml:"md5"`
		ExtInfo           string         `xml:"extinfo"`
		SourceUsername    string         `xml:"sourceusername"`
		SourceDisplayName string         `xml:"sourcedisplayname"`
		CommentUrl        string         `xml:"commenturl"`
		AppAttach         struct {
			TotalLen          string `xml:"totallen"`
			AttachId          string `xml:"attachid"`
			EmoticonMd5       string `xml:"emoticonmd5"`
			FileExt           string `xml:"fileext"`
			FileUploadToken   string `xml:"fileuploadtoken"`
			OverwriteNewMsgId string `xml:"overwrite_newmsgid"`
			FileKey           string `xml:"filekey"`
			CdnAttachUrl      string `xml:"cdnattachurl"`
			AesKey            string `xml:"aeskey"`
			EncryVer          string `xml:"encryver"`
		} `xml:"appattach"`
		WeAppInfo struct {
			PagePath       string `xml:"pagepath"`
			Username       string `xml:"username"`
			Appid          string `xml:"appid"`
			AppServiceType string `xml:"appservicetype"`
		} `xml:"weappinfo"`
		WebSearch string `xml:"websearch"`
	} `xml:"appmsg"`
	FromUsername string `xml:"fromusername"`
	Scene        string `xml:"scene"`
	AppInfo      struct {
		Version string `xml:"version"`
		AppName string `xml:"appname"`
	} `xml:"appinfo"`
	CommentUrl string `xml:"commenturl"`
}

// IsFromApplet 判断当前的消息类型是否来自小程序
func (a *AppMessageData) IsFromApplet() bool {
	return a.AppMsg.Appid != ""
}

// IsArticle 判断当前的消息类型是否为文章
func (a *AppMessageData) IsArticle() bool {
	return a.AppMsg.Type == AppMsgTypeUrl
}

// IsFile 判断当前的消息类型是否为文件
func (a AppMessageData) IsFile() bool {
	return a.AppMsg.Type == AppMsgTypeAttach
}

func (m *Message) String() string {
	return fmt.Sprintf("<%s:%s>", m.MsgType, m.MsgId)
}

// IsAt 判断消息是否为@消息
func (m *Message) IsAt() bool {
	return m.isAt
}
