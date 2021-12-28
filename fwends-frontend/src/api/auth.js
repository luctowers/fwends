import { useEffect, useState } from "react";
import { lazyPromise, dynamicScriptLoad } from "./util";

let authenticatedGlobal = false;
const eventTarget = new EventTarget();

// TODO: prevent running this call for users that definetly aren't authenticated
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

const authConfig = lazyPromise(() =>
  fetch("/api/auth/config")
    .then(response => {
      if (response.ok) {
        return response.json();
      } else {
        throw new Error("Failed to load auth config");
      }
    })
);

const googleAuthClient = lazyPromise(() => {
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
        authenticatedGlobal = true;
        eventTarget.dispatchEvent(new Event("update"));
      } else {
        throw new Error("Failed to authenticate with server");
      }
    })
    .catch(console.error);
}

export function authClear() {
  authenticatedGlobal = false;
  eventTarget.dispatchEvent(new Event("update"));
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

export function useAuth() {
  let [authenticated, setAuthenticated] = useState(false);
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
