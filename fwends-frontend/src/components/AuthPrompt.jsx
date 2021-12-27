import React, { useState } from "react";
import { useAuthPrompt } from "../api/auth";
import { GoogleSigninButton } from "./GoogleSigninButton";
import {XIcon} from "@primer/octicons-react";

export default function AuthPrompt() {
  const [show, setShow] = useState(false);
  const [loaded, setLoaded] = useState(false);

  useAuthPrompt(() => {
    if (loaded) {
      setShow(true);
    } else {
      setLoaded(true);
      setTimeout(() => {
        setShow(true);
      }, 250);
    }
  }, [loaded]);

  function hide() {
    setShow(false);
  }

  return (
    <div className={
      "p-6 flex items-center justify-center gap-4 fixed bottom-0 left-0 right-0 transition-transform ease-in-out duration-500 " +
      (show ? "translate-y-0" : "translate-y-3/4")
    }>
      {loaded && <GoogleSigninButton />}
      <button onClick={hide} className="h-10 w-10 button-frost">
        <XIcon className="w-5 h-5" />
      </button>
    </div>
  );
}
