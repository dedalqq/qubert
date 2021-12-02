package pluginTools

type ActionType string

const (
	ActionTypeReload ActionType = "reload"
	ActionTypeModal             = "modal"
	ActionSetArgs               = "set-args"
	ActionAlert                 = "alert"
)

type ActionResult struct {
	ActionType ActionType  `json:"type"`
	Options    interface{} `json:"options"`
}

func NewReloadActionResult() ActionResult {
	return ActionResult{
		ActionType: ActionTypeReload,
	}
}

type ActionResultModalOptions struct {
	Title   LabelData `json:"title"`
	Element Element   `json:"content"`
	Actions []*Button `json:"actions"`
}

func NewModalActionResult(title string, element Element, actions ...*Button) ActionResult {
	return ActionResult{
		ActionType: ActionTypeModal,
		Options: ActionResultModalOptions{
			Title: LabelData{
				Text: title,
				Icon: "info-circle",
			},
			Element: element,
			Actions: actions,
		},
	}
}

func NewFormModalActionResult(title string, form *Form) ActionResult {
	return ActionResult{
		ActionType: ActionTypeModal,
		Options: ActionResultModalOptions{
			Title: LabelData{
				Text: title,
				Icon: "info-circle",
			},
			Element: form.options.FormElements,
			Actions: form.options.Actions,
		},
	}
}

type ActionResultAlertOptions struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

func NewAlertActionResult(title string, text string) ActionResult {
	return ActionResult{
		ActionType: ActionAlert,
		Options: ActionResultAlertOptions{
			Title: title,
			Text:  text,
		},
	}
}

func NewErrorAlertActionResult(err error) ActionResult {
	return ActionResult{
		ActionType: ActionAlert,
		Options: ActionResultAlertOptions{
			Title: "Error",
			Text:  err.Error(),
		},
	}
}

type ActionResultSetArgsOptions struct {
	Args []string `json:"args"`
}

func NewSetArgsActionResult(args ...string) ActionResult {
	return ActionResult{
		ActionType: ActionSetArgs,
		Options: ActionResultSetArgsOptions{
			Args: args,
		},
	}
}
