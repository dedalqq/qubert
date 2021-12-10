package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	. "qubert/pluginTools"

	"qubert/uuid"
)

type service struct {
	uuid uuid.UUID

	Name        string   `json:"name"`
	CMD         string   `json:"cmd"`
	Args        []string `json:"args"`
	Dir         string   `json:"dir"`
	Env         []string `json:"env"`
	Description string   `json:"description,omitempty"`
	Autostart   bool     `json:"autostart,omitempty"`

	process      *os.Process
	processState *os.ProcessState

	startedAt time.Time
}

func (s *service) start(cb func()) error {
	attr := &os.ProcAttr{
		Dir: s.Dir,
		Env: os.Environ(),
		Files: []*os.File{
			os.Stdin,
			nil,
			nil,
		},
		Sys: &syscall.SysProcAttr{
			//Noctty: true,
			//Credential: &syscall.Credential{
			//	Uid: uint32(os.Getuid()),
			//	Gid: uint32(os.Getuid()),
			//},
		},
	}

	var err error

	s.process, err = os.StartProcess(
		s.CMD,
		append(
			[]string{s.CMD},
			s.Args...,
		), attr)

	if err != nil {
		return err
	}

	s.startedAt = time.Now()

	go func() {
		s.processState, err = s.process.Wait()
		s.process = nil

		cb()
	}()

	return nil
}

func (s *service) sendSignal(sig syscall.Signal) error {
	if p := s.process; p != nil {
		return p.Signal(sig)
	}

	return errors.New("service not started")
}

type PluginSettings struct {
	Services []*service `json:"services"`
}

func (ps *PluginSettings) FindServiceByName(name string) *service {
	for _, s := range ps.Services {
		if s.Name == name {
			return s
		}
	}

	return nil
}

func (ps *PluginSettings) FindServiceByUUID(uuid uuid.UUID) *service {
	for _, s := range ps.Services {
		if s.uuid == uuid {
			return s
		}
	}

	return nil
}

type Plugin struct {
	api      PluginAPI
	ctx      context.Context
	settings PluginSettings
}

func (p *Plugin) ID() string {
	return "services"
}

func (p *Plugin) Title() string {
	return "Services"
}

func (p *Plugin) Icon() string {
	return "card-checklist"
}

func (p *Plugin) Run(ctx context.Context, api PluginAPI) error {
	p.api = api
	p.ctx = ctx

	err := p.api.LoadModuleConfig(&p.settings)
	if err != nil {
		return err
	}

	for _, s := range p.settings.Services {
		s.uuid = uuid.New()

		if s.Autostart {
			_ = s.start(func() {
				p.api.Reload()
			}) // TODO error handling
		}
	}

	return nil
}

type serviceCreateDef struct {
	Name    string `json:"name"`
	Command string `json:"cmd"`
}

type serviceUpdateDef struct {
	Name        *string   `json:"name"`
	Command     *string   `json:"cmd"`
	Args        *[]string `json:"args"`
	Dir         *string   `json:"dir"`
	Description *string   `json:"description"`
	Autostart   *bool     `json:"autostart"`
}

type signalReqData struct {
	Signal string `json:"signal"`
}

func (p *Plugin) Actions() ActionsMap {
	return ActionsMap{
		"add-service": func(args []string, data io.Reader) ActionResult {
			action := args[0]

			nameInput := NewInput("name")
			cmdInput := NewInput("cmd")

			if action == "save" {
				var sd serviceCreateDef

				err := json.NewDecoder(data).Decode(&sd)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				nameInput.SetValue(sd.Name)
				cmdInput.SetValue(sd.Command)

				isValid := true

				if err = p.validateName(sd.Name, ""); err != nil {
					isValid = false
					nameInput.SetErrorText(err.Error())
				}

				if err = p.validateCommand(sd.Command); err != nil {
					isValid = false
					cmdInput.SetErrorText(err.Error())
				}

				if isValid {
					p.settings.Services = append(p.settings.Services, &service{
						uuid: uuid.New(),
						Name: sd.Name,
						Args: []string{},
						CMD:  sd.Command,
						Env:  []string{},
					})

					err = p.api.SaveModuleConfig(&p.settings)
					if err != nil {
						return NewErrorAlertActionResult(err)
					}

					return NewReloadActionResult()
				}
			}

			return NewFormModalActionResult(
				"Add new service",
				NewForm().
					AddWithTitle("Name", nameInput).
					AddWithTitle("Command", cmdInput).
					AddActionButtons(
						NewButton("Cancel", "none").SetStyle(StyleSecondary),
						NewButton("Add", "add-service", "save"),
					),
			)
		},

		"update": func(args []string, data io.Reader) ActionResult {
			var sd serviceUpdateDef

			serviceID := uuid.UUID(args[0])

			s := p.settings.FindServiceByUUID(serviceID)
			if s == nil {
				return NewErrorAlertActionResult(errors.New("service not found"))
			}

			err := json.NewDecoder(data).Decode(&sd)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			if sd.Name != nil {
				if err = p.validateName(*sd.Name, s.uuid); err != nil {
					return NewErrorAlertActionResult(err)
				}

				s.Name = *sd.Name
			}

			if sd.Command != nil {
				if err = p.validateCommand(*sd.Command); err != nil {
					return NewErrorAlertActionResult(err)
				}

				s.CMD = *sd.Command
			}

			if unsafe.Pointer(sd.Args) != nil {
				s.Args = *sd.Args
			}

			if sd.Dir != nil {
				if err = p.validateWorkDir(*sd.Dir); err != nil {
					return NewErrorAlertActionResult(err)
				}

				s.Dir = *sd.Dir
			}

			if sd.Description != nil {
				s.Description = *sd.Description
			}

			if sd.Autostart != nil {
				s.Autostart = *sd.Autostart
			}

			err = p.api.SaveModuleConfig(&p.settings)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			return NewReloadActionResult()
		},

		"select-service": func(args []string, data io.Reader) ActionResult {
			if len(args) == 0 {
				return NewSetArgsActionResult()
			}

			serviceID := uuid.UUID(args[0])

			return NewSetArgsActionResult(serviceID.String())
		},

		"rename": func(args []string, data io.Reader) ActionResult {
			serviceID := uuid.UUID(args[0])
			action := args[1]

			s := p.settings.FindServiceByUUID(serviceID)
			if s == nil {
				return NewErrorAlertActionResult(errors.New("service not found"))
			}

			nameInput := NewInput("name").SetValue(s.Name)

			if action == "save" {
				var sd serviceCreateDef

				err := json.NewDecoder(data).Decode(&sd)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				err = p.validateName(sd.Name, s.uuid)
				if err == nil {
					s.Name = sd.Name

					err = p.api.SaveModuleConfig(&p.settings)
					if err != nil {
						return NewErrorAlertActionResult(err)
					}

					return NewReloadActionResult()
				}

				nameInput.SetErrorText(err.Error())
			}

			return NewFormModalActionResult(
				"Rename service",
				NewForm().
					AddWithTitle("Name", nameInput).
					AddActionButtons(
						NewButton("Cancel", "none").SetStyle(StyleSecondary),
						NewButton("Save", "rename", serviceID.String(), "save"),
					),
			)
		},

		"start": func(args []string, data io.Reader) ActionResult {
			serviceID := uuid.UUID(args[0])

			s := p.settings.FindServiceByUUID(serviceID)
			if s == nil {
				return NewErrorAlertActionResult(errors.New("service not found"))
			}

			err := s.start(func() {
				p.api.Reload()
			})

			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			return NewReloadActionResult()
		},

		"send-signal": func(args []string, data io.Reader) ActionResult {
			serviceID := uuid.UUID(args[0])
			send := len(args) > 1 && args[1] == "send"

			s := p.settings.FindServiceByUUID(serviceID)
			if s == nil {
				return NewErrorAlertActionResult(errors.New("service not found"))
			}

			if send {
				reqData := signalReqData{}

				err := json.NewDecoder(data).Decode(&reqData)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				sig, err := strconv.Atoi(reqData.Signal)
				if err != nil {
					return NewErrorAlertActionResult(errors.New("incorrect signal"))
				}

				err = s.sendSignal(syscall.Signal(sig))
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				return NewReloadActionResult()
			}

			selectInput := NewSelect("signal")

			for _, sig := range signals {
				selectInput.AddNamedOption(fmt.Sprintf("(%d) %s", sig.Signal, sig.string), fmt.Sprintf("%d", sig.Signal))
			}

			return NewFormModalActionResult(
				"Send signal",
				NewForm().
					AddWithTitle("Select signal", selectInput).
					AddActionButtons(
						NewButton("Cancel", "none").SetStyle(StyleSecondary),
						NewButton("Send", "send-signal", serviceID.String(), "send").SetStyle(StyleDanger),
					),
			)
		},

		"signal": func(args []string, data io.Reader) ActionResult {
			serviceID := uuid.UUID(args[0])
			confirm := len(args) > 2 && args[2] == "confirm"

			sig, err := strconv.Atoi(args[1])
			if err != nil {
				return NewErrorAlertActionResult(errors.New("incorrect signal"))
			}

			if confirm {
				return NewModalActionResult(
					"Send signal",
					NewLabel("Do you sure about this?"),
					NewButton("Send", "signal", serviceID.String(), args[1]).SetStyle(StyleDanger),
					NewButton("Cancel", "none"),
				)
			}

			s := p.settings.FindServiceByUUID(serviceID)
			if s == nil {
				return NewErrorAlertActionResult(errors.New("service not found"))
			}

			err = s.sendSignal(syscall.Signal(sig))
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			return NewReloadActionResult()
		},

		"delete": func(args []string, data io.Reader) ActionResult {
			serviceID := uuid.UUID(args[0])
			confirm := len(args) > 1 && args[1] == "confirm"

			if confirm {
				return NewModalActionResult(
					"Delete service",
					NewLabel("Do you sure about this?"),
					NewButton("Delete", "delete", serviceID.String()).SetStyle(StyleDanger),
					NewButton("Cancel", "none"),
				)
			}

			for i, s := range p.settings.Services {
				if s.uuid == serviceID {
					p.settings.Services = append(p.settings.Services[:i], p.settings.Services[i+1:]...)
					break
				}
			}

			err := p.api.SaveModuleConfig(&p.settings)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			return NewReloadActionResult()
		},

		"none": func(args []string, data io.Reader) ActionResult {
			return NewReloadActionResult()
		},
	}
}

func (p *Plugin) validateName(name string, id uuid.UUID) error {
	if name == "" {
		return errors.New("service name can't de empty")
	}

	if s := p.settings.FindServiceByName(name); s == nil || (id != "" && s.uuid == id) {
		return nil
	}

	return errors.New("service with this name already exist")
}

func (p *Plugin) validateCommand(cmd string) error {
	if cmd == "" {
		return errors.New("command can't de empty")
	}

	_, err := os.Stat(cmd)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("file not exist")
		}

		return err
	}

	return nil
}

func (p *Plugin) validateWorkDir(dir string) error {
	if dir == "" {
		return nil
	}

	s, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("directory not exist")
		}

		if !s.IsDir() {
			return errors.New("file is not a directory")
		}

		return err
	}

	return nil
}

func (p *Plugin) RenderService(serviceID uuid.UUID) Page {
	s := p.settings.FindServiceByUUID(serviceID)

	if s == nil {
		return NewPage("Error", NewText("Service not found."))
	}

	var (
		pid       string
		startedAt string
		controls  *Line
	)

	if pr := s.process; pr != nil {

		pid = fmt.Sprintf("%d", pr.Pid)
		startedAt = s.startedAt.Format(time.RFC1123)
		controls = NewLine(
			NewButton("Stop", "signal", serviceID.String(), "2", "confirm").SetStyle(StyleDanger),
			NewButton("Kill", "signal", serviceID.String(), "9", "confirm").SetStyle(StyleDanger),
			NewButton("Reload", "signal", serviceID.String(), "1").SetStyle(StyleDanger),
			//NewButton("Restart", "service-action", "restart", serviceID.String()).SetStyle(StyleDanger),
			NewButton("Send signal", "send-signal", serviceID.String()).SetStyle(StyleDanger),
		)
	} else {

		controls = NewLine(
			NewButton("Start", "start", serviceID.String()).SetImage("play").SetStyle(StyleSuccess),
		)
	}

	return NewPage(
		fmt.Sprintf("Services %s", s.Name),
		NewButton("Back", "select-service").SetImage("arrow-left-short"),
		NewHeader("Service info"),
		NewElementsList().SetModeLine().
			AddElementWithTitle(
				NewLabel("Name").SetStrong(true),
				NewInputEdit("name", s.Name, "update", serviceID.String()),
			).
			AddElementWithTitle(
				NewLabel("Command").SetStrong(true),
				NewInputEdit("cmd", s.CMD, "update", serviceID.String()),
			).
			AddElementWithTitle(
				NewLabel("Args").SetStrong(true),
				NewTagsEdit("args", s.Args, "update", serviceID.String()),
			).
			AddElementWithTitle(
				NewLabel("Work directory").SetStrong(true),
				NewInputEdit("dir", s.Dir, "update", serviceID.String()),
			).
			AddElementWithTitle(
				NewLabel("Description").SetStrong(true),
				NewTextareaEdit("description", s.Description, "update", serviceID.String()),
			).
			AddElementWithTitle(
				NewLabel("Auto start").SetStrong(true),
				NewSwitch("autostart").SetAction(NewAction("update", serviceID.String())).SetValue(s.Autostart),
			),
		NewHeader("Service status"),
		NewElementsList().SetModeLine().
			AddElementWithTitle(NewLabel("Status").SetStrong(true), statusBadge(s)).
			AddElementWithTitle(NewLabel("PID").SetStrong(true), NewLabel(pid)).
			AddElementWithTitle(NewLabel("Started at").SetStrong(true), NewLabel(startedAt)),
		NewHeader("Service controls"),
		controls,
		//NewHeader("Service output"),
		//NewTerminal(),
	)
}

func (p *Plugin) RenderServiceList() Page {
	table := NewTable("Services", "status", "Pid")

	for _, s := range p.settings.Services {
		table.AddLine(
			NewElementsList().AddElements(
				NewLine(
					NewButton(s.Name, "select-service", s.uuid.String()).SetLinkStyle(),
					serviceDropdown(s),
				),
			).SetModeLine(),
			statusBadge(s),
			pidLabel(s),
		)
	}

	return NewPage(
		"Services",
		NewButton("Add", "add-service", "").SetImage("plus-lg"),
		table,
	)
}

func (p *Plugin) Render(args []string) Page {
	if len(args) > 0 {
		serviceID := uuid.UUID(args[0])

		return p.RenderService(serviceID)
	}

	return p.RenderServiceList()
}

func statusBadge(s *service) *Badge {
	if s.process != nil {
		return NewBadge("running").SetStyle(StyleSuccess)
	}

	if ps := s.processState; ps == nil || ps.Success() {
		return NewBadge("stopped").SetStyle(StyleSecondary)
	}

	return NewBadge("failed").SetStyle(StyleDanger)
}

func serviceDropdown(s *service) *Dropdown {
	dd := NewDropdown()
	dd.AddItem("pencil", "Rename", "rename", s.uuid.String(), "")
	dd.AddSeparator()

	var signalAction string
	var startAction string

	if s.process != nil {
		signalAction = "signal"
	} else {
		startAction = "start"
	}

	dd.AddItem("play", "Start", startAction, s.uuid.String())
	dd.AddItem("stop", "Stop", signalAction, s.uuid.String(), "2", "confirm")
	dd.AddItem("x", "Kill", signalAction, s.uuid.String(), "9", "confirm")
	dd.AddItem("arrow-repeat", "Reload", signalAction, s.uuid.String(), "1")

	dd.AddSeparator()
	dd.AddDangerItem("trash", "Delete", "delete", s.uuid.String(), "confirm")

	return dd
}

func pidLabel(s *service) *Label {
	if p := s.process; p != nil {
		return NewLabel("%d", s.process.Pid)
	}

	return NewLabel("")
}
