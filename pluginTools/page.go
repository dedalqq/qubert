package pluginTools

type Page struct {
	Title    string        `json:"title"`
	Elements *ElementsList `json:"elements"`
}

func (p *Page) AddElements(elements ...Element) *Page {
	p.Elements.AddElements(elements...)

	return p
}

func NewPage(title string, elements ...Element) Page {
	return Page{
		Title:    title,
		Elements: NewElementsList().AddElements(elements...),
	}
}
