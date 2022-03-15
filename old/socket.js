let socket = {
    status: null,
    isRunning: false,

    onmessage: function(message) {
        let data = JSON.parse(message.data);

        if (data.type === "close") {
            core.initLoginPage();
            socket.isRunning = false;
            return
        }

        if (data.type === "alert") {
            events.showEvent(data.options.title, data.options.text);
        }

        core.loadModule()
    },

    onopen: function() {
        ui.clear(this.status);

        this.status.append(ui.text("OnLine"))
    },

    onclose: function() {
        if (socket.isRunning) {
            ui.clear(this.status);
            this.status.append(ui.text("OffLine"));
            setTimeout(this.connect, 1000)
        }
    },

    connect: function() {
        let url_data = window.location.href.split("/");

        let s = new WebSocket(`ws://${url_data[2]}/socket`);

        s.onopen = function() { socket.onopen() };
        s.onclose = function() { socket.onclose() };
        s.onmessage = function(message) { socket.onmessage(message) };
        s.onerror = function(e) { console.log(e) }
    },

    run: function() {
        this.isRunning = true;
        this.connect()
    }
};