import React, { forwardRef, useEffect, useRef } from "react";

export default forwardRef(function SpanInput(props, ref) {
	const { className, contentEditable, onChange, onEnter, placeholder } = props;
	const internalRef = useRef();
	useEffect(() => {
		ref.current = internalRef.current;
	});
	function handleChange() {
		if (ref.current.innerText.trim() === "") {
			ref.current.innerText = "";
		}
		// TODO: prevent extraneous onChange calls
		if (onChange) {
			onChange(ref.current.innerText);
		}
	}
	function handleKeyDown(event) {
		if(event.key === "Enter"){
			event.preventDefault();
			if (onEnter) {
				onEnter();
			}
		}
	}
	return (
		<span
			ref={internalRef}
			onInput={handleChange}
			onBlur={handleChange}
			onKeyDown={handleKeyDown}
			className={"span-input " + (className || "")}
			contentEditable={!(contentEditable === false)}
			placeholder={placeholder} />
	);
});
