let core = {
    selectedModuleNum: -1,
    selectedModuleArgs: [],

    reloadTimeOut: null,

    modal: null,

    pushState: function() {
        history.pushState({
            moduleNum: this.selectedModuleNum,
            args: this.selectedModuleArgs,
        }, "", "");
    },

    popState: function(data) {
        this.selectedModuleNum = data.moduleNum;
        this.selectedModuleArgs = data.args;

        this.loadModule()
    },

    loadModules: function() {
        client.get("/modules_list", function(data) {
            let modules = data.modules;

            for (const i in modules) {
                let link = ui.link(ui.iconText(modules[i].icon, modules[i].name), function() {
                    core.selectModule(i)
                });

                ui.appendLine(link, ui.left);
            }
        })
    },

    selectModule: function(module_num) {
        this.selectedModuleNum = module_num;
        this.selectedModuleArgs = [];

        if (core.reloadTimeOut) {
            clearTimeout(core.reloadTimeOut);
            core.reloadTimeOut = null;
        }

        this.pushState();
        this.loadModule()
    },

    loadModule: function() {
        if (this.selectedModuleNum < 0) {
            return
        }

        let params = [`module_num=${this.selectedModuleNum}`];
        for (const modArg of this.selectedModuleArgs) {
            params.push(`args=${modArg}`)
        }

        client.get(`/module?${params.join("&")}`, function(data) {
            ui.clear(ui.main);

            ui.append(ui.pageTitle(data.title), ui.main);
            core.appendElements(data, ui.main);

            if (data.reload > 0) {
                core.reloadTimeOut = setTimeout(function() { core.loadModule() }, data.reload * 1000)
            }
        })
    },

    action: function(button, action, args, data) {
        let bt_text = null;

        if (button) {
            bt_text = button.innerHTML;
            ui.clear(button);
            button.append(ui.spinner("loading ..."));
            button.blur()
        }

        client.post(`/action?module_num=${core.selectedModuleNum}&action=${action}`, function(data) {
            core.closeModal();

            if (button) {
                ui.clear(button);
                button.innerHTML = bt_text;
                button.focus()
            }

            core.actionResponseHandler(data)
        }, {
            args: args,
            data: data,
        })
    },

    createElement: function(element_name, options) {  // TODO one parameter
        switch (element_name) {
            case "button":
                return uitool.button(options, function() {
                    core.action(this, options.action.action, options.action.args, null)
                });
            case "table":
                return uitool.table(options);
            case "form":
            // return uitool.form(options)
            case "text":
                return uitool.text(options);
            case "label":
                return uitool.label(options);
            case "card":
                return uitool.card(options);
            case "progress":
                return uitool.progress(options);
            case "columns":
                return uitool.columns(options);
            case "row":
                return uitool.row(options);
            case "text-viewer":
                return uitool.textViewer(options);
            case "header":
                return uitool.header(options);
            case "attribute-list":
                return uitool.attributeList(options);
            case "badge-list":
                return uitool.badgeList(options);
            case "value-editor":
                return uitool.valueEditor(options);
            case "switcher":
                return uitool.switcher(options);
        }
    },

    appendElements: function(data, content) { // TODO move to uitool
        if (!data.elements) {
            return
        }

        for (const el of data.elements) {
            let element = this.createElement(el.name, el.options);
            ui.appendLine(element, content)
        }
    },

    actionResponseHandler: function(data) {
        switch (data.type) {
            case "set-args":
                this.selectedModuleArgs = data.options.args;

                this.pushState();
                this.loadModule();
                return;

            case "modal":
                let modal_body = ui.element("div", null, [], {});
                let form = null;

                switch (data.options.modal_type) {
                    case "content":
                        this.appendElements(data.options.content, modal_body);
                        break;

                    case "form":
                        form = uitool.formElements(data.options.content.form_items, modal_body);
                        break;

                    default:
                        return
                }

                let action_buttons = [];

                for (const a of data.options.actions) {
                    action_buttons.push(uitool.button(a, function() {
                        let data = null;
                        if (form) {
                            data = uitool.getFormData(form);
                        }

                        core.action(this, a.action.action, a.action.args, data);
                    }))
                }

                this.modal = this.createModal(data.options, modal_body, action_buttons);

                this.showModal(this.modal);
                break;

            case "empty":
                this.loadModule();
        }
    },

    createModal: function(options, body, actions) {
        let modal = ui.element("div", null, ["modal", "shadow"], {role: "dialog"});

        let dialog = ui.element("div", modal, ["modal-dialog", "shadow-lg"], {
            role: "document"
        });

        let content = ui.element("div", dialog, ["modal-content"]);

        let header = ui.element("div", content, ["modal-header"]);
        let h5 = ui.element("h5", header, ["modal-title"]);

        if (options.icon) {
            let icon = ui.icon(options.icon, h5);
            ui.text(" ", h5);

            if (options.icon_color) {
                icon.style.color = options.icon_color
            }
        }

        ui.text(options.title, h5);

        let close_button = ui.element("button", header, ["close"], {type: "button"});
        close_button.onclick = function() {
            core.closeModal();
        };

        ui.element("span", close_button, []).innerHTML = "&times;";

        ui.element("div", content, ["modal-body"]).append(body);

        let footer = ui.element("div", content, ["modal-footer"]);

        for (const a of actions) {
            footer.append(a)
        }

        return modal
    },

    showModal: function(modal) {
        modal.classList.add("show");
        document.body.append(modal)
    },

    closeModal: function() {
        if (this.modal) {
            this.modal.remove();
            this.modal = null
        }
    },

    initLoginPage: function() {
        ui.clear(document.body);

        let login = ui.element("div", document.body, ["login", "shadow-sm"]);

        ui.element("h1", login, []).append(ui.text("Test 123"));
        let form = ui.element("form", login, []);

        let login_input = ui.element("input", form, ["form-control"], {type: "text", name: "username", placeholder: "login"});
        let password_input = ui.element("input", form, ["form-control"], {type: "password", name: "password", placeholder: "password"});

        login_input.focus();

        login_input.onkeypress = function(e) {
            if (e.key === "Enter") {
                if (login_input.value !== "") {
                    password_input.focus()
                }

                return false
            }
        };

        form.onsubmit = function() {
            if (login_input.focused) {
                return false
            }

            client.post("/user", function(data) {
                if (data.user) {
                    core.initMainPage();
                } else {
                    login_input.classList.add("is-invalid");
                    password_input.classList.add("is-invalid");
                }
            }, {login: login_input.value, password: password_input.value});

            return false
        };

        let login_button = ui.element("button", form, ["btn", "btn-primary", "btn-block"], {type: "submit"});
        login_button.append(ui.text("login"));
    },

    initMainPage: function() {
        ui.clear(document.body);

        let container = ui.element("div", document.body, ["container"]);

        let navbar = ui.element("nav", container, ["navbar", "navbar-dark", "bg-dark", "shadow-sm"]);
        let navTitle = ui.element("div", navbar);
        ui.element("span", navTitle, ["navbar-brand", "mb-0", "h1"]).append(ui.text("Test 123"));
        socket.status = ui.element("div", navbar, ["navbar-brand"], {id: "status"});

        let div = ui.element("div", container, ["row"]);

        let left_div = ui.element("div", div, ["col-2"]);
        ui.left = ui.element("div", left_div, [], {id: "left"});

        let main_div = ui.element("div", div, ["col-10"]);
        ui.main = ui.element("div", main_div, [], {id: "main"});

        events.events = ui.element("div", document.body, [], {id: "events"});

        this.loadModules();
        socket.run()
    },

    run: function() {
        window.onpopstate = function(e) {
            core.popState(e.state);
        };

        document.addEventListener("keyup", function(e) {
            if (e.key === "Escape") {
                core.closeModal()
            }
        });

        client.get("/user", function(data) {
            if (!data.user) {
                core.initLoginPage();
                return
            }

            core.initMainPage();
        });
    },
};