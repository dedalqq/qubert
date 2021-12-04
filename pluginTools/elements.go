package pluginTools

import (
	"encoding/json"
	"io"
)

type PluginAPI interface {
	SaveModuleConfig(cfg interface{}) error
	LoadModuleConfig(cfg interface{}) error
	Send(data interface{}, args ...string)
	SendUpdate(updateData Update, args ...string) bool
	Reload(args ...string)
	SafeRun(f func())
	Exit()
	Shutdown() error
	Restart() error
	Version() (string, string)
}

type ActionsMap map[string]func(args []string, data io.Reader) ActionResult

type ElementStyle string

const (
	StylePrimary   ElementStyle = "primary"
	StyleSecondary              = "secondary"
	StyleSuccess                = "success"
	StyleDanger                 = "danger"
	StyleWarning                = "warning"
	StyleInfo                   = "info"
	StyleLight                  = "light"
	StyleDark                   = "dark"
)

type ElementType string

const (
	ElementTypeButton       ElementType = "button"
	ElementTypeText                     = "text"
	ElementTypeLabel                    = "label"
	ElementTypeFormLabel                = "form-label"
	ElementTypeInput                    = "input"
	ElementTypeSelect                   = "select"
	ElementTypeTextarea                 = "textarea"
	ElementSwitch                       = "switch"
	ElementTypeElementsList             = "element-list"
	ElementTypeForm                     = "form"
	ElementTypeTable                    = "table"
	ElementTypeBadge                    = "badge"
	ElementTypeImage                    = "image"
	ElementTypeHeader                   = "header"
	ElementTypeCodeEditor               = "codeEditor"
	ElementTypeCard                     = "card"
	ElementDropdown                     = "dropdown"
	ElementInputEdit                    = "input-edit"
	ElementTagsEdit                     = "tags-edit"
	ElementTextareaEdit                 = "textarea-edit"
	ElementLine                         = "line"
	ElementProgress                     = "progress"
	ElementTerminal                     = "terminal"
)

type Element interface {
	Type() ElementType
}

type FormElement interface {
	Element

	ID() string
}

type Update interface {
	ElementID() string
	UpdateType() ElementType
}

type LabelData struct {
	Text   string `json:"text"`
	Icon   string `json:"icon,omitempty"`
	Num    int    `json:"num,omitempty"`
	Strong bool   `json:"strong,omitempty"`
}

type ElementStruct struct {
	ElementType ElementType `json:"type"`
	Options     interface{} `json:"options"`
}

type Action struct {
	CMD  string   `json:"cmd"`
	Args []string `json:"args,omitempty"`
}

func MarshalJSON(elementType ElementType, option interface{}) ([]byte, error) {
	return json.Marshal(ElementStruct{
		ElementType: elementType,
		Options:     option,
	})
}
