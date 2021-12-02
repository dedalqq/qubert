package application

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/dedalqq/omg.httpserver"
	"github.com/gorilla/websocket"

	"qubert/resources"
)

type resource struct {
	fs.File
	contentType string
}

func (r resource) ContentType() string {
	return r.contentType
}

var contentTypes = map[string]string{
	".css":  "text/css",
	".js":   "application/javascript",
	".html": "text/html",
	".wasm": "application/wasm",
	".svg":  "image/svg+xml",
	".ttf":  "font/ttf",
}

func newDefaultHandler(rs *resources.Storage) httpserver.Handler {
	return httpserver.Handler{
		Get: func(ctx context.Context, r *http.Request, args []string) interface{} {
			p := r.URL.Path
			if p == "/" {
				p = "index.html"
			}

			f, err := rs.Get(path.Join("content", p))
			if err != nil {
				return httpserver.NewError(http.StatusNotFound, "Not exist")
			}

			return resource{
				File:        f,
				contentType: contentTypes[filepath.Ext(p)],
			}
		},
	}
}

const (
	contextKeySession = "session"
	tokenHeader       = "X-access-token"
)

func setSession(ctx context.Context, sn *Session) context.Context {
	return context.WithValue(ctx, contextKeySession, sn)
}

func getSession(ctx context.Context) *Session {
	if v, ok := ctx.Value(contextKeySession).(*Session); ok {
		return v
	}

	panic("session not set")
}

func newAuthMiddleware(a *Application) httpserver.HandlerMiddleware {
	return func(handler httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx context.Context, req *http.Request, args []string) interface{} {
			token := req.Header.Get(tokenHeader)

			sn := a.sm.sessionByToken(token)
			if sn == nil {
				return httpserver.NewError(http.StatusUnauthorized, "forbidden")
			}

			ctx = setSession(ctx, sn)

			return handler(ctx, req, args)
		}
	}
}

type user struct {
	UserName string `json:"user-name"`
}

func newUserHandler(a *Application) httpserver.Handler {
	return httpserver.Handler{
		Middlewares: []httpserver.HandlerMiddleware{
			newAuthMiddleware(a),
		},

		Get: func(ctx context.Context, req *http.Request, args []string) interface{} {
			sn := getSession(ctx)

			return &user{
				UserName: sn.userName,
			}
		},

		Delete: func(ctx context.Context, req *http.Request, args []string) interface{} {
			return nil
		},
	}
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access-token"`
}

func newLoginHandler(a *Application) httpserver.Handler {
	return httpserver.Handler{
		Post: func(ctx context.Context, req *http.Request, args []string) interface{} {
			reqData := loginRequest{}

			err := json.NewDecoder(req.Body).Decode(&reqData)
			if err != nil {
				return httpserver.NewError(http.StatusUnauthorized, "incorrect login or password")
			}

			if a.authUser(reqData.Login, reqData.Password) {
				sn := a.sm.newSession(reqData.Login)

				return &loginResponse{
					AccessToken: sn.token,
				}
			}

			return httpserver.NewError(http.StatusUnauthorized, "incorrect login or password")
		},
	}
}

type pluginInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Icon  string `json:"icon"`
}

type MainPage struct {
	HostName       string       `json:"host-name"`
	Plugins        []pluginInfo `json:"plugins"`
	HostBadgeColor string       `json:"host-badge-color"`
}

func newMainPageHandler(a *Application) httpserver.Handler {
	return httpserver.Handler{
		Middlewares: []httpserver.HandlerMiddleware{
			newAuthMiddleware(a),
		},

		Get: func(ctx context.Context, req *http.Request, args []string) interface{} {
			data := &MainPage{
				HostName:       getHostName(),
				HostBadgeColor: "#ffffff",
			}

			if a.cfg.HostBadgeColor != "" {
				data.HostBadgeColor = a.cfg.HostBadgeColor
			}

			for _, p := range a.pc.pluginsList() {
				data.Plugins = append(data.Plugins, pluginInfo{
					ID:    p.ID(),
					Title: p.Title(),
					Icon:  p.Icon(),
				})
			}

			return data
		},
	}
}

type actionRequest struct {
	CMD  string          `json:"cmd"`
	Args []string        `json:"args"`
	Data json.RawMessage `json:"data"`
}

func newPluginActionHandler(a *Application) httpserver.Handler {
	return httpserver.Handler{
		Middlewares: []httpserver.HandlerMiddleware{
			newAuthMiddleware(a),
		},

		Post: func(ctx context.Context, req *http.Request, args []string) interface{} {
			pluginID := args[0]
			pluginInstance := a.pc.pluginByID(pluginID)

			requestData := actionRequest{}
			err := json.NewDecoder(req.Body).Decode(&requestData)
			if err != nil {
				return httpserver.NewError(http.StatusInternalServerError, "body parsing error")
			}

			if command, ok := pluginInstance.Actions()[requestData.CMD]; ok {
				return command(requestData.Args, bytes.NewBuffer(requestData.Data))
			}

			return httpserver.NewError(http.StatusNotFound, "action not found")
		},
	}
}

type renderRequest struct {
	Args []string `json:"args"`
}

func newPluginRenderHandler(a *Application) httpserver.Handler {
	return httpserver.Handler{
		Middlewares: []httpserver.HandlerMiddleware{
			newAuthMiddleware(a),
		},

		Post: func(ctx context.Context, req *http.Request, args []string) interface{} {
			pluginID := args[0]
			pluginInstance := a.pc.pluginByID(pluginID)

			data := renderRequest{}

			err := json.NewDecoder(req.Body).Decode(&data)
			if err != nil {
				return httpserver.NewError(http.StatusInternalServerError, "body parsing error")
			}

			return pluginInstance.Render(data.Args)
		},
	}
}

type wsSetLocation struct {
	Module string   `json:"module"`
	Args   []string `json:"args"`
}

func newWebSocketHandler(a *Application) httpserver.Handler {
	const weProtocolHeader = "Sec-WebSocket-Protocol"

	var upgrader = websocket.Upgrader{
		Subprotocols: []string{"a2"},
	}

	return httpserver.Handler{
		StdHandler: func(ctx context.Context, w http.ResponseWriter, r *http.Request, args []string) (ctn bool) {
			data := strings.Split(r.Header.Get(weProtocolHeader), ", ")

			if len(data) < 2 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			token := data[1]

			sn := a.sm.sessionByToken(token)
			if sn == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			c, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				a.log.Error(errors.Wrap(err, "webSocket upgrader failed"))
				return
			}

			client := sn.newClient(token, r.RemoteAddr, c)

			a.log.Info("webSocket: New client connected [%v]", r.RemoteAddr)

			defer func() {
				sn.removeClient(c)
			}()

			go func() {
				<-sn.ctx.Done()
				c.Close()
			}()

			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
						a.log.Error(errors.Wrap(err, "webSocket read error"))
					}

					break
				}

				var res wsSetLocation
				err = json.NewDecoder(bytes.NewReader(message)).Decode(&res)
				if err != nil {
					a.log.Error(errors.Wrap(err, "webSocket parse message failed"))

					continue
				}

				client.setLocation(res.Module, res.Args)
			}

			a.log.Info("webSocket: Client [%v] going away", r.RemoteAddr)

			return
		},
	}
}
