package pluginTools

type CodeEditorOptions struct {
	Name      string `json:"name"`
	ElementID string `json:"id"`
	Value     string `json:"value"`
}

type CodeEditor struct {
	options CodeEditorOptions
}

func (e *CodeEditor) Type() ElementType            { return ElementTypeCodeEditor }
func (e *CodeEditor) MarshalJSON() ([]byte, error) { return MarshalJSON(e.Type(), e.options) }

func (e *CodeEditor) ID() string {
	return e.options.ElementID
}

func NewCodeEditor(name string, value string) *CodeEditor {
	return &CodeEditor{
		options: CodeEditorOptions{
			ElementID: name,
			Name:      name,
			Value:     value,
		},
	}
}
