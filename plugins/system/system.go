package system

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"

	. "qubert/pluginTools"
)

const (
	hostNameFile = "/etc/hostname"
)

type PluginSettings struct {
}

type Plugin struct {
	api      PluginAPI
	ctx      context.Context
	settings PluginSettings
}

func (p *Plugin) ID() string {
	return "system"
}

func (p *Plugin) Title() string {
	return "System"
}

func (p *Plugin) Icon() string {
	return "gear"
}

func (p *Plugin) Run(ctx context.Context, api PluginAPI) error {
	p.api = api
	p.ctx = ctx

	err := p.api.LoadModuleConfig(&p.settings)
	if err != nil {
		return err
	}

	p.runCpuMonitor()

	return nil
}

func (p *Plugin) Actions() ActionsMap {
	return ActionsMap{
		"action": func(args []string, data io.Reader) ActionResult {
			action := args[0]

			var err error

			switch action {
			case "stop":
				p.api.Exit()
			case "shutdown":
				err = p.api.Shutdown()
			case "restart":
				err = p.api.Restart()
			}

			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			return NewReloadActionResult()
		},

		"update": func(args []string, data io.Reader) ActionResult {
			reqData := struct {
				Value string `json:"value"`
			}{}

			err := json.NewDecoder(data).Decode(&reqData)
			if err != nil {
				panic(err)
			}

			switch args[0] {
			case "host-name":
				err = applyHostName(reqData.Value)
			}

			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			return NewReloadActionResult()
		},
	}
}

func (p *Plugin) Render(args []string) Page {
	version, commit := p.api.Version()
	hostName := getHostName()

	return NewPage(
		"System",
		NewHeader("Main"),
		NewElementsList().SetModeLine().
			AddElementWithTitle(NewLabel("Version:").SetStrong(true), NewLabel(version)).
			AddElementWithTitle(NewLabel("Commit:").SetStrong(true), NewLabel(commit)).
			AddElementWithTitle(NewLabel("Go version:").SetStrong(true), NewLabel(runtime.Version())).
			AddElementWithTitle(NewLabel("Arch/OS:").SetStrong(true), NewLabel("%s/%s", runtime.GOARCH, runtime.GOOS)).
			AddElementWithTitle(NewLabel("PID:").SetStrong(true), NewLabel("%d", os.Getpid())).
			AddElementWithTitle(NewLabel("Started by:").SetStrong(true), NewLabel("%d", os.Getuid())),
		NewButton("Stop", "action", "stop").SetStyle(StyleDanger),
		NewHeader("Host"),
		NewElementsList().SetModeLine().
			AddElementWithTitle(NewLabel("Host name").SetStrong(true), NewInputEdit("value", hostName, "update", "host-name")).
			AddElementWithTitle(NewLabel("CPU usage").SetStrong(true), NewProgress(0).SetID("cpu-usage")),
		NewLine(
			NewButton("Shutdown", "action", "shutdown").SetStyle(StyleDanger),
			NewButton("Restart", "action", "restart").SetStyle(StyleDanger),
		),
	)
}

func getHostName() string {
	data, err := ioutil.ReadFile(hostNameFile)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(data), "\n")

	if len(lines) > 0 {
		return lines[0]
	}

	return ""
}

func applyHostName(hostName string) error {
	data := fmt.Sprintln(hostName)
	return ioutil.WriteFile(hostNameFile, []byte(data), 0644)
}

func (p *Plugin) runCpuMonitor() {
	go func() {
		var idle0, total0 uint64

		for {
			idle1, total1, err := cpuSample()
			if err != nil {
				fmt.Println(err)
			}

			idleTicks := idle1 - idle0
			totalTicks := total1 - total0

			idle0 = idle1
			total0 = total1

			cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks

			wasSent := p.api.SendUpdate(NewUpdateProgress("cpu-usage", uint(cpuUsage)))

			if wasSent {
				time.Sleep(time.Second / 5)
			} else {
				time.Sleep(time.Second * 2)
			}
		}
	}()
}
