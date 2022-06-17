package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"plugin"
	"sync"
	"syscall"

	"qubert/internal/logger"
	"qubert/pluginTools"
)

type pluginAPI struct {
	moduleID string
	sessions *sessionManager
	wg       *sync.WaitGroup
	ctx      context.Context
	c        *pluginController
}

func (i *pluginAPI) SaveModuleConfig(cfg interface{}) error {
	buf := bytes.NewBuffer([]byte{})

	encoder := json.NewEncoder(buf)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(cfg)
	if err != nil {
		return err
	}

	i.c.settings.set(i.moduleID, buf.Bytes())

	return i.c.saveSettings()
}

func (i *pluginAPI) LoadModuleConfig(cfg interface{}) error {
	err := json.NewDecoder(bytes.NewReader(i.c.settings.get(i.moduleID))).Decode(cfg)
	if err == io.EOF {
		return nil
	}

	return err
}

func (i *pluginAPI) SafeRun(f func()) {
	i.wg.Add(1)
	defer i.wg.Done()

	f()
}

type eventData struct {
	EventType string      `json:"type"`
	Options   interface{} `json:"options,omitempty"`
}

type eventUpdateOptions struct {
	ID      string                  `json:"id"`
	Element pluginTools.ElementType `json:"element"`
	Data    interface{}             `json:"data"`
}

func (i *pluginAPI) Send(data interface{}, args ...string) {
	i.sessions.send(data, i.moduleID, args)
}

func (i *pluginAPI) SendUpdate(updateData pluginTools.Update, args ...string) bool {
	return i.sessions.send(eventData{
		EventType: "update",
		Options: eventUpdateOptions{
			ID:      updateData.ElementID(),
			Element: updateData.UpdateType(),
			Data:    updateData,
		},
	}, i.moduleID, args)
}

func (i *pluginAPI) Reload(args ...string) {
	i.sessions.send(eventData{
		EventType: "reload",
	}, i.moduleID, args)
}

func (i *pluginAPI) Exit() {
	cancelContext(i.ctx)
}

func (i *pluginAPI) Shutdown() error {
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
}

func (i *pluginAPI) Restart() error {
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

func (i *pluginAPI) Version() (string, string) {
	return i.c.version, i.c.commit
}

type iPlugin interface {
	ID() string
	Title() string
	Icon() string
	Run(context.Context, pluginTools.PluginAPI) error
	Actions() pluginTools.ActionsMap
	Render([]string) pluginTools.Page
}

type subRendersPlugin interface {
	SubRenders() []pluginTools.SubPageRender
}

type pluginSettings struct {
	mx sync.Mutex

	Plugins map[string]json.RawMessage `json:"plugins"`
}

func (ps *pluginSettings) set(name string, value []byte) {
	ps.mx.Lock()
	defer ps.mx.Unlock()

	ps.Plugins[name] = value
}

func (ps *pluginSettings) get(name string) []byte {
	ps.mx.Lock()
	defer ps.mx.Unlock()

	return ps.Plugins[name]
}

type pluginController struct {
	mx sync.Mutex

	settingsFile string
	settings     pluginSettings

	sessions *sessionManager

	plugins []iPlugin

	version string
	commit  string
}

func newPluginController(settingsFile string, sessions *sessionManager) (*pluginController, error) {
	c := &pluginController{
		settingsFile: settingsFile,
		sessions:     sessions,
	}

	err := c.loadSettings()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *pluginController) setVersion(version string, commit string) {
	c.version = version
	c.commit = commit
}

func (c *pluginController) loadSettings() error {
	if _, err := os.Stat(c.settingsFile); os.IsNotExist(err) {
		c.settings = pluginSettings{
			Plugins: make(map[string]json.RawMessage),
		}

		return c.saveSettings()
	}

	file, err := os.Open(c.settingsFile)
	if err != nil {
		return err
	}

	defer file.Close()

	return json.NewDecoder(file).Decode(&c.settings)
}

func (c *pluginController) saveSettings() error {
	err := os.MkdirAll(path.Dir(c.settingsFile), 0744)
	if err != nil {
		return err
	}

	file, err := os.Create(c.settingsFile)
	if err != nil {
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(&c.settings)
	if err != nil {
		return err
	}

	return nil
}

func (c *pluginController) initPlugins(ctx context.Context, wg *sync.WaitGroup, log *logger.Logger, pp ...iPlugin) error {
	var err error

	c.mx.Lock()
	defer c.mx.Unlock()

	for _, p := range pp {
		c.plugins = append(c.plugins, p)

		wg.Add(1)
		go func(p iPlugin) {
			defer wg.Done()
			err = p.Run(ctx, &pluginAPI{
				moduleID: p.ID(),
				sessions: c.sessions,
				wg:       wg,
				ctx:      ctx,
				c:        c,
			})

			if err != nil {
				log.Error(err)
			}
		}(p)
	}

	return nil
}

func (c *pluginController) loadExternalPlugins(dir string, log *logger.Logger) ([]iPlugin, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.so"))
	if err != nil {
		return nil, err
	}

	var externalPlugins []iPlugin

	for _, f := range files {
		fmt.Println(f)
		p, err := plugin.Open(f)
		if err != nil {
			log.Error(err)
		}

		i, err := p.Lookup("PluginInstance")
		if err != nil {
			log.Error(err)
		}

		if pl, ok := i.(iPlugin); ok {
			externalPlugins = append(externalPlugins, pl)
		} else {
			log.Error(fmt.Errorf("failed to load [%s] plugin", f))
		}
	}

	return externalPlugins, nil
}

func (c *pluginController) pluginsList() []iPlugin {
	return c.plugins
}

func (c *pluginController) pluginByID(id string) iPlugin {
	for _, p := range c.plugins {
		if p.ID() == id {
			return p
		}
	}

	return nil
}
