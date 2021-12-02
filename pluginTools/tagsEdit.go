package pluginTools

type TagsEditOptions struct {
	Action Action   `json:"action"`
	Name   string   `json:"name"`
	Value  []string `json:"value"`
}

type TagsEdit struct {
	options TagsEditOptions
}

func (t *TagsEdit) Type() ElementType            { return ElementTagsEdit }
func (t *TagsEdit) MarshalJSON() ([]byte, error) { return MarshalJSON(t.Type(), t.options) }

func NewTagsEdit(name string, value []string, cmd string, args ...string) *TagsEdit {
	return &TagsEdit{
		options: TagsEditOptions{
			Action: Action{
				CMD:  cmd,
				Args: args,
			},
			Name:  name,
			Value: value,
		},
	}
}
