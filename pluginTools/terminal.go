package pluginTools

type TerminalOption struct {
}

type Terminal struct {
	options TerminalOption
}

func (t *Terminal) Type() ElementType            { return ElementTerminal }
func (t *Terminal) MarshalJSON() ([]byte, error) { return MarshalJSON(t.Type(), t.options) }

func NewTerminal() *Terminal {
	return &Terminal{
		options: TerminalOption{},
	}
}
