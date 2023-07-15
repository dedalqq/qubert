package systemd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/pkg/errors"

	. "qubert/pluginTools"
)

type PluginSettings struct {
	PinedServices []string `json:"pined-services"`
}

type Plugin struct {
	api      PluginAPI
	ctx      context.Context
	settings PluginSettings

	dbusConn    *dbus.Conn
	systendChan chan string

	statusChan <-chan map[string]*dbus.UnitStatus
	errChan    <-chan error
}

func (p *Plugin) ID() string {
	return "systemd"
}

func (p *Plugin) Title() string {
	return "Systemd"
}

func (p *Plugin) Icon() string {
	return "lightning-charge"
}

func (p *Plugin) Run(ctx context.Context, api PluginAPI) error {
	p.api = api
	p.ctx = ctx

	p.systendChan = make(chan string, 10)

	go func() {
		for {
			select {
			case <-p.statusChan:
				p.api.Reload()
			case e := <-p.errChan:
				fmt.Println(e)
			case <-p.systendChan: // TODO send alert if not done
				//fmt.Println("q1", q)
				//p.api.Reload()
				//fmt.Println("q2")
			case <-ctx.Done():
				return
			}
		}
	}()

	err := p.api.LoadModuleConfig(&p.settings)
	if err != nil {
		return err
	}

	p.dbusConn, err = dbus.NewWithContext(ctx)
	if err != nil {
		return err
	}

	p.statusChan, p.errChan = p.dbusConn.SubscribeUnits(time.Second / 10)

	return nil
}

func (p *Plugin) Actions() ActionsMap {
	return ActionsMap{
		"pin": func(args []string, data io.Reader) ActionResult {
			serviceName := args[0]

			for _, ps := range p.settings.PinedServices {
				if ps == serviceName {
					return NewReloadActionResult()
				}
			}

			p.settings.PinedServices = append(p.settings.PinedServices, serviceName)
			err := p.api.SaveModuleConfig(&p.settings)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			return NewReloadActionResult()
		},

		"unpin": func(args []string, data io.Reader) ActionResult {
			serviceName := args[0]

			for i, ps := range p.settings.PinedServices {
				if ps == serviceName {
					p.settings.PinedServices = append(p.settings.PinedServices[:i], p.settings.PinedServices[i+1:]...)
					err := p.api.SaveModuleConfig(&p.settings)
					if err != nil {
						return NewErrorAlertActionResult(err)
					}

					return NewReloadActionResult()
				}
			}

			return NewReloadActionResult()
		},

		"action": func(args []string, data io.Reader) ActionResult {
			var (
				action   = args[0]
				unitName = args[1]
				err      error
			)

			switch action {
			case "start":
				_, err = p.dbusConn.StartUnitContext(p.ctx, unitName, "fail", p.systendChan)
			case "stop":
				_, err = p.dbusConn.StopUnitContext(p.ctx, unitName, "fail", p.systendChan)
			case "restart":
				_, err = p.dbusConn.RestartUnitContext(p.ctx, unitName, "fail", p.systendChan)
			case "try-restart":
				_, err = p.dbusConn.TryRestartUnitContext(p.ctx, unitName, "fail", p.systendChan)
			case "reload":
				_, err = p.dbusConn.ReloadUnitContext(p.ctx, unitName, "fail", p.systendChan)
			}

			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			return NewReloadActionResult()
		},

		"edit": func(args []string, data io.Reader) ActionResult {
			var (
				action   = args[0]
				unitName = args[1]
				err      error
			)

			properties, err := p.dbusConn.GetAllPropertiesContext(p.ctx, unitName)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			if unitFile, ok := properties["FragmentPath"]; ok {
				if action == "save" {
					resp := struct {
						Data string `json:"data"`
					}{}

					err = json.NewDecoder(data).Decode(&resp)
					if err != nil {
						return NewErrorAlertActionResult(err)
					}

					err = ioutil.WriteFile(unitFile.(string), []byte(resp.Data), 0644)
					if err != nil {
						return NewErrorAlertActionResult(err)
					}

					err = p.dbusConn.ReloadContext(p.ctx)
					if err != nil {
						return NewErrorAlertActionResult(err)
					}

					return NewReloadActionResult()
				}

				unitContent, err := ioutil.ReadFile(unitFile.(string))
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				return NewFormModalActionResult(
					"Edit unit file",
					NewForm().
						AddWithTitle("Unit file content", NewCodeEditor("data", string(unitContent))).
						AddActionButtons(
							NewButton("Cancel", "none").SetStyle(StyleSecondary),
							NewButton("Save", "edit", "save", unitName),
						),
				)
			}

			return NewErrorAlertActionResult(errors.New("Unit file not found"))
		},

		"create-service": func(args []string, data io.Reader) ActionResult {
			action := args[0]

			if action == "save" {
				resp := struct {
					Name        string `json:"name"`
					Exec        string `json:"exec"`
					Description string `json:"description"`
				}{}

				err := json.NewDecoder(data).Decode(&resp)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				err = ioutil.WriteFile(
					fmt.Sprintf("/etc/systemd/system/%s.service", resp.Name),
					[]byte(NewServiceTemplate(resp.Description, resp.Exec)),
					0644,
				)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				err = p.dbusConn.ReloadContext(p.ctx)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				return NewReloadActionResult()
			}

			return NewFormModalActionResult(
				"New service",
				NewForm().
					AddWithTitle("Service name", NewInput("name")).
					AddWithTitle("Exec", NewInput("exec")).
					AddWithTitle("Description", NewInput("description")).
					AddActionButtons(
						NewButton("Cancel", "none").SetStyle(StyleSecondary),
						NewButton("Create", "create-service", "save"),
					),
			)
		},

		"none": func(args []string, data io.Reader) ActionResult {
			return NewReloadActionResult()
		},
	}
}

func NewServiceTemplate(description string, exec string) string {
	tpl := `[Unit]
Description=%s

[Service]
Type=oneshot
ExecStart=%s
# ExecStop=


[Install]
WantedBy=multi-user.target`

	return fmt.Sprintf(tpl, description, exec)
}

func AddServiceToTable(table *Table, unit dbus.UnitStatus, action string) {
	badgesLine := NewLine()

	statusBadge := NewBadge(unit.ActiveState)

	switch unit.ActiveState {
	case "active":
		statusBadge.SetStyle(StyleSuccess)
	case "failed":
		statusBadge.SetStyle(StyleDanger)
	case "inactive":
		statusBadge.SetStyle(StyleSecondary)
	}

	badgesLine.Add(statusBadge)

	if unit.LoadState == "loaded" {
		loadedBadge := NewBadge(unit.LoadState)
		badgesLine.Add(loadedBadge)
	}

	if unit.SubState == "running" {
		badgesLine.Add(NewBadge(unit.SubState).SetStyle(StyleSuccess))
	}

	table.AddLine(
		NewElementsList().AddElements(
			NewLine(
				NewLabel(unit.Name),
				NewButton("", action, unit.Name).SetImage("pin-angle").SetLinkStyle(),
				NewDropdown().
					AddItem("play-fill", "Start", "action", "start", unit.Name).
					AddItem("stop-fill", "Stop", "action", "stop", unit.Name).
					AddItem("arrow-repeat", "Restart", "action", "restart", unit.Name).
					AddItem("chat-right-dots", "reload", "action", "reload", unit.Name).
					AddSeparator().
					AddItem("lightning-charge-fill", "Try restart", "action", "try-restart", unit.Name).
					AddSeparator().
					AddItem("brightness-high-fill", "Enable", "").
					AddItem("moon-fill", "Disable", "").
					AddSeparator().
					AddItem("pencil-square", "Edit unit", "edit", "", unit.Name).
					AddItem("trash", "Delete", ""),
			),
			badgesLine,
		).SetModeLine(),
	)
}

func (p *Plugin) Render(args []string) Page {

	units, err := p.dbusConn.ListUnitsContext(p.ctx)
	if err != nil {
		return NewPage("Systemd",
			NewText(err.Error()),
		)
	}

	pinnedServices := NewTable("Pinned services")

	for _, ps := range p.settings.PinedServices {
		for _, u := range units {
			if ps == u.Name {
				AddServiceToTable(pinnedServices, u, "unpin")
			}
		}
	}

	otherServices := NewTable("Other services")

main:
	for _, u := range units {
		if !strings.HasSuffix(u.Name, ".service") {
			continue main
		}

		for _, ps := range p.settings.PinedServices {
			if ps == u.Name {
				continue main
			}
		}

		AddServiceToTable(otherServices, u, "pin")
	}

	return NewPage("Systemd",
		NewButton("Create service", "create-service", ""),
		pinnedServices,
		otherServices,
	)
}

func jsonDump(data any) string {
	buf := bytes.NewBuffer([]byte{})
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "    ")
	_ = enc.Encode(data)
	return buf.String()
}
