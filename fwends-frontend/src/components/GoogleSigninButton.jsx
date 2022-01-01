import React, { useEffect, useRef } from "react";
import { useGoogleAuthClient } from "../api/auth";

export function GoogleSigninButton() {
	let buttonRef = useRef();
	let [initialized] = useGoogleAuthClient();
	useEffect(() => {
		if (initialized) {
			window.google.accounts.id.renderButton(
				buttonRef.current,
				{ theme: "outline", size: "large" }
			);
		}
	}, [initialized]);
	return (
		<div ref={buttonRef}></div>
	);
}
