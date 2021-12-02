async function main(e) {
    window.onpopstate = function(e) {
        core.popState(e.state).then();
    };

    let token = window.localStorage.getItem('token');

    if (token !== null) {
        await core.initMainPage(token)

        return
    }

    core.initLoginPage()
}