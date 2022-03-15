package testPlugins

import (
	"context"
	"fmt"
	"io"
	"time"

	. "qubert/pluginTools"
)

type Plugin struct {
	api PluginAPI
}

func (p *Plugin) ID() string {
	return "test-plugin"
}

func (p *Plugin) Title() string {
	return "Test plugin"
}

func (p *Plugin) Icon() string {
	return "shop"
}

func (p *Plugin) Run(ctx context.Context, api PluginAPI) error {
	p.api = api

	t := time.NewTicker(5 * time.Second)

	var i uint = 0

	for {
		i++

		select {
		case <-t.C:
			api.Send(struct {
				Data string `json:"data"`
			}{
				Data: "omg",
			})

			api.SendUpdate(NewUpdateProgress("test_progress", i))
		case <-ctx.Done():
			return nil
		}
	}

}

func (p *Plugin) Actions() ActionsMap {
	return ActionsMap{
		"test": func(args []string, data io.Reader) ActionResult {
			fmt.Println(args)

			//return NewModalActionResult(NewButton("ok", "omg", "qq"), NewButton("fail", "omg", "qq"))

			return NewAlertActionResult("omg", "ppc")
		},

		"omg": func(args []string, data io.Reader) ActionResult {
			fmt.Println(args)

			return NewReloadActionResult()
		},

		"open-form": func(args []string, data io.Reader) ActionResult {
			return NewFormModalActionResult(
				"Test modal form",
				NewForm().
					AddWithTitle("Text field name", NewInput("qq")).
					AddWithTitle("Text textarea", NewTextarea("omg").SetErrorText("omg")).
					AddWithTitle("select", NewSelect("ppc").AddOption("qq").AddOption("ww")).
					AddWithTitle("Code", NewCodeEditor("lol", "qq\nww")).
					AddActionButton(NewButton("Send", "form-data")),
			)
		},
	}
}

func (p *Plugin) Render(args []string) Page {
	return NewPage("omg",
		NewLabel("ppc"),
		NewButton("qq", "test", "q1", "q2").SetDisabled(true).SetImage("asterisk"),
		NewProgress(25).SetID("test_progress"),
		NewLine(NewBadge("qq"), NewBadge("ww"), NewButton("qq", "qq"), NewDropdown().AddItem("shop", "qq", "qwe")),
		NewTagsEdit("qq", []string{"omg", "ppc", "wtf"}, "qq"),
		NewButton("Open form", "open-form", "q1", "q2"),
		NewForm().AddWithTitle("q1", NewInput("q1")).AddWithTitle("q2", NewInput("q2")).AddActionButton(NewButton("Send", "send")).AddActionButton(NewButton("Clear", "clear")),
		NewTable("omg", "ppc").AddLine(NewLabel("qq"), NewLabel("ww")).AddLine(NewLabel("qq"), NewLabel("ww")),
		NewBSImage("shop"),
		NewText("qq"),
		NewCard("card", NewText("qq")).SetHeaderIcon("shop").SetAdditionalHeader(NewDropdown().AddItem("shop", "qq", "qwe").AddSeparator().AddItem("shop", "ww", "qwe")),
		NewInput("qq"),
		NewInputEdit("name", "qq", "action"),
	)
}
