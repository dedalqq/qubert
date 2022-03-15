let uitool = {
    button: function(options, action) {
        let button_classes = ["btn", "btn-sm"];

        if (options.style) {
            button_classes.push(`btn-${options.style}`)
        } else {
            button_classes.push("btn-primary")
        }

        let button = ui.element("button", null, button_classes);

        button.type = "button";

        if (options.icon) {
            button.append(ui.iconText(options.icon, options.text));
        } else {
            button.append(ui.text(options.text));
        }

        if (options.color) {
            button.style.color = options.color
        }

        if (options.disabled) {
            button.disabled = true
        }

        button.onclick = action;

        return button
    },

    table: function(options) {
        let table = ui.element("table", null, ["table", "table-sm"]); // "table-bordered", "scrollable"
        let thead = ui.element("thead", table);
        let tr = ui.element("tr", thead);

        for (const h of options.headers) {
            let c = ui.element("th", tr, ["column_icon"], {scope: "col"});

            if (h.width) {
                c.style.width = `${h.width}px`
            }

            c.append(ui.text(h.title))
        }

        let tbody = ui.element("tbody", table);

        for (const r of options.rows) {
            let tr = ui.element("tr", tbody);
            for (const i in r) {
                if (i >= options.headers.length) {
                    continue
                }

                let cell = ui.element("td", tr);

                switch (options.headers[i].decorator) {
                    case "text":
                        cell.append(ui.text(r[i]));
                        continue;
                    case "control":
                        if (r[i]) {
                            cell.append(uitool.button(r[i], function () {
                                core.action(this, r[i].action.action, r[i].action.args, null);
                            }));
                        }
                        continue;
                    case "badge":
                        if (r[i]) {
                            ui.element("span", cell, ["badge", "badge-danger"]).append(ui.text(r[i]))
                        }
                }
            }
        }

        if (options.card_title) {
            let div = ui.element("div", null, ["card"]);
            ui.element("div", div, ["card-header"]).append(ui.text(options.card_title));
            div.append(table);

            return div
        }

        return table
    },

    text: function(options) {
        let p = ui.element("p");
        ui.text(options.text, p);
        return p
    },

    label: function(options) {
        let p = ui.element("span");
        ui.text(options.text, p);
        return p
    },

    card: function(options) {
        let div = ui.element("div", null, ["card", "shadow-sm"]);

        if (options.width) {
            div.style.width = `${options.width}px`
        }

        let head = ui.element("div", div, ["card-header"]);
        if (options.icon) {
            head.append(ui.iconText(options.icon, options.title, ["mr-auto"]))
        } else {
            ui.element("strong", head, ["mr-auto"]).append(ui.text(options.title));
        }

        if (options.button) {
            let button = core.createElement("button", options.button);
            head.append(button)
        }

        let body = ui.element("div", div, ["card-body"]);

        core.appendElements(options, body);

        return div
    },

    progress: function(options) {
        let div = ui.element("div", null, ["row"]);

        ui.element("label", div, ["col-sm-4"]).append(ui.text(options.title));

        let right_div = ui.element("div", div, ["col-sm-8"]);
        let progress = ui.element("div", right_div, ["progress"]);
        ui.element("div", progress, ["progress-bar"]).style.width = `${options.value}%`;

        return div
    },

    columns: function(options) {
        let div = ui.element("div", null, ["row", `row-cols-${options.columns_num}`]);

        if (!options.elements) {
            return div
        }

        let columns = [];

        for (let i = 0; i < options.columns_num; i++) {
            columns.push(ui.element("div", div, ["col", "column"]));
        }

        for (const i in options.elements) {
            let element = core.createElement(options.elements[i].name, options.elements[i].options);
            ui.appendLine(element, columns[i%options.columns_num])
        }

        return div
    },

    row: function(options) {
        let div = ui.element("div", null, ["elements-row"]);

        for (const el of options.elements) {
            let element = core.createElement(el.name, el.options);
            div.append(element)
        }

        return div
    },

    textViewer: function(options) {
        let div = ui.element("div", null, ["text-viewer"]);
        div.append(ui.text(options.text));
        div.addEventListener("DOMNodeInserted", function () {
            div.scrollTo(0, div.scrollHeight);
        }, false);

        return div
    },

    header: function(options) {
        let h = ui.element("h5", null, ["title"]);
        h.append(ui.text(options.text));
        return h
    },

    attributeList: function(options) {
        let div = ui.element("div", null, []);

        for (const a of options.attributes) {
            let line = ui.element("div", div, ["row"]);

            ui.element("span", line, [`col-${options.name_size}`]).append(ui.text(a.name));
            let value = ui.element("div", line, [`col-${options.element_size}`]);
            value.append(core.createElement(a.element.name, a.element.options));

            if (a.description) {
                ui.element("div", value, ["attr-description"]).append(ui.text(a.description.text))
            }
        }

        return div
    },

    badgeList: function(options) {
        let span = ui.element("span", null, []);

        for (const v of options.badges) {
            let classes = ["badge"];

            if (v.style) {
                classes.push(`badge-${v.style}`)
            } else {
                classes.push("badge-primary")
            }

            ui.element("span", span, classes).append(ui.text(v.text));
            ui.text(" ", span);
        }

        return span
    },

    valueEditor: function(options) {
        let div = ui.element("div", null, []);

        let elementConstructor = uitool.formElementConstructor(options.input.type);

        let render = function() {
            ui.clear(div);

            let value = elementConstructor.render(options.input);

            div.append(value);

            let edit_button = ui.element("button", div, ["btn", "btn-sm", "btn-link"]);

            ui.element("i", edit_button, ["fas", 'fa-pencil-alt']);

            edit_button.onclick = function() {
                ui.clear(div);

                let form = ui.element("form", div, []);
                form.onsubmit = function() { return false };

                let apply_button = ui.element("button", div, ["btn", "btn-sm", "btn-link", "apply-edit-icon"]);

                let input = elementConstructor.input(options.input, null);

                input.addEventListener("focusout", function(e) {
                    if (input.contains(e.relatedTarget) || apply_button === e.relatedTarget) {
                        return
                    }

                    render();
                });

                form.append(input);

                ui.element("i", apply_button, ["fas", "fa-check"]);

                apply_button.onclick = function() {
                    core.action(apply_button, options.action, options.args, uitool.getFormData(form));
                };


                input.focus()
            };
        };

        render();

        return div
    },

    num: 0,

    switcher: function(options) {
        let div = ui.element("div", null, ["custom-control", "custom-switch"]);

        let element_id = `switcher-${options.name}-${this.num}`;

        this.num++;

        let form = ui.element("form", div, []);

        let input = ui.element("input", form, ["custom-control-input"], {
            type: "checkbox", name: options.name, id: element_id,
        });

        input.checked = options.value;

        ui.element("label", form, ["custom-control-label"], {for: element_id});

        input.onclick = function () {
            input.disabled = true;
            core.action(null, options.action.action, options.action.args, uitool.getFormData(form));
        };

        return div
    },

    formElements: function(form_elements, parent) {
        let form = ui.element("form", parent, []);

        for (const form_element of form_elements) {
            this.addFormLine(form, true, form_element)
        }

        return form
    },

    formElementConstructor: function(type) {
        switch (type) {
            case "switcher":
                return this.inputSwitcher;
            case "textarea":
                return this.textarea;
            case "tags-editor":
                return this.tagEditor;
            case "text":
            case "number":
                return this.input;
            case "select":
                return this.select;
        }
    },

    addFormLine: function(form, horizontal, form_element) {
        if (horizontal) {
            let line = ui.element("div", null, ["form-group", "row"]);

            ui.element("label", line, ["col-sm-4", "col-form-label"], {
                for: form_element.input.name
            }).append(ui.text(form_element.title));

            let div = ui.element("div", line, ["col-sm-8", "row"]);

            let element = this.formElementConstructor(form_element.input.type).input(form_element.input, form);

            if (form_element.invalid) {
                element.classList.add("is-invalid")
            }

            div.append(element);

            if (form_element.description) { // TODO duplicate
                ui.element("small", div, ["form-text", "text-muted"]).append(ui.text(form_element.description))
            }

            form.append(line);

            // return
        }

        // let div = ui.element("div", null, ["form-group"]);
        //
        // ui.element("label", div, [], {
        //     for: form_element.input.name
        // }).append(ui.text(form_element.title));
        //
        // let element = this.formElementConstructor(form_element.input.type).input(form_element.input, form)
        //
        // if (form_element.invalid) {
        //     element.classList.add("is-invalid")
        // }
        //
        // div.append(element);
        //
        // if (form_element.description) {
        //     ui.element("small", div, ["form-text", "text-muted"]).append(ui.text(form_element.description))
        // }
        //
        // form.append(div);
    },

    textarea: {
        input: function(element, form) {
            return ui.element("textarea", null, ["form-control", "form-control-sm"], {
                name: element.name, value: element.value
            });
        },

        render: function(element) {
            let span = ui.element("span", null, []);

            let lines = element.value.split("\n");

            for (let i = 0; i < lines.length - 1; i++) {
                span.append(ui.text(lines[i]));
                ui.element("br", span)
            }

            span.append(ui.text(lines[lines.length - 1]));

            return span
        },
    },

    input: {
        input: function (element, form) {
            return ui.element("input", null, ["form-control", "form-control-sm"], {
                type: element.type, name: element.name, autocomplete: "off", value: element.value,
            });
        },

        render: function (element) {
            let span = ui.element("span", null, []);
            span.append(ui.text(element.value));

            return span
        },
    },

    select: {
        input: function(element, form) {
            let select = ui.element("select", null, ["custom-select", "form-control-sm"], {name: element.name});

            for (const option of element.options) {
                let opt = ui.element("option", select, [], {
                    value: option.value,
                });

                opt.append(ui.text(option.text));

                if (option.value === element.value) {
                    opt.selected = true
                }
            }

            if (element.on_change) {
                select.onchange = function() {
                    core.action(null, element.on_change.action, element.on_change.args, uitool.getFormData(form));
                }
            }

            return select
        },

        render: function(element) {
            let span = ui.element("span", null, []);

            let value = element.value;

            for (const option of element.options) {
                if (option.value === element.value) {
                    value = option.text
                }
            }

            span.append(ui.text(value));

            return span
        },
    },

    inputSwitcher: {
        num: 0,

        input: function(element, form) {
            let div = ui.element("div", null, ["custom-control", "custom-switch"]);

            let element_id = `switcher-${element.name}-${this.num}`;

            ui.element("input", div, ["custom-control-input"], {
                type: "checkbox", name: element.name, id: element_id,
            });

            ui.element("label", div, ["custom-control-label"], {for: element_id});

            return div
        },

        render: function(element) {
            return ""
        },
    },

    tagEditor: {
        input: function (element, form) {
            let el = ui.element("div", null, ["form-control", "form-control-sm", "tag-editor"], {tabindex: 1});

            let input = ui.element("input", el, [], {name: `[]${element.name}`});

            el.onclick = function () {
                input.focus()
            };
            el.onfocus = function () {
                input.focus()
            };

            input.style.width = "20px";

            input.onkeydown = function () {
                input.style.width = `${30 + (input.value.length * 7)}px`
            };

            let add = function (value) {
                let tag = ui.element("span", null, ["badge", "badge-primary"]);

                ui.text(value, tag);
                ui.element("input", tag, [], {type: "hidden", value: value, name: `[]${element.name}`});
                let sp = ui.text(" ", tag);

                let b = ui.element("span", tag, []);
                b.innerHTML = "&times;";
                b.onclick = function (e, el) {
                    tag.remove();
                    sp.remove()

                    return false
                };

                el.insertBefore(tag, input);
                el.insertBefore(ui.text(" "), input);
            };

            for (const v of element.value) {
                add(v)
            }

            input.onkeyup = function (e) {
                if (e.key === "Enter" && input.value) {
                    add(input.value);
                    input.value = "";
                    input.style.width = "20px";
                }
            };

            return el
        },

        render: function(element) {
            let span = ui.element("span", null, []);

            for (const v of element.value) {
                ui.element("span", span, ["badge", "badge-primary"]).append(ui.text(v));
                ui.text(" ", span);
            }

            return span
        },
    },
};
