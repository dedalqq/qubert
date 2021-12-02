package pluginTools

type TableOptions struct {
	Header []string    `json:"header"`
	Body   [][]Element `json:"body"`
}

type Table struct {
	options TableOptions
}

func (t *Table) Type() ElementType            { return ElementTypeTable }
func (t *Table) MarshalJSON() ([]byte, error) { return MarshalJSON(t.Type(), t.options) }

func (t *Table) AddLine(elements ...Element) *Table {
	t.options.Body = append(t.options.Body, elements)

	return t
}

func NewTable(header ...string) *Table {
	return &Table{
		options: TableOptions{
			Header: header,
		},
	}
}
