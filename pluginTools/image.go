package pluginTools

const BSImageFile = "bootstrap-icons.svg"

type ImageOptions struct {
	FileSVG   string `json:"svg"`
	ImageName string `json:"name"`
	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
}

type Image struct {
	options ImageOptions
}

func (i *Image) Type() ElementType            { return ElementTypeImage }
func (i *Image) MarshalJSON() ([]byte, error) { return MarshalJSON(i.Type(), i.options) }

func NewBSImage(name string) *Image {
	return &Image{
		options: ImageOptions{
			FileSVG:   BSImageFile,
			ImageName: name,
			Width:     16,
			Height:    16,
		},
	}
}
