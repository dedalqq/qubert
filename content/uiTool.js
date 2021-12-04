let uiTool = {
    button: function(options) {
        let element = "button"
        let classes = []

        if (options.style === "link") {
            element = "a"
        } else {
            classes.push("btn", "btn-sm")

            if (options.style) {
                classes.push(`btn-${options.style}`)
            }
        }

        if (options.action.cmd === "") {
            classes.push("disabled")
        }

        return {
            tag: element,
            classes: classes,
            onclick: core.actionFunc(options.action),
            el: uiTool.createLabelElements(options)
        }
    },

    text: function(options) {
        return {tag: "p", text: options.text}
    },

    label: function(options) {
        return {tag: "span", el: uiTool.createLabelElements(options)}
    },

    formLabel: function(options) {
        return {tag: "label", classes: ["form-label"], text: options.text, for: options.for}
    },

    form: function(options) {

        let elements = [
            uiTool.createElement(options.elements),
        ]

        for (const a of options.actions) {
            elements.push(uiTool.createElement(a), {text: " "})
        }

        return {tag: "form", el: elements}
    },

    input: function(options) {
        let classes = ["form-control"]

        if (options.error !== undefined) {
            classes.push("is-invalid")
        }

        let elements = [{
            tag: "input",
            classes: classes,
            name: options.name,
            id: options.id,
            type: options.type,
            value: options.value,
            autocomplete: options.name,
        }]

        if (options.error !== undefined) {
            elements.push({tag: "div", classes: ["invalid-feedback"], text: options.error})
        }

        return {tag: "div", el: elements}
    },

    select: function (options) {
        let selectOptions = []

        for (const v in options.options) {
            let opt = {tag: "option", value: v, text: options.options[v]}

            if (options.value === v) {
                opt.selected = "selected"
            }

            selectOptions.push(opt)
        }

        let select = {
            tag: "select",
            classes: ["form-control"],
            name: options.name,
            id: options.id,
            el: selectOptions,
        }

        if (options["change-action"]) {
            select.onchange = core.actionFunc(options["change-action"])
        }

        return select
    },

    textarea: function(options) {
        let classes = ["form-control"]

        if (options.error !== undefined) {
            classes.push("is-invalid")
        }

        let elements = [{
            tag: "textarea",
            classes: classes,
            name: options.name,
            id: options.id,
            text: options.value,
            autocomplete: false,
        }]

        if (options.error !== undefined) {
            elements.push({tag: "div", classes: ["invalid-feedback"], text: options.error})
        }

        return {tag: "div", el: elements}
    },

    switch: function (options) {
        return {tag: "div", classes: ["form-check", "form-switch"], el: [
            {
                tag: "input",
                classes: ["form-check-input"],
                name: options.name,
                id: options.id,
                type: "checkbox",
                role: "switch",
            }
        ]}
    },

    elementList: function(options) {
        let elements = []

        switch (options.mode) {
            case "line":
                for (const el of options.elements) {
                    let subElements = []

                    if (el.title !== undefined) {
                        subElements.push({tag: "div", classes: ["col-4"], el: [
                            uiTool.createElement(el.title),
                        ]})
                    }

                    subElements.push({tag: "div", classes: ["col-8"], el: [
                        uiTool.createElement(el.item),
                    ]})

                    elements.push({tag: "div", classes: ["row"], el: subElements})
                }

                break
            default:
                for (const el of options.elements) {
                    let subElements = []

                    if (el.title !== undefined) {
                        subElements.push(uiTool.createElement(el.title))
                    }

                    subElements.push(uiTool.createElement(el.item))

                    elements.push({tag: "div", classes: ["mb-3"], el: subElements})
                }

                break
        }

        return {tag: "div", el: elements}
    },

    table: function(options) {
        let header = []

        for (const h of options.header) {
            header.push({tag: "th", text: h})
        }

        let body = []

        if (options.body) {
            for (const line of options.body) {
                let lineItems = []

                for (const item of line) {
                    lineItems.push({tag: "td", el: [uiTool.createElement(item)]})
                }

                body.push({tag: "tr", el: lineItems})
            }
        }

        return {tag: "table", classes: ["table", "table-hover"], el: [
            {tag: "thead", el: [
                    {tag: "tr", el: header},
                ]},
            {tag: "tbody", el: body},
        ]}
    },

    badge: function(options) {
        return {tag: "span", classes: ["badge", `bg-${options.style}`], text: options.text}
    },

    image: function(options) {
        return {tag: "svg", width: options.width, height: options.height, use: `${options.svg}#${options.name}`}
    },

    icon: function(options) {
        return {tag: "i", classes: ["bi", options.name]}
    },

    header: function(options) {
        return {tag: "h3", text: options.text}
    },

    codeEditor: function(options) {
        let hiddenInput

        return {tag: "div", el: [
            {tag: "input", type: "hidden", name: options.name, cb: function (e) {
                hiddenInput = e
                hiddenInput.value = options.value
            }},
            {tag: "div", id: "editor", classes: ["form-control"], cb: function(e) {
                let editor = ace.edit(e);

                editor.setValue(options.value, -1)
                editor.on("change", function (e) { hiddenInput.value = editor.getValue() })
            }}
        ]}
    },

    card: function(options) {
        let headerElements = [
            {tag: "div", classes: ["me-auto"], el: uiTool.createLabelElements(options.header)}
        ]

        if (options.additional) {
            headerElements.push(uiTool.createElement(options.additional))
        }

        return {tag: "div", classes: ["card", "shadow-sm"], el: [
            {tag: "div", classes: ["card-header", "d-flex", "flex-row"], el: headerElements},
            {tag: "div", classes: ["card-body"], el: [uiTool.createElement(options.body)]},
        ]}
    },

    dropdown: function(options) {
        let elements = []

        let menu = null

        for (const i of options.items) {
            if (i === null) {
                elements.push({tag: "li", el: [
                    {tag: "hr", classes: ["dropdown-divider"]}
                ]})

                continue
            }

            let classes = ["dropdown-item", "d-flex", "gap-2", "align-items-center"]

            if (i.danger) {
                classes.push("dropdown-item-danger")
            }

            if (i.cmd === "") {
                classes.push("disabled")
            }

            elements.push({tag: "li", el: [
                {
                    tag: "a",
                    classes: classes,
                    el: uiTool.createLabelElements(i),
                    onclick: core.actionFunc(i, function (e, el) {
                        menu.remove()
                        menu = null
                    }),
                    onmousedown: function (e, el) {
                        return false
                    }
                }
            ]})
        }

        let div = null

        return {tag: "div", el: [
            {
                tag: "a",
                onclick: function(e, el) {
                    if (menu) {
                        return false
                    }

                    ui.build({tag: "ul", classes: ["dropdown-menu", "mx-0", "shadow"], el: elements, cb: function(e) {
                        menu = e
                        menu.style.position = "absolute"
                        menu.style.display = "block"
                    }, parent: div})

                    return false
                },
                onblur: function (e, el) {
                    if (menu) {
                        menu.remove()
                        menu = null
                    }

                    return false
                },
                el: [{tag: "i", classes: ["bi", "bi-list"]}]
            },
        ], cb: function (e) {
            div = e
        }}
    },

    textareaEdit: function(options) {
        let renderElement = function () {
            return {tag: "span", text: options.value}
        }

        let renderInput = function (cb) {
            return {
                tag: "textarea",
                classes: ["form-control"],
                id: options.name,
                name: options.name,
                text: options.value,
                cb: cb,
            }
        }

        return uiTool.valueEditor(options.action, renderElement, renderInput)
    },

    inputEdit: function(options) {
        let renderElement = function () {
            return {tag: "span", text: options.value}
        }

        let renderInput = function (cb) {
            return {
                tag: "input",
                classes: ["form-control"],
                id: options.name,
                name: options.name,
                type: "text",
                value: options.value,
                cb: cb,
            }
        }

        return uiTool.valueEditor(options.action, renderElement, renderInput)
    },

    tagsEdit: function(options) {
        let createTag = function (text, editable) {
            let itemElement = null

            let item = {tag: "span", classes: ["badge", "bg-primary", "me-1"], cb: function (e) { itemElement = e }}

            if (editable) {
                item.el = [
                    {text: text},
                    {tag: "i", classes: ["ms-1", "bi", "bi-x-lg"], onclick: function (e, el) {
                            itemElement.remove()

                            return false
                        }, onmousedown: function (e, el) {
                            return false
                        }
                    },
                    {tag: "input", type: "hidden", value: text, name: `[]${options.name}`}
                ]
            } else {
                item.text = text
            }

            return item
        }

        let renderItems = function (editable) {
            let items = []

            if (options.value) {
                for (let text of options.value) {
                    items.push(createTag(text, editable))
                }
            }

            return items
        }

        let renderElement = function () {
            return {tag: "span", el: renderItems(false)}
        }

        let renderInput = function (cb) {
            let items = renderItems(true)

            items.push({tag: "input", name: `[]${options.name}`, type: "text", cb: function (el) {
                el.style.width = `2ch`

                cb(el)
            }, onkeypress: function (e, el) {
                if (e.key !== "Enter" || el.value === "") {
                    return true
                }

                let newTag = ui.build(createTag(el.value, true))
                el.parentNode.insertBefore(newTag, el)
                el.value = ""
                el.style.width = `2ch`
            }, oninput: function (e, el) {
                el.style.width = `${el.value.length+2}ch`

                return true
            }})

            return {tag: "div", classes: ["tags-editor", "form-control"], el: items}
        }

        return uiTool.valueEditor(options.action, renderElement, renderInput)
    },

    line: function(options) {
        let elements = []

        for (const el of options.elements) {
            elements.push({tag: "div", classes: ["me-2"], el: [uiTool.createElement(el)]})
        }

        return {tag: "div", classes: ["d-flex"], el: elements}
    },

    progress: function(options) {
        return {tag: "div", id: options.id, classes: ["progress"], el: [
            {tag: "div", classes: ["progress-bar"]}
        ]}
    },

    terminal: function() {
        return {tag: "div", id: "terminal", cb: function (e) {
            let term = new Terminal();
            term.open(e);
            term.write('Hello from \x1B[1;3;31mxterm.js\x1B[0m $ ')
        }}
    },

    createElement: function(element) {
        switch (element.type) {
            case "button":
                return uiTool.button(element.options);
            case "text":
                return uiTool.text(element.options);
            case "label":
                return uiTool.label(element.options);
            case "form-label":
                return uiTool.formLabel(element.options);
            case "input":
                return uiTool.input(element.options);
            case "select":
                return uiTool.select(element.options);
            case "textarea":
                return uiTool.textarea(element.options);
            case "switch":
                return uiTool.switch(element.options);
            case "element-list":
                return uiTool.elementList(element.options);
            case "form":
                return uiTool.form(element.options);
            case "table":
                return uiTool.table(element.options);
            case "badge":
                return uiTool.badge(element.options);
            case "image":
                return uiTool.image(element.options);
            case "icon":
                return uiTool.icon(element.options);
            case "header":
                return uiTool.header(element.options);
            case "codeEditor":
                return uiTool.codeEditor(element.options);
            case "card":
                return uiTool.card(element.options);
            case "dropdown":
                return uiTool.dropdown(element.options);
            case "input-edit":
                return uiTool.inputEdit(element.options);
            case "tags-edit":
                return uiTool.tagsEdit(element.options);
            case "textarea-edit":
                return uiTool.textareaEdit(element.options);
            case "line":
                return uiTool.line(element.options);
            case "progress":
                return uiTool.progress(element.options);
            case "terminal":
                return uiTool.terminal(element.options);
            default:
                console.warn(`un know element type: ${element.type}`);
        }
    },

    valueEditor: function(action, renderElement, renderInput) {
        let form

        let check = {tag: "a", el: [
            {tag: "i", classes: ["bi", "bi-check-lg", "ms-2", "input-edit"]},
        ], onclick: core.actionFunc(action), onmousedown: function (e, el) {
            return false
        }}

        function render() {
            ui.clear(form)
            ui.build({tag: "div", el: [
                renderElement(),
                {tag: "a", onclick: function (e, el) {
                    ui.clear(form)

                    let inputElement = null

                    form.append(ui.build(renderInput(function (e) {
                        inputElement = e
                        inputElement.onblur = function(e, el) {
                            render()

                            return false
                        }
                        let keypressFunc = inputElement.onkeypress
                        inputElement.onkeypress = function(e, el) {
                            if (e.ctrlKey && e.code === "Enter") {
                                core.actionFunc(action)(e, el)
                            } else if (keypressFunc) {
                                keypressFunc(e, el)
                            }
                        }
                    })), ui.build(check))

                    inputElement.focus()

                    return false
                }, el: [{tag: "i", classes: ["bi-pencil", "ms-2"]}]},
            ], parent: form})
        }

        return {tag: "form", classes: ["d-flex"], cb: function (e) {
            form = e
            render()
        }}
    },

    createLabelElements: function (labelData) {
        let elements = []

        if (labelData.icon) {
            let imageClasses = ["bi", `bi-${labelData.icon}`]

            if (labelData.text) {
                imageClasses.push("me-2")
            }

            elements.push({tag: "i", classes: imageClasses})
        }

        if (labelData.text) {
            if (labelData.strong) {
                elements.push({tag: "strong", text: labelData.text})
            } else {
                elements.push({text: labelData.text})
            }
        }

        return elements
    },
}
