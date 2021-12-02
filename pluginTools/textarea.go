package pluginTools

type TextareaOptions struct {
	Name      string `json:"name"`
	ElementID string `json:"id"`
	Value     string `json:"value,omitempty"`
	Error     string `json:"error,omitempty"`
}

type Textarea struct {
	options TextareaOptions
}

func (t *Textarea) Type() ElementType            { return ElementTypeTextarea }
func (t *Textarea) MarshalJSON() ([]byte, error) { return MarshalJSON(t.Type(), t.options) }

func (t *Textarea) ID() string {
	return t.options.ElementID
}

func (t *Textarea) SetID(id string) *Textarea {
	t.options.ElementID = id

	return t
}

func (t *Textarea) SetValue(v string) *Textarea {
	t.options.Value = v

	return t
}

func (t *Textarea) SetErrorText(text string) *Textarea {
	t.options.Error = text

	return t
}

func NewTextarea(name string) *Textarea {
	return &Textarea{
		options: TextareaOptions{
			Name:      name,
			ElementID: name,
		},
	}
}
