package pluginTools

type DropdownItem struct {
	LabelData
	Action

	Danger bool `json:"danger,omitempty"`
}

type DropdownOptions struct {
	Items []*DropdownItem `json:"items"`
}

type Dropdown struct {
	options DropdownOptions
}

func (d *Dropdown) Type() ElementType            { return ElementDropdown }
func (d *Dropdown) MarshalJSON() ([]byte, error) { return MarshalJSON(d.Type(), d.options) }

func (d *Dropdown) add(icon string, text string, danger bool, cmd string, args ...string) *Dropdown {
	d.options.Items = append(d.options.Items, &DropdownItem{
		LabelData: LabelData{
			Icon: icon,
			Text: text,
		},
		Action: Action{
			CMD:  cmd,
			Args: args,
		},
		Danger: danger,
	})

	return d
}

func (d *Dropdown) AddItem(icon string, text string, cmd string, args ...string) *Dropdown {
	return d.add(icon, text, false, cmd, args...)
}

func (d *Dropdown) AddDangerItem(icon string, text string, cmd string, args ...string) *Dropdown {
	return d.add(icon, text, true, cmd, args...)
}

func (d *Dropdown) AddSeparator() *Dropdown {
	d.options.Items = append(d.options.Items, nil)

	return d
}

func NewDropdown() *Dropdown {
	return &Dropdown{
		options: DropdownOptions{},
	}
}
