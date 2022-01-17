import React, { useState } from "react";
import { useAuthStatus, useAuthPrompt } from "../api/auth";
import { GoogleSigninButton } from "./GoogleSigninButton";
import {XIcon} from "@primer/octicons-react";

export default function AuthPrompt() {
	const [show, setShow] = useState(false);
	const [loaded, setLoaded] = useState(false);
	const authenticated = useAuthStatus();

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
			"p-4 flex items-center justify-center gap-4 fixed bottom-0 left-0 right-0 transition-transform ease-in-out duration-500 border-t border-neutral bg-white " +
      (show && !authenticated ? "translate-y-0" : "translate-y-full")
		}>
			{loaded && <GoogleSigninButton />}
			<button onClick={hide} className="h-10 w-10 button-frost">
				<XIcon className="w-5 h-5" />
			</button>
		</div>
	);
}
