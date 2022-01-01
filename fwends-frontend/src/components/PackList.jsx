export default PackList;

import { PlusIcon } from "@primer/octicons-react";
import React from "react";
import { authPrompt, useAuthStatus } from "../api/auth";
import { useNavigate } from "react-router-dom";

function PackList() {
	const authenticated = useAuthStatus();
	const navigate = useNavigate();

	function handleNewPack() {
		if (authenticated) {
			navigate("./new");
		} else {
			authPrompt();
		}
	}

	return (
		<div className='grid gap-4 sm:gap-6 md:gap-8 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'>
			<button onClick={handleNewPack}>
				<div key='add' className='h-32 button-frost'>
					<PlusIcon className="w-5 h-5" />
				</div>
			</button>
		</div>
	);
}
