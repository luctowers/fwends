export default App;

import React from "react";
import { Link, useLocation } from "react-router-dom";
import { useAuthStatus, authClear, useAuthConfig } from "../api/auth";
import { LockIcon } from "@primer/octicons-react";

function App() {
	const authenticated = useAuthStatus();
	const [authConfig] = useAuthConfig();
	const location = useLocation();
	const navItems = [
		["/", "FWENDS"],
		["/pieces", "PIECES"],
		["/packs", "PACKS"],
	];
	return (
		<div className='p-6 flex items-center justify-center sm:justify-start space-x-4 sm:space-x-6 md:space-x-8 underline-offset-1'>
			{
				navItems.map(([path,label]) =>
					<Link
						key={path} 
						to={path}
						className={
							// apply bold to home/root
							(
								path == "/" ? "font-bold" : ""
							) +
							// underline current location
							(
								(
									path == "/" ? 
										location.pathname == "/" :
										location.pathname.startsWith(path)
								) ? " underline" : ""
							)
						}>
						{label}
					</Link>
				)
			}
			{authenticated && authConfig && authConfig.enable &&
				<button onClick={authClear}>
					<LockIcon />
				</button>
			}
		</div>
	);
}
