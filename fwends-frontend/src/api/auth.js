import { useEffect, useState } from "react";
import Cookies from "js-cookie";
import { lazyPromise, dynamicScriptLoad } from "./util";

let authenticatedGlobal = false;
const eventTarget = new EventTarget();
const sessionPresenceCookie = "fwends_session_presence";

if (Cookies.get(sessionPresenceCookie) === "true") {
	fetch("/api/auth")
		.then(response => {
			if (response.ok) {
				return response.json();
			} else {
				throw new Error("Failed to verify session");
			}
		})
		.then(status => {
			if (!authenticatedGlobal && status) {
				authenticatedGlobal = true;
				eventTarget.dispatchEvent(new Event("update"));
			}
		})
		.catch(console.error);
}

const authConfig = fetch("/api/auth/config")
	.then(response => {
		if (response.ok) {
			return response.json();
		} else {
			throw new Error("Failed to load auth config");
		}
	});

authConfig
	.then(config => {
		if (!config.enable) {
			authenticatedGlobal = true;
			eventTarget.dispatchEvent(new Event("update"));
		}
	});

const googleAuthClient = lazyPromise(() => {
	let clientId = authConfig
		.then(config => {
			if (!config.enable || !config.services.google) {
				throw new Error("Google authentication is not enabled");
			} else {
				return config.services.google;
			}
		});
	return Promise.all([
		clientId,
		dynamicScriptLoad("https://accounts.google.com/gsi/client")
	])
		.then(data =>
			window.google.accounts.id.initialize({
				client_id: data[0],
				callback: handleGoogleCredentialResponse,
			})
		);
});

function handleGoogleCredentialResponse(response) {
	fetch("/api/auth", {
		method: "POST",
		headers: {
			"Content-Type": "application/json"
		},
		body: JSON.stringify({
			token: response.credential,
			service: "google"
		})
	})
		.then(response => {
			if (response.ok) {
				Cookies.set(sessionPresenceCookie, "true");
				authenticatedGlobal = true;
				eventTarget.dispatchEvent(new Event("update"));
			} else {
				throw new Error("Failed to authenticate with server");
			}
		})
		.catch(console.error);
}

export function authClear() {
	Cookies.remove(sessionPresenceCookie);
	authenticatedGlobal = false;
	eventTarget.dispatchEvent(new Event("update"));
}

export function authPrompt() {
	if (!authenticatedGlobal) {
		googleAuthClient()
			.then(() => {
				window.google.accounts.id.prompt(notification => {
					if (notification.isNotDisplayed()) {
						eventTarget.dispatchEvent(new Event("prompt"));
					}
				});
			})
			.catch(console.error);
	}
}

export function useAuthStatus() {
	let [authenticated, setAuthenticated] = useState(authenticatedGlobal);
	useEffect(() => {
		function handleUpdate() {
			setAuthenticated(authenticatedGlobal);
		}
		eventTarget.addEventListener("update", handleUpdate);
		return () => {
			eventTarget.removeEventListener("update", handleUpdate);
		};
	}, []);
	return authenticated;
}

export function useAuthConfig() {
	let [config, setConfig] = useState();
	let [error, setError] = useState();
	useEffect(() => {
		authConfig
			.then(setConfig)
			.catch(setError);
	}, []);
	return [config, error];
}

export function useGoogleAuthClient() {
	let [initialized, setInitialized] = useState(false);
	let [error, setError] = useState();
	useEffect(() => {
		googleAuthClient()
			.then(() => setInitialized(true))
			.catch(setError);
	}, []);
	return [initialized, error];
}

export function useAuthPrompt(callback, deps) {
	useEffect(() => {
		eventTarget.addEventListener("prompt", callback);
		return () => {
			eventTarget.removeEventListener("prompt", callback);
		};
	}, deps);
}
