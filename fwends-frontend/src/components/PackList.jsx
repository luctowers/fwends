export default PackList;

import { TransitionGroup, CSSTransition } from "react-transition-group";
import { AlertIcon, LockIcon, PlusIcon } from "@primer/octicons-react";
import React, { useEffect, useState } from "react";
import { authPrompt, useAuthStatus } from "../api/auth";
import { useNavigate, Link } from "react-router-dom";
import { jsonRequest } from "../api/requests";
import { backend } from "../api/endpoints";

function PackList() {
	const authenticated = useAuthStatus();
	const navigate = useNavigate();
	const [packs, setPacks] = useState([]);
	const [error, setError] = useState(null);
	const [lockButton, setLockButton] = useState(false);
	const [lockTimeout, setLockTimeout] = useState(undefined);

	function handleNewPack() {
		if (authenticated) {
			navigate("./new");
		} else {
			if (!lockButton) {
				setLockButton(true);	
			}
			clearTimeout(lockTimeout);
			setLockTimeout(setTimeout(() => setLockButton(false), 1000));
			authPrompt();
		}
	}

	function NumericDisplay(props) {
		return (
			<div>
				<span  className="text-xl font-bold">
					{props.value}
				</span>
				<br/>
				{props.label}
			</div>
		);
	}

	useEffect(() => {
		jsonRequest(backend + "/packs").then(setPacks).catch(setError);
	}, []);

	return (
		<div className="grid gap-4 sm:gap-6 md:gap-8 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
			<button onClick={handleNewPack}>
				<div key="add" className="h-16 sm:h-32 button-frost">
					<div className="overlay w-5 h-5">
						<TransitionGroup>
							<CSSTransition key={lockButton} classNames="fade" timeout={500}>
								{ lockButton ? <LockIcon className="w-5 h-5" /> : <PlusIcon className="w-5 h-5" /> }
							</CSSTransition>
						</TransitionGroup>
					</div>
				</div>
			</button>
			{
				packs.map(pack =>
					<Link to={"/packs/"+pack.id} key={pack.id}>
						<div key={pack.id} className="h-32 button-frost flex-col">
							<div className="font-bold">
								{pack.title}
							</div>
							<span className="border border-neutral rounded px-1">{pack.hash.slice(0,7)}</span>
							<div className="flex justify-center space-x-16">
								<NumericDisplay value={pack.roleCount} label="Roles" />
								<NumericDisplay value={pack.stringCount} label="Strings" />
							</div>
						</div>
					</Link>
				)
			}
			{
				error &&
				<button onClick={() => window.location.reload()}>
					<div key="error" className="h-32 button-frost">
						<AlertIcon className="w-5 h-5" />
					</div>
				</button>
			}
		</div>
	);
}
