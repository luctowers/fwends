export default App;

import React from "react";
import { Link } from "react-router-dom";
import { useAuthStatus, authClear } from "../api/auth";
import { LockIcon } from "@primer/octicons-react";

function App() {
  const authenticated = useAuthStatus();
  return (
    <div className='p-6 flex items-center justify-center sm:justify-start space-x-4 sm:space-x-6 md:space-x-8'>
      <Link className='font-bold' to='/'>
        FWENDS
      </Link>
      <Link to='/pieces'>
        PIECES
      </Link>
      <Link to='/packs'>
        PACKS
      </Link>
      {authenticated && 
        <button onClick={authClear}>
          <LockIcon />
        </button>
      }
    </div>
  );
}
