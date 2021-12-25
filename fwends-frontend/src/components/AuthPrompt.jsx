import React, { useState } from "react";
import { useAuthPrompt } from "../api/auth";
import { GoogleSigninButton } from "./GoogleSigninButton";
import {XIcon} from "@primer/octicons-react";

export default function AuthPrompt() {
  const [show, setShow] = useState(false);

  useAuthPrompt(() => {
    setShow(true);
  }, []);

  function hide() {
    setShow(false);
  }

  return (
    <div className={"m-6 flex items-center justify-center gap-4 fixed bottom-0 left-0 right-0 " + (show ? "" : "hidden")}>
      <GoogleSigninButton />
      <button onClick={hide} className="h-10 w-10 button-frost">
        <XIcon className="w-5 h-5" />
      </button>
    </div>
  );
}
