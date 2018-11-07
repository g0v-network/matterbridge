package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/zfjagann/golang-ring"
	"github.com/spf13/viper"
)

type Api struct {
	Server *echo.Echo
	Messages ring.Ring
	sync.RWMutex
	*bridge.Config
}

type ApiMessage struct {
	Text     string `json:"text"`
	Username string `json:"username"`
	UserID   string `json:"userid"`
	Avatar   string `json:"avatar"`
	Gateway  string `json:"gateway"`
}

func New(cfg *bridge.Config) bridge.Bridger {
	e := echo.New()
	b := &Api{Config: cfg, Server: e}
	e.HideBanner = true
	e.HidePort = true
	b.Messages = ring.Ring{}
	if b.GetInt("Buffer") != 0 {
		b.Messages.SetCapacity(b.GetInt("Buffer"))
	}
	if b.GetString("Token") != "" {
		e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
			return key == b.GetString("Token"), nil
		}))
	}

	e.GET("/api/health", b.handleHealthcheck)
	e.PUT("/api/reload", b.handleConfigReload)
	e.GET("/api/messages", b.handleMessages)
	e.GET("/api/stream", b.handleStream)
	e.POST("/api/message", b.handlePostMessage)
	return b
}

func (b *Api) Connect() error {
	go func() {
		if b.GetString("BindAddress") == "" {
			b.Log.Fatalf("No BindAddress configured.")
		}
		b.Log.Infof("Listening on %s", b.GetString("BindAddress"))
		b.Log.Info(b.Server.Start(b.GetString("BindAddress")))
	}()
	return nil
}
func (b *Api) Disconnect() error {
	ctx := context.Background()
	if err := b.Server.Shutdown(ctx); err != nil {
		b.Log.Info(err)
	}
	return nil

}
func (b *Api) JoinChannel(channel config.ChannelInfo) error {
	return nil

}

func (b *Api) Send(msg config.Message) (string, error) {
	b.Lock()
	defer b.Unlock()
	// ignore delete messages
	if msg.Event == config.EVENT_MSG_DELETE {
		return "", nil
	}
	b.Messages.Enqueue(&msg)
	return "", nil
}

func (b *Api) handleHealthcheck(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (b *Api) handleConfigReload(c echo.Context) error {
	cfgURL := b.GetString("ConfigURL")
	if cfgURL == "" {
		b.Log.Warning("Reload API triggered, but no config file url set.")
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}

	b.Log.Debugf("Reloading config from remote file: " + cfgURL)
	_, err := url.ParseRequestURI(cfgURL)
	if err != nil {
		b.Log.Error("Malformed config file url: ", err)
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	res, err := http.Get(cfgURL)
	if err != nil {
		b.Log.Error("Failed to fetch remote config file: ", err)
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	defer res.Body.Close()
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		b.Log.Error("Error reading remote config file: ", err)
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	cfgfile := b.GetConfigFile()
	err = ioutil.WriteFile(cfgfile, content, 0644)
	if err != nil {
		b.Log.Error("Failed to write remote config file: ", err)
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	viper.ReadInConfig()

	b.Remote <- config.Message{Username: "system", Text: config.EVENT_RELOAD_CONFIG, Channel: "api", Account: "", Event: config.EVENT_RELOAD_CONFIG}
	return c.String(http.StatusAccepted, "Accepted")
}

func (b *Api) handlePostMessage(c echo.Context) error {
	message := config.Message{}
	if err := c.Bind(&message); err != nil {
		return err
	}
	// these values are fixed
	message.Channel = "api"
	message.Protocol = "api"
	message.Account = b.Account
	message.ID = ""
	message.Timestamp = time.Now()
	b.Log.Debugf("Sending message from %s on %s to gateway", message.Username, "api")
	b.Remote <- message
	return c.JSON(http.StatusOK, message)
}

func (b *Api) handleMessages(c echo.Context) error {
	b.Lock()
	defer b.Unlock()
	c.JSONPretty(http.StatusOK, b.Messages.Values(), " ")
	b.Messages = ring.Ring{}
	return nil
}

func (b *Api) handleStream(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)
	greet := config.Message{
		Event:config.EVENT_API_CONNECTED,
		Timestamp:time.Now(),
	}
	if err := json.NewEncoder(c.Response()).Encode(greet); err != nil {
		return err
	}
	c.Response().Flush()
	closeNotifier := c.Response().CloseNotify()
	for {
		select {
		case <-closeNotifier:
			return nil
		default:
			msg := b.Messages.Dequeue()
			if msg != nil {
				if err := json.NewEncoder(c.Response()).Encode(msg); err != nil {
					return err
				}
				c.Response().Flush()
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}
