package interfaces

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/vishvananda/netlink"

	. "qubert/pluginTools"
)

type PluginSettings struct {
}

type Plugin struct {
	api      PluginAPI
	ctx      context.Context
	settings PluginSettings
}

func (p *Plugin) ID() string {
	return "network-interfaces"
}

func (p *Plugin) Title() string {
	return "Interfaces"
}

func (p *Plugin) Icon() string {
	return "hdd-network"
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
		"select-dev": func(args []string, data io.Reader) ActionResult {
			return NewSetArgsActionResult(args...)
		},

		"add-ip-address": func(args []string, data io.Reader) ActionResult {
			linkName := args[0]
			confirm := "confirm" == args[1]

			addrInput := NewInput("ip-addr")

			if confirm {
				reqData := struct {
					//Label  string `json:"label"`
					IPAddr string `json:"ip-addr"`
				}{}

				err := json.NewDecoder(data).Decode(&reqData)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				addrInput.SetValue(reqData.IPAddr)

				ip, ipNet, err := net.ParseCIDR(reqData.IPAddr)
				if err == nil {
					link, err := netlink.LinkByName(linkName)
					if err != nil {
						return NewErrorAlertActionResult(err)
					}

					err = netlink.AddrAdd(link, &netlink.Addr{
						IPNet: &net.IPNet{
							IP:   ip,
							Mask: ipNet.Mask,
						},
						Label: link.Attrs().Name,
					})
					if err != nil {
						return NewErrorAlertActionResult(err)
					}

					return NewReloadActionResult()
				} else {
					addrInput.SetErrorText("Failed to parse ip addr")
				}
			}

			form := NewForm()

			//form.AddWithTitle("Label", NewInput("label"))
			form.AddWithTitle("IP address/masc", addrInput)

			form.AddActionButton(NewButton("Add", "add-ip-address", linkName, "confirm"))

			return NewFormModalActionResult(fmt.Sprintf("Add IP address for %s", linkName), form)
		},

		"delete-ip-address": func(args []string, data io.Reader) ActionResult {
			linkName := args[0]
			ipAddr := args[1]
			confirm := "confirm" == args[2]

			if confirm {
				return NewModalActionResult(
					"Delete IP address",
					NewLabel("Do you sure about this?"),
					NewButton("Delete", "delete-ip-address", linkName, ipAddr, "").SetStyle(StyleDanger),
					NewButton("Cancel", "none"),
				)
			}

			link, err := netlink.LinkByName(linkName)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			ip, ipNet, err := net.ParseCIDR(ipAddr)
			if err != nil {
				return NewReloadActionResult()
			}

			err = netlink.AddrDel(link, &netlink.Addr{
				IPNet: &net.IPNet{
					IP:   ip,
					Mask: ipNet.Mask,
				},
			})
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

func (p *Plugin) Render(args []string) Page {
	if len(args) > 0 {
		devName := args[0]
		return p.renderDevice(devName)
	}

	return p.renderDevList()
}

func (p *Plugin) renderDevice(devName string) Page {
	link, err := netlink.LinkByName(devName)
	if err != nil {
		panic(err)
	}

	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		panic(err)
	}

	addressTable := NewTable("#", "IP address", "Label", "")

	for n, a := range addrs {
		addressTable.AddLine(
			NewLabel("%d", n+1).SetStrong(true),
			NewLabel(a.IPNet.String()),
			NewLabel(a.Label),
			NewImageButton("trash", "delete-ip-address", link.Attrs().Name, a.IPNet.String(), "confirm").SetLinkStyle(),
		)
	}

	return NewPage(
		fmt.Sprintf("Network interface %s", link.Attrs().Name),
		NewButton("Back", "select-dev"),
		NewHeader("Interface info"),
		NewElementsList().SetModeLine().
			AddElementWithTitle(NewLabel("Name").SetStrong(true), NewLabel(link.Attrs().Name)).
			AddElementWithTitle(NewLabel("Type").SetStrong(true), NewLabel(link.Type())).
			AddElementWithTitle(NewLabel("Mac").SetStrong(true), NewLabel(link.Attrs().HardwareAddr.String())),
		NewHeader("IP addresses"),
		NewButton("Add address", "add-ip-address", link.Attrs().Name, ""),
		addressTable,
	)
}

func (p *Plugin) renderDevList() Page {
	links, err := netlink.LinkList()
	if err != nil {
		panic(err)
	}

	table := NewTable("Dev name", "Type", "Mac", "IP")

	for _, l := range links {
		addrs, err := netlink.AddrList(l, netlink.FAMILY_V4)
		if err != nil {
			panic(err)
		}

		addrsList := NewElementsList().SetModeLine()

		for _, a := range addrs {
			addrsList.AddElements(NewLabel(a.IPNet.String()))
		}

		table.AddLine(
			NewButton(l.Attrs().Name, "select-dev", l.Attrs().Name).SetLinkStyle(),
			NewLabel(l.Type()),
			NewLabel(l.Attrs().HardwareAddr.String()),
			addrsList,
		)
	}

	return NewPage(
		"Network interfaces",
		table,
	)
}
