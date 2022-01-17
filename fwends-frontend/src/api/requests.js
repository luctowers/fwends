export function jsonRequest() {
	return window.fetch.apply(null, arguments)
		.then(response => {
			if (!response.ok) {
				throw new Error("Response status " + response.status);
			} else {
				return response.json();
			}
		});
}
