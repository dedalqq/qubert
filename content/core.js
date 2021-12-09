let core = {
    selectedModule: null,
    selectedSubModule: 0,
    selectedModuleArgs: [],

    main: null,
    toastContainer: null,

    pushState: function() {
        history.pushState({
            module: core.selectedModule,
            subMod: core.selectedSubModule,
            args: core.selectedModuleArgs,
        }, "", "");
    },

    popState: async function(data) {
        if (data) {
            this.selectedModule = data.module;
            this.selectedSubModule = data.subMod;
            this.selectedModuleArgs = data.args;

            await this.renderModule(true)
        }
    },

    selectModule: async function(module, subModule) {
        this.selectedModule = module;
        this.selectedSubModule = subModule;
        this.selectedModuleArgs = [];

        this.pushState();

        await this.renderModule(true)

        ws.setLocation(this.selectedModule, this.selectedSubModule, this.selectedModuleArgs)
    },

    actionFunc: function(action, cb) {
        return function (e, el) {
            (async function() {
                try {
                    let action_result = await client.post(`/api/plugins/${core.selectedModule}/action`, {
                        cmd: action.cmd,
                        args: action.args,
                        data: core.getFormData(e.target),
                    })

                    let parent = el.parentNode
                    while (parent.localName !== "body") {
                        if (parent.classList.contains("modal")) {
                            parent.remove()
                            break
                        }
                        parent = parent.parentNode
                    }

                    await core.actionResponseHandler(action_result)
                } finally {
                    if (cb !== undefined) {
                        cb(e, el)
                    }
                }
            })().then()

            return false
        }
    },

    getFormData: function(element) {
        let form = undefined

        while (true) {
            if (element === document.body) {
                return undefined
            }

            if (element.localName === "form") {
                form = element
                break
            }

            element = element.parentNode
        }

        let data = {};

        for (const el of form.elements) {
            if (!el.name) {
                continue
            }

            if (el.name.startsWith('[]')) {
                let name = el.name.substr(2);

                if (!data[name]) {
                    data[name] = []
                }

                if (el.value) {
                    data[name].push(el.value)
                }
            } else if (el.type === "checkbox") {
                data[el.name] = el.checked
            } else if (el.type === "number") {
                data[el.name] = parseInt(el.value)
            } else {
                data[el.name] = el.value
            }
        }

        return data
    },

    update: async function(id, elementType, data) {
        switch (elementType) {
            case "progress":
                core.updateProgress(id, data)
        }
    },

    updateProgress: function (id, data) {
        let el = document.querySelector(`div#${id}.progress`)
        if (el) {
            el.children[0].style.width = `${data.value}%`
        }
    },

    renderModule: async function(fade) {
        if (core.main === null) {
            return null
        }

        if (core.selectedModule === null) {
            return null
        }

        let page_data = await client.post(`/api/plugins/${core.selectedModule}`, {
            sub_mod: core.selectedSubModule,
            args: core.selectedModuleArgs,
        })

        if (fade) {
            await core.transition(core.main, {opacity: 0}, 300)
        }

        ui.clear(core.main)

        ui.build({tag: "div", el: [
            {tag: "h2", text: page_data.title},
            {tag:"br"},
            uiTool.createElement(page_data.elements)
        ], parent: core.main})

        if (fade) {
            await core.transition(core.main, {opacity: 1}, 300)
        }
    },

    initLoginPage: function() {
        core.main = null

        ui.clear(document.body);

        let form
        let login_input
        let password_input

        ui.build({tag: "div", classes: ["login", "card", "shadow"], el: [
            {tag: "img", src: "apple-touch-icon.png"},
            {tag: "h1", text: "Qubert"},
            {tag: "form", cb: function (e) {form = e}, el: [
                {tag: "input", classes: ["form-control", "mb-3"], type: "text", name: "username", placeholder: "login", cb: function (e) {login_input = e}},
                {tag: "input", classes: ["form-control", "mb-3"], type: "password", name: "password", placeholder: "password", cb: function (e) {password_input = e}},
                {tag: "button", classes: ["btn", "btn-primary", "btn-block"], type: "submit", text: "Login"},
            ]},
            {tag: "span", text: "by dedal.qq (c) 2021"},
        ], parent: document.body})

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
            (async function() {
                if (login_input.focused) {
                    return
                }

                try {
                    let data = await client.post("/api/login", {
                        login: login_input.value,
                        password: password_input.value
                    })

                    if (data["access-token"]) {
                        window.localStorage.setItem('token', data["access-token"]);

                        await core.initMainPage(data["access-token"]);
                    }
                } catch (e) {
                    login_input.classList.add("is-invalid");
                    password_input.classList.add("is-invalid");
                }
            })().then()

            return false
        };
    },

    initMainPage: async function (token) {
        client.setToken(token)

        ws.connect(token)

        let pluginsLinks = []
        let hostName = ""
        let hostNameColor = ""

        try {
            // let userData = await client.get("/api/user")
            let mainPageData = await client.get("/api/main")

            let links = []

            for (let plugin of mainPageData.plugins) {
                let subPages = []

                for (let i in plugin["sub-pages"]) {
                    subPages.push({tag: "li", el: [
                        {tag: "a", classes: ["nav-link", "link-dark"], text: plugin["sub-pages"][i].title, href: "#", onclick: async function(e) {
                                for (let l of links) {
                                    l.classList.remove("active")
                                }

                                e.target.classList.add("active")

                                await core.selectModule(plugin.id, parseInt(i) + 1)
                        }, cb: function(e) {
                            links.push(e)
                        }},
                    ]})
                }

                pluginsLinks.push(
                    {tag: "div", el: [
                        {tag: "a", classes: ["nav-link", "link-dark"], el: [
                            {tag: "i", classes: ["bi", `bi-${plugin.icon}`, "me-2"]},
                            {text: plugin.title},
                        ], href: "#", onclick: async function(e) {
                            for (let l of links) {
                                l.classList.remove("active")
                            }

                            e.target.classList.add("active")

                            await core.selectModule(plugin.id, 0)
                        }, cb: function(e) {
                            links.push(e)
                        }},
                        {tag: "ul", classes: ["sub-module-list", "list-unstyled", "small"], el: subPages},
                    ]}
                )
            }

            hostName = mainPageData["host-name"]
            hostNameColor = mainPageData["host-badge-color"]
        } catch (e) {
            window.localStorage.removeItem('token');
            core.initLoginPage()

            return
        }

        ui.clear(document.body);

        ui.build({tag: "div", classes: ["container-xl"], el: [
            {tag: "div", classes: ["bg-dark", "shadow-sm", "rounded", "p-3", "my-3", "d-flex"], el: [
                {tag: "h1", classes: ["text-white"], el: [
                    {tag: "img", src: "apple-touch-icon.png", classes: ["logo", "me-3"]},
                    {text: "Qubert"},
                ]},
            ]},
            {tag: "div", classes: ["row"], el: [
                {tag: "div", classes: ["col-sm-2"], el: [
                    {tag: "div", classes: ["host-info", "card", "shadow-sm", "p-2"], el: [
                        {tag: "span", text: "Host name"},
                        {tag: "div", text: hostName},
                    ], cb: function (e) { e.style.backgroundColor = hostNameColor }},
                    {tag: "hr"},
                    {tag: "ul", classes: ["nav", "nav-pills", "flex-column"], el: [
                        {tag: "li", classes: ["nav-item"], el: pluginsLinks}
                    ]},
                    {tag: "hr"},
                    {tag: "span", classes: ["version"], text: "By dedal.qq (c) 2021"},
                ]},
                {tag: "div", classes: ["col-sm-10"], el: [
                    {tag: "div", classes: ["main", "rounded", "bg-white", "shadow-sm", "p-3"], cb: function(e) { core.main = e }, el: [

                    ]},
                ]},
            ]},
            {tag: "div", classes: ["toast-container"], cb: function(e) { core.toastContainer = e }},
        ], parent: document.body})
    },

    actionResponseHandler: async function(data) {
        switch (data.type) {
            case "set-args":
                this.selectedModuleArgs = data.options.args;

                core.pushState();
                await core.renderModule(true);

                return;
            case "alert":
                core.postToast(data.options.title, data.options.text)

                return;
            case "modal":
                let modal

                let actions = []

                for (let a of data.options.actions) {
                    actions.push(uiTool.createElement(a))
                }

                ui.build({tag: "div", classes: ["modal", "show"], el: [
                    {tag: "div", classes: ["modal-dialog"], el: [
                        {tag: "div", classes: ["modal-content", "shadow"], el: [
                            {tag: "div", classes: ["modal-header"], el: [
                                {tag: "h5", classes: ["modal-title"], el: uiTool.createLabelElements(data.options.title)},
                                {tag: "button", type: "button", classes: ["btn-close"], onclick: function (e, el) {
                                    modal.remove()
                                }},
                            ]},
                            {tag: "form", el: [
                                {tag: "div", classes: ["modal-body"], el: [uiTool.createElement(data.options.content)]},
                                {tag: "div", classes: ["modal-footer"], el: actions}
                            ]},
                        ]},
                    ]},
                ], parent: document.body, cb: function (e) { modal = e }})

                return
            default:
                await core.renderModule(false);
        }
    },

    transition: async function(el, options, duration) {
        let opts = []

        for (const [key, value] of Object.entries(options)) {
            opts.push(`${key} ${duration}ms`)
            el.style[key] = value
        }

        el.style.transition = opts.join(",")

        await new Promise(function (resolve) {
            el.ontransitionend = function () {
                el.ontransitionend = null
                el.style.transition = null

                resolve()
            }
        });
    },

    postToast: function(title, text) {
        let toast

        ui.build({tag: "div", classes: ["toast", "show"], el: [
            {tag: "div", classes: ["toast-header"], el: [
                {tag: "div", classes: ["me-auto"], el: uiTool.createLabelElements({
                    icon: "info-circle",
                    text: title,
                })},
                {tag:"button", type: "button", classes: ["btn-close"], onclick: function () {
                    toast.remove()

                    return false
                }}
            ]},
            {tag: "div", classes: ["toast-body"], text: text}
        ], cb: function(e) { toast = e }, parent: core.toastContainer})

        setTimeout(function () {
            if (toast) {
                toast.remove()
            }
        }, 5000)
    },
}
