package panel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"qubert/uuid"

	. "qubert/pluginTools"
)

type widget struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Type string    `json:"type"`

	FileName    string `json:"file-name"`
	Description string `json:"description"`
}

type PluginSettings struct {
	Widgets []widget `json:"widgets"`
}

type Plugin struct {
	api      PluginAPI
	ctx      context.Context
	settings PluginSettings
}

func (p *Plugin) ID() string {
	return "panel"
}

func (p *Plugin) Title() string {
	return "Admin panel"
}

func (p *Plugin) Icon() string {
	return "shop"
}

func (p *Plugin) Run(ctx context.Context, api PluginAPI) error {
	p.api = api
	p.ctx = ctx

	err := p.api.LoadModuleConfig(&p.settings)
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) Actions() ActionsMap {
	return ActionsMap{
		"show-add-form": func(args []string, data io.Reader) ActionResult {
			return NewFormModalActionResult(
				"Add element",
				NewForm().
					AddWithTitle("Name", NewInput("name")).
					AddWithTitle("Element type", NewSelect("type").AddNamedOption("Edit file", "edit-file")).
					AddActionButton(NewButton("add", "create-element")),
			)
		},

		"create-element": func(args []string, data io.Reader) ActionResult {
			dataReq := struct {
				Name string `json:"name"`
				Type string `json:"type"`
			}{}

			err := json.NewDecoder(data).Decode(&dataReq)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			p.settings.Widgets = append(p.settings.Widgets, widget{
				ID:   uuid.New(),
				Name: dataReq.Name,
				Type: dataReq.Type,
			})

			err = p.api.SaveModuleConfig(&p.settings)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			return NewReloadActionResult()
		},

		"edit-file-edit-widget": func(args []string, data io.Reader) ActionResult {
			var wd *widget

			for i := range p.settings.Widgets {
				if string(p.settings.Widgets[i].ID) == args[0] {
					wd = &p.settings.Widgets[i]
					break
				}
			}

			if wd == nil {
				return NewErrorAlertActionResult(errors.New("widget not found"))
			}

			if args[1] == "open" {
				return NewFormModalActionResult(
					"Edit widget",
					NewForm().
						AddWithTitle("Name", NewInput("name").SetValue(wd.Name)).
						AddWithTitle("File path", NewInput("file").SetValue(wd.FileName)).
						AddWithTitle("Description", NewTextarea("description").SetValue(wd.Description)).
						AddActionButton(NewButton("Save", "edit-file-edit-widget", args[0], "save")),
				)
			}

			dataReq := struct {
				Name        string `json:"name"`
				File        string `json:"file"`
				Description string `json:"description"`
			}{}

			err := json.NewDecoder(data).Decode(&dataReq)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			wd.FileName = dataReq.File
			wd.Name = dataReq.Name
			wd.Description = dataReq.Description

			err = p.api.SaveModuleConfig(&p.settings)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			return NewReloadActionResult()
		},

		"edit-file": func(args []string, data io.Reader) ActionResult {
			var wd *widget

			for i := range p.settings.Widgets {
				if string(p.settings.Widgets[i].ID) == args[0] {
					wd = &p.settings.Widgets[i]
					break
				}
			}

			if wd == nil {
				return NewErrorAlertActionResult(errors.New("widget not found"))
			}

			if args[1] == "save" {
				dataReq := struct {
					File string `json:"file"`
				}{}

				err := json.NewDecoder(data).Decode(&dataReq)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				err = ioutil.WriteFile(wd.FileName, []byte(dataReq.File), 0644)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				return NewReloadActionResult()
			}

			fileContent, err := ioutil.ReadFile(wd.FileName)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			return NewFormModalActionResult(
				"Edit File",
				NewForm().
					Add(NewCodeEditor("file", string(fileContent))).
					AddActionButton(NewButton("Save", "edit-file", args[0], "save")),
			)
		},
	}
}

func (p *Plugin) Render(args []string) Page {
	var widgetsElements []Element

	for _, w := range p.settings.Widgets {
		switch w.Type {
		case "edit-file":
			body := NewElementsList().SetModeLine()

			info, err := os.Stat(w.FileName)
			if err != nil {
				continue // TODO
			}

			body.AddElementWithTitle(NewLabel("File name").SetStrong(true), NewInputEdit("qq", w.FileName, ""))
			body.AddElementWithTitle(NewLabel("File size").SetStrong(true), NewLabel(fmt.Sprintf("%d", info.Size())))
			body.AddElementWithTitle(NewLabel("Last edit").SetStrong(true), NewLabel(info.ModTime().Format(time.RFC822)))

			if w.Description != "" {
				body.AddElementWithTitle(NewLabel("Description").SetStrong(true), NewText(w.Description))
			}

			body.AddElements(NewButton("Edit file", "edit-file", string(w.ID), ""))

			widgetsElements = append(
				widgetsElements,
				NewCard(w.Name, body).SetAdditionalHeader(
					NewImageButton("pencil", "edit-file-edit-widget", string(w.ID), "open"),
				),
			)
		}
	}

	return NewPage(
		"Admin panel",
		NewButton("Add", "show-add-form"),
		NewElementsList().AddElements(widgetsElements...),
	)
}
