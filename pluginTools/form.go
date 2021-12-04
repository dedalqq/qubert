package pluginTools

type FormOptions struct {
	FormElements *ElementsList `json:"elements"`
	Actions      []*Button     `json:"actions"`
}

type Form struct {
	options FormOptions
}

func (f *Form) Type() ElementType            { return ElementTypeForm }
func (f *Form) MarshalJSON() ([]byte, error) { return MarshalJSON(f.Type(), f.options) }

func (f *Form) AddActionButtons(button ...*Button) *Form {
	f.options.Actions = append(f.options.Actions, button...)

	return f
}

func (f *Form) Add(elements ...Element) *Form {
	f.options.FormElements.AddElements(elements...)

	return f
}

func (f *Form) AddWithTitle(title string, element FormElement) *Form {
	f.options.FormElements.AddElementWithTitle(NewFormLabel(title).SetFor(element.ID()), element)

	return f
}

func NewForm() *Form {
	return &Form{
		options: FormOptions{
			FormElements: NewElementsList(),
			Actions:      []*Button{},
		},
	}
}
