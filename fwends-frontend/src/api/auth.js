import { useEffect, useState } from "react";
import { lazyPromise, dynamicScriptLoad } from "./util";

const eventTarget = new EventTarget();

const authConfig = lazyPromise(() =>
  fetch("/api/auth/config")
    .then(response => {
      if (response.ok) {
        let body = response.json();
        body.then();
        return body;
      } else {
        throw new Error("Failed to load auth config");
      }
    })
);

function googleAuthClient() {
  let clientId = authConfig()
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
        callback: console.log,
        cancel_on_tap_outside: true
      })
    );
}

export function authPrompt() {
  return googleAuthClient()
    .then(() => {
      window.google.accounts.id.prompt(notification => {
        if (notification.isNotDisplayed()) {
          eventTarget.dispatchEvent(new Event("prompt"));
        }
      });
    })
    .catch(console.error);
}

export function useAuthConfig() {
  let [config, setConfig] = useState();
  let [error, setError] = useState();
  useEffect(() => {
    authConfig()
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
