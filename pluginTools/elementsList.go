package pluginTools

type ElementListItem struct {
	Title Element `json:"title,omitempty"`
	Item  Element `json:"item"`
}

type ElementListsOptions struct {
	Mode     string            `json:"mode,omitempty"`
	Elements []ElementListItem `json:"elements"`
}

type ElementsList struct {
	options ElementListsOptions
}

func (el *ElementsList) Type() ElementType            { return ElementTypeElementsList }
func (el *ElementsList) MarshalJSON() ([]byte, error) { return MarshalJSON(el.Type(), el.options) }

func (el *ElementsList) SetModeLine() *ElementsList {
	el.options.Mode = "line"

	return el
}

func (el *ElementsList) SetModeDefault() *ElementsList {
	el.options.Mode = ""

	return el
}

func (el *ElementsList) AddElements(elements ...Element) *ElementsList {
	for _, e := range elements {
		el.options.Elements = append(el.options.Elements, ElementListItem{
			Item: e,
		})
	}
	return el
}

func (el *ElementsList) AddElementWithTitle(title Element, element Element) *ElementsList {
	el.options.Elements = append(el.options.Elements, ElementListItem{
		Title: title,
		Item:  element,
	})

	return el
}

func NewElementsList() *ElementsList {
	return &ElementsList{
		options: ElementListsOptions{
			Elements: []ElementListItem{},
		},
	}
}
