export { authenticate };

function AsyncSingleton(fn) {
    let singleton;
    return async function() {
        if (singleton) {
            return singleton;
        } else {
            singleton = fn.apply(null, arguments)
                .then(result => {
                    singleton = undefined;
                    return result;
                })
                .catch(error => {
                    singleton = undefined;
                    throw error;
                });
        }
        return singleton;
    };
}

let fetchAuthInfo = AsyncSingleton(async function() {
    let response = await fetch("/api/auth/");
    if (response.ok) {
        return response.json();
    } else {
        throw new Error("Failed to fetch auth info");
    }
});

let googleAuthenticate = AsyncSingleton(function(clientId) {
    return new Promise((resolve) => {
        window.google.accounts.id.initialize({
            client_id: clientId,
            callback: response => {
                resolve(response);
            }
        });
        window.google.accounts.id.prompt();
    });
});

let authenticate = AsyncSingleton(async function() {
    let info = await fetchAuthInfo();
    if (!info.enable) {
        return;
    } else {
        await googleAuthenticate(info.services.google);
        return;
    }
});
