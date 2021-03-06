package btelegram

import (
	"html"
	"strconv"
	"strings"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/42wim/matterbridge/bridge/helper"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	unknownUser = "unknown"
	HTMLFormat  = "HTML"
	HTMLNick    = "htmlnick"
)

type Btelegram struct {
	c *tgbotapi.BotAPI
	*bridge.Config
	avatarMap map[string]string // keep cache of userid and avatar sha
}

func New(cfg *bridge.Config) bridge.Bridger {
	return &Btelegram{Config: cfg, avatarMap: make(map[string]string)}
}

func (b *Btelegram) Connect() error {
	var err error
	b.Log.Info("Connecting")
	b.c, err = tgbotapi.NewBotAPI(b.GetString("Token"))
	if err != nil {
		b.Log.Debugf("%#v", err)
		return err
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.c.GetUpdatesChan(u)
	if err != nil {
		b.Log.Debugf("%#v", err)
		return err
	}
	b.Log.Info("Connection succeeded")
	go b.handleRecv(updates)
	return nil
}

func (b *Btelegram) Disconnect() error {
	return nil
}

func (b *Btelegram) JoinChannel(channel config.ChannelInfo) error {
	return nil
}

func (b *Btelegram) Send(msg config.Message) (string, error) {
	b.Log.Debugf("=> Receiving %#v", msg)

	// get the chatid
	chatid, err := strconv.ParseInt(msg.Channel, 10, 64)
	if err != nil {
		return "", err
	}

	// map the file SHA to our user (caches the avatar)
	if msg.Event == config.EventAvatarDownload {
		return b.cacheAvatar(&msg)
	}

	if b.GetString("MessageFormat") == HTMLFormat {
		msg.Text = makeHTML(msg.Text)
	}

	// Delete message
	if msg.Event == config.EventMsgDelete {
		if msg.ID == "" {
			return "", nil
		}
		msgid, err := strconv.Atoi(msg.ID)
		if err != nil {
			return "", err
		}
		_, err = b.c.DeleteMessage(tgbotapi.DeleteMessageConfig{ChatID: chatid, MessageID: msgid})
		return "", err
	}

	// Upload a file if it exists
	if msg.Extra != nil {
		for _, rmsg := range helper.HandleExtra(&msg, b.General) {
			if _, err := b.sendMessage(chatid, rmsg.Username, rmsg.Text); err != nil {
				b.Log.Errorf("sendMessage failed: %s", err)
			}
		}
		// check if we have files to upload (from slack, telegram or mattermost)
		if len(msg.Extra["file"]) > 0 {
			b.handleUploadFile(&msg, chatid)
		}
	}

	// edit the message if we have a msg ID
	if msg.ID != "" {
		msgid, err := strconv.Atoi(msg.ID)
		if err != nil {
			return "", err
		}
		if strings.ToLower(b.GetString("MessageFormat")) == HTMLNick {
			b.Log.Debug("Using mode HTML - nick only")
			msg.Text = html.EscapeString(msg.Text)
		}
		m := tgbotapi.NewEditMessageText(chatid, msgid, msg.Username+msg.Text)
		if b.GetString("MessageFormat") == HTMLFormat {
			b.Log.Debug("Using mode HTML")
			m.ParseMode = tgbotapi.ModeHTML
		}
		if b.GetString("MessageFormat") == "Markdown" {
			b.Log.Debug("Using mode markdown")
			m.ParseMode = tgbotapi.ModeMarkdown
		}
		if strings.ToLower(b.GetString("MessageFormat")) == HTMLNick {
			b.Log.Debug("Using mode HTML - nick only")
			m.ParseMode = tgbotapi.ModeHTML
		}
		_, err = b.c.Send(m)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	// Post normal message
	return b.sendMessage(chatid, msg.Username, msg.Text)
}

func (b *Btelegram) getFileDirectURL(id string) string {
	res, err := b.c.GetFileDirectURL(id)
	if err != nil {
		return ""
	}
	return res
}

func (b *Btelegram) sendMessage(chatid int64, username, text string) (string, error) {
	m := tgbotapi.NewMessage(chatid, "")
	m.Text = username + text
	if b.GetString("MessageFormat") == HTMLFormat {
		b.Log.Debug("Using mode HTML")
		m.ParseMode = tgbotapi.ModeHTML
	}
	if b.GetString("MessageFormat") == "Markdown" {
		b.Log.Debug("Using mode markdown")
		m.ParseMode = tgbotapi.ModeMarkdown
	}
	if strings.ToLower(b.GetString("MessageFormat")) == HTMLNick {
		b.Log.Debug("Using mode HTML - nick only")
		m.Text = username + html.EscapeString(text)
		m.ParseMode = tgbotapi.ModeHTML
	}
	res, err := b.c.Send(m)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(res.MessageID), nil
}

func (b *Btelegram) cacheAvatar(msg *config.Message) (string, error) {
	fi := msg.Extra["file"][0].(config.FileInfo)
	/* if we have a sha we have successfully uploaded the file to the media server,
	so we can now cache the sha */
	if fi.SHA != "" {
		b.Log.Debugf("Added %s to %s in avatarMap", fi.SHA, msg.UserID)
		b.avatarMap[msg.UserID] = fi.SHA
	}
	return "", nil
}
