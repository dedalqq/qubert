package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"time"

	. "qubert/pluginTools"
)

type RecordType string

const (
	RecordTypeA     = "A"
	RecordTypeAAAA  = "AAAA"
	RecordTypeMX    = "MX"
	RecordTypeCNAME = "CNAME"
	RecordTypeDNAME = "DNAME"
	RecordTypeTXT   = "TXT"
	RecordTypePTR   = "PTR"
	RecordTypeSRV   = "SRV"
)

var allType = []RecordType{
	RecordTypeA,
	RecordTypeAAAA,
	RecordTypeMX,
	RecordTypeCNAME,
	RecordTypeDNAME,
	RecordTypeTXT,
	RecordTypePTR,
	RecordTypeSRV,
}

type NameServer struct {
	Name string `json:"name"`
	Addr net.IP `json:"addr"`
}

type Record struct {
	Name  string     `json:"name"`
	Type  RecordType `json:"type"`
	Value string     `json:"value"`
}

type SOA struct {
	Serial     uint
	Refresh    time.Duration
	Retry      time.Duration
	Expire     time.Duration
	MinimumTTL time.Duration
}

type Zone struct {
	Origin string
	TTL    time.Duration
	SOA    *SOA

	Name        string        `json:"name"`
	NameServers []*NameServer `json:"name-servers"`
	Records     []*Record     `json:"records"`
}

func (z *Zone) ZoneName() string {
	if z.Name == "@" {
		return z.Origin
	}

	return z.Name
}

func (z *Zone) recordByNameType(name string, t RecordType) *Record {
	for _, r := range z.Records {
		if r.Name == name && r.Type == t {
			return r
		}
	}

	return nil
}

func (z *Zone) addRecord(name string, t RecordType, value string) {
	z.Records = append(z.Records, &Record{
		Name:  name,
		Type:  t,
		Value: value,
	})
}

func (z *Zone) deleteRecordByNameType(name string, t RecordType) {
	for i, r := range z.Records {
		if r.Name == name && r.Type == t {
			z.Records = append(z.Records[:i], z.Records[i+1:]...)
			return
		}
	}
}

type PluginSettings struct {
	BindVarFolder    string `json:"bind-var-folder"`
	MasterFolder     string `json:"master-folder"`
	ConfigFile       string `json:"config-file"`
	SystemctlService string `json:"systemctl-service"`

	Zones []*Zone `json:"zones"`
}

func defaultSettings() PluginSettings {
	return PluginSettings{
		BindVarFolder:    "/var/named",
		MasterFolder:     "master",
		ConfigFile:       "/var/named/config.cfg",
		SystemctlService: "named",
	}
}

func (ps *PluginSettings) zoneExist(name string) bool {
	for _, z := range ps.Zones {
		if name == z.ZoneName() {
			return true
		}
	}

	return false
}

func (ps *PluginSettings) addZone(name string, nameServers ...*NameServer) *Zone {
	zone := &Zone{
		Origin: name,
		TTL:    1 * time.Hour,
		Name:   "@",

		SOA: &SOA{
			Serial:     uint(time.Now().Unix()),
			Refresh:    8 * time.Hour,
			Retry:      30 * time.Minute,
			Expire:     24 * 7 * time.Hour,
			MinimumTTL: 1 * time.Hour,
		},

		NameServers: nameServers,
	}

	ps.Zones = append(ps.Zones, zone)

	return zone
}

func (ps *PluginSettings) deleteZone(name string) {
	for i, z := range ps.Zones {
		if name == z.ZoneName() {
			ps.Zones = append(ps.Zones[:i], ps.Zones[i+1:]...)
			break
		}
	}
}

func (ps *PluginSettings) zoneByName(name string) *Zone {
	for _, z := range ps.Zones {
		if name == z.ZoneName() {
			return z
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
	return "dns"
}

func (p *Plugin) Title() string {
	return "DNS"
}

func (p *Plugin) Icon() string {
	return "house"
}

func (p *Plugin) Run(ctx context.Context, api PluginAPI) error {
	p.api = api
	p.ctx = ctx

	p.settings = defaultSettings()

	err := p.api.LoadModuleConfig(&p.settings)
	if err != nil {
		return err
	}

	for _, dir := range []string{p.settings.BindVarFolder, p.settings.MasterFolder} {
		err = os.MkdirAll(dir, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func validAndParseNameServerAddr(name string, value string, input *Input) *NameServer {
	if value == "" {
		input.SetErrorText("IP address must be not empty")
		return nil
	}

	input.SetValue(value)

	ip := net.ParseIP(value)
	if ip == nil {
		input.SetErrorText("incorrect IP address")
		return nil
	}

	return &NameServer{
		Name: name,
		Addr: ip,
	}
}

func (p *Plugin) Actions() ActionsMap {
	return ActionsMap{
		"select-zone": func(args []string, data io.Reader) ActionResult {
			return NewSetArgsActionResult(true, args...)
		},

		"add-new-dns-zone": func(args []string, data io.Reader) ActionResult {
			nameInput := NewInput("zone-name")

			nameServer1 := NewInput("ns1")
			nameServer2 := NewInput("ns2")

			form := NewForm()
			form.AddWithTitle("Zone name", nameInput)
			form.AddWithTitle("Name server 1", nameServer1)
			form.AddWithTitle("Name server 2", nameServer2)
			form.AddActionButtons(NewButton("Add", "add-new-dns-zone", "add"))

			formResponse := NewFormModalActionResult("Add new zone", form)

			if len(args) > 0 {
				reqData := struct {
					ZoneName string `json:"zone-name"`

					NS1 string `json:"ns1"`
					NS2 string `json:"ns2"`
				}{}

				err := json.NewDecoder(data).Decode(&reqData)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				nameInput.SetValue(reqData.ZoneName)

				if reqData.ZoneName == "" {
					nameInput.SetErrorText("zone must be not empty")
					return formResponse
				}

				if p.settings.zoneExist(reqData.ZoneName) {
					nameInput.SetErrorText("zone exist")
					return formResponse
				}

				ns1 := validAndParseNameServerAddr("ns1", reqData.NS1, nameServer1)
				ns2 := validAndParseNameServerAddr("ns2", reqData.NS2, nameServer2)

				if ns1 == nil || ns2 == nil {
					return formResponse
				}

				zone := p.settings.addZone(reqData.ZoneName, ns1, ns2)

				err = p.api.SaveModuleConfig(&p.settings)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				err = p.updateZone(zone)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				return NewReloadActionResult()
			}

			return formResponse
		},

		"delete-zone": func(args []string, data io.Reader) ActionResult {
			zoneName := args[0]
			confirm := "confirm" == args[1]

			if confirm {
				return NewModalActionResult(
					fmt.Sprintf("Delete zone %s", zoneName),
					NewLabel("Do you sure about this?"),
					NewButton("Delete", "delete-zone", zoneName, "").SetStyle(StyleDanger),
					NewButton("Cancel", "none").SetStyle(StyleSecondary),
				)
			}

			p.settings.deleteZone(zoneName)

			err := p.api.SaveModuleConfig(&p.settings)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			err = p.deleteZone(zoneName)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			return NewReloadActionResult()
		},

		"add-zone-record": func(args []string, data io.Reader) ActionResult {
			zoneName := args[0]

			zone := p.settings.zoneByName(zoneName)
			if zone == nil {
				return NewErrorAlertActionResult(fmt.Errorf("faile to get zone [%s]", zoneName))
			}

			recordForm, stdForm := newZoneRecordForm(nil, "add-zone-record", zoneName, "add")
			formResponse := NewFormModalActionResult("Add new zone", stdForm)

			if len(args) > 1 {
				reqData, err := recordForm.parseAndValidate(data)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				if reqData == nil {
					return formResponse
				}

				if r := zone.recordByNameType(reqData.Name, reqData.Type); r != nil {
					recordForm.setExistError()
					return formResponse
				}

				zone.addRecord(reqData.Name, reqData.Type, reqData.Value)

				err = p.api.SaveModuleConfig(&p.settings)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				err = p.updateZone(zone)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				return NewReloadActionResult()
			}

			return formResponse
		},

		"edit-record": func(args []string, data io.Reader) ActionResult {
			zoneName := args[0]

			zone := p.settings.zoneByName(zoneName)
			if zone == nil {
				return NewErrorAlertActionResult(fmt.Errorf("faile to get zone [%s]", zoneName))
			}

			recordName := args[1]
			recordType := RecordType(args[2])

			record := zone.recordByNameType(recordName, recordType)
			if record == nil {
				return NewErrorAlertActionResult(fmt.Errorf("faile to get record [%s] with zone [%s]", recordName, recordType))
			}

			recordForm, stdForm := newZoneRecordForm(record, "edit-record", zoneName, recordName, string(recordType), "save")
			formResponse := NewFormModalActionResult("Add new zone", stdForm)

			if len(args) > 3 {
				reqData, err := recordForm.parseAndValidate(data)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				if reqData == nil {
					return formResponse
				}

				record.Value = reqData.Value

				err = p.api.SaveModuleConfig(&p.settings)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				err = p.updateZone(zone)
				if err != nil {
					return NewErrorAlertActionResult(err)
				}

				return NewReloadActionResult()
			}

			return formResponse
		},

		"delete-record": func(args []string, data io.Reader) ActionResult {
			zoneName := args[0]

			zone := p.settings.zoneByName(zoneName)
			if zone == nil {
				return NewErrorAlertActionResult(fmt.Errorf("faile to get zone [%s]", zoneName))
			}

			recordName := args[1]
			recordType := RecordType(args[2])

			confirm := "confirm" == args[3]

			if confirm {
				return NewModalActionResult(
					fmt.Sprintf("Delete record %s [%s]", recordName, recordType),
					NewLabel("Do you sure about this?"),
					NewButton("Delete", "delete-record", zoneName, recordName, string(recordType), "").SetStyle(StyleDanger),
					NewButton("Cancel", "none").SetStyle(StyleSecondary),
				)
			}

			zone.deleteRecordByNameType(recordName, recordType)

			err := p.api.SaveModuleConfig(&p.settings)
			if err != nil {
				return NewErrorAlertActionResult(err)
			}

			err = p.updateZone(zone)
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

type zoneRecordForm struct {
	nameInput  *Input
	typeInput  *Select
	valueInput *Input
}

func newZoneRecordForm(record *Record, action string, args ...string) (*zoneRecordForm, *Form) {
	nameInput := NewInput("name")
	typeInput := NewSelect("type")
	valueInput := NewInput("value")

	for _, t := range allType {
		typeInput.AddOption(string(t))
	}

	if record != nil {
		nameInput.SetValue(record.Name)
		typeInput.SetValue(string(record.Type))
		valueInput.SetValue(record.Value)

		nameInput.SetDisable(true)
		typeInput.SetDisable(true)
	}

	form := NewForm()
	form.AddWithTitle("Name", nameInput)
	form.AddWithTitle("Record type", typeInput)
	form.AddWithTitle("Value", valueInput)
	form.AddActionButtons(NewButton("Add", action, args...))

	return &zoneRecordForm{
		nameInput:  nameInput,
		typeInput:  typeInput,
		valueInput: valueInput,
	}, form
}

type recordData struct {
	Name  string     `json:"name"`
	Type  RecordType `json:"type"`
	Value string     `json:"value"`
}

func (f *zoneRecordForm) setExistError() {
	f.nameInput.SetErrorText("record with type exist")
}

func (f *zoneRecordForm) parseAndValidate(data io.Reader) (*recordData, error) {
	reqData := &recordData{}

	err := json.NewDecoder(data).Decode(reqData)
	if err != nil {
		return nil, err
	}

	f.nameInput.SetValue(reqData.Name)
	f.typeInput.SetValue(string(reqData.Type))
	f.valueInput.SetValue(reqData.Value)

	if reqData.Name == "" {
		f.nameInput.SetErrorText("must be not empty")
		return nil, nil
	}

	if reqData.Value == "" {
		f.valueInput.SetErrorText("must be not empty")
		return nil, nil
	}

	return reqData, nil
}

func (p *Plugin) Render(args []string) Page {
	if len(args) > 0 {
		devName := args[0]
		return p.renderZone(devName)
	}

	zonesTable := NewTable("name", "")

	for _, z := range p.settings.Zones {
		zonesTable.AddLine(
			NewButton(z.ZoneName(), "select-zone", z.ZoneName()).SetLinkStyle(),
			NewImageButton("trash", "delete-zone", z.ZoneName(), "confirm").SetLinkStyle(),
		)
	}

	return NewPage(
		"DNS Zones",
		NewButton("Add DNS zone", "add-new-dns-zone"),
		NewHeader("Zones"),
		zonesTable,
	)
}

func (p *Plugin) renderZone(zoneName string) Page {
	if zone := p.settings.zoneByName(zoneName); zone != nil {
		recordsTable := NewTable("name", "type", "value", "")

		for _, r := range zone.Records {
			dd := NewDropdown()
			dd.AddItem("pencil", "Edit", "edit-record", zoneName, r.Name, string(r.Type))
			dd.AddSeparator()
			dd.AddDangerItem("trash", "Delete", "delete-record", zoneName, r.Name, string(r.Type), "confirm")

			recordsTable.AddLine(
				NewLabel(r.Name),
				NewBadge(string(r.Type)),
				NewLabel(r.Value).SetMonospace(true),
				dd,
			)
		}

		nameServersInfo := NewElementsList().SetModeLine()

		for _, ns := range zone.NameServers {
			editor := NewInputEdit(ns.Name, ns.Addr.String(), "update-name-server")
			nameServersInfo.AddElementWithTitle(NewLabel("%s.%s", ns.Name, zone.ZoneName()).SetStrong(true), editor)
		}

		return NewPage(
			fmt.Sprintf("Zone %s", zoneName),
			NewButton("Back", "select-zone"),
			NewHeader(fmt.Sprintf("Information about zone %s", zoneName)),
			nameServersInfo,
			NewHeader(fmt.Sprintf("Records of zone %s", zoneName)),
			NewButton("Add record", "add-zone-record", zoneName),
			recordsTable,
		)
	}

	return NewPage(
		fmt.Sprintf("Zone %s", zoneName),
		NewLabel("Not found"),
	)
}

func (p *Plugin) updateZone(z *Zone) error {
	z.SOA.Serial = uint(time.Now().Unix())
	err := updateZone(path.Join(p.settings.BindVarFolder, p.settings.MasterFolder), z)
	if err != nil {
		return err
	}

	return p.update()
}

func (p *Plugin) deleteZone(zoneName string) error {
	err := os.Remove(path.Join(p.settings.BindVarFolder, p.settings.MasterFolder, zoneName))
	if err != nil {
		return err
	}

	return p.update()
}

func (p *Plugin) update() error {
	err := updateConfig(p.settings.ConfigFile, p.settings.MasterFolder, p.settings.Zones)
	if err != nil {
		return err
	}

	if p.settings.SystemctlService != "" {
		return exec.Command("systemctl", "reload", "named").Run()
	}

	return nil
}
