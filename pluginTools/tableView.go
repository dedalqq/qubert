package pluginTools

type TableViewOptions struct {
	Header       []*TableHeader `json:"header"`
	Body         [][]string     `json:"body"`
	SelectAction *Action        `json:"select-action"`
	DataItem     int            `json:"data-item"`
}

type TableView struct {
	options TableViewOptions
}

func (t *TableView) Type() ElementType            { return ElementTypeTableView }
func (t *TableView) MarshalJSON() ([]byte, error) { return MarshalJSON(t.Type(), t.options) }

type TableHeader struct {
	HeaderType string `json:"type"`
	Title      string `json:"title"`
	BodyItems  []int  `json:"items"`
	Proportion int    `json:"proportion"`
	Width      string `json:"width"`
}

func NewTableHeaderText(title string, textItem int) *TableHeader {
	return &TableHeader{
		HeaderType: "text",
		Title:      title,
		BodyItems:  []int{textItem},
	}
}

func NewTableHeaderButton(title string, textItem, styleItem, cmdItem int, argsItem ...int) *TableHeader {
	return &TableHeader{
		HeaderType: "button",
		Title:      title,
		BodyItems:  append([]int{textItem, styleItem, cmdItem}, argsItem...),
	}
}

func NewTableHeaderIcon(title string, iconItem int) *TableHeader {
	return &TableHeader{
		HeaderType: "icon",
		Title:      title,
		BodyItems:  []int{iconItem},
	}
}

func NewTableView(header ...*TableHeader) *TableView {
	return &TableView{
		options: TableViewOptions{
			Header: header,
		},
	}
}

func (t *TableView) AddLine(elements ...string) *TableView {
	t.options.Body = append(t.options.Body, elements)

	return t
}

func (t *TableView) SetSelectAction(dataItem int, cmd string, args ...string) *TableView {
	t.options.DataItem = dataItem
	t.options.SelectAction = &Action{
		CMD:  cmd,
		Args: args,
	}

	return t
}
