package pluginTools

type ActionType string

const (
	ActionTypeReload     ActionType = "reload"
	ActionTypeModal                 = "modal"
	ActionTypeArgs                  = "set-args"
	ActionTypeAlert                 = "alert"
	ActionTypePartUpdate            = "part-update"
)

type ActionResult struct {
	ActionType ActionType  `json:"type"`
	Options    interface{} `json:"options"`
}

type ActionResultReloadOptions struct {
	Fade bool `json:"fade"`
}

func NewReloadActionResult() ActionResult {
	return ActionResult{
		ActionType: ActionTypeReload,
		Options: ActionResultReloadOptions{
			Fade: true,
		},
	}
}

func NewReloadWithoutFadeActionResult() ActionResult {
	return ActionResult{
		ActionType: ActionTypeReload,
		Options:    ActionResultReloadOptions{},
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
		ActionType: ActionTypeAlert,
		Options: ActionResultAlertOptions{
			Title: title,
			Text:  text,
		},
	}
}

func NewErrorAlertActionResult(err error) ActionResult {
	return ActionResult{
		ActionType: ActionTypeAlert,
		Options: ActionResultAlertOptions{
			Title: "Error",
			Text:  err.Error(),
		},
	}
}

type ActionResultSetArgsOptions struct {
	Args []string `json:"args"`
	Fade bool     `json:"fade"`
}

func NewSetArgsActionResult(fade bool, args ...string) ActionResult {
	return ActionResult{
		ActionType: ActionTypeArgs,
		Options: ActionResultSetArgsOptions{
			Fade: fade,
			Args: args,
		},
	}
}

type ActionResultPartUpdateOptions struct {
	ElementID string  `json:"id"`
	Element   Element `json:"element"`
}

func NewPartUpdateActionResult(elementID string, el Element) ActionResult {
	return ActionResult{
		ActionType: ActionTypePartUpdate,
		Options: ActionResultPartUpdateOptions{
			ElementID: elementID,
			Element:   el,
		},
	}
}
