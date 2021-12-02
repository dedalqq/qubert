package pluginTools

type HeaderOptions struct {
	Text string `json:"text"`
}

type Header struct {
	options HeaderOptions
}

func (h *Header) Type() ElementType            { return ElementTypeHeader }
func (h *Header) MarshalJSON() ([]byte, error) { return MarshalJSON(h.Type(), h.options) }

func NewHeader(text string) *Header {
	return &Header{
		options: HeaderOptions{
			Text: text,
		},
	}
}
