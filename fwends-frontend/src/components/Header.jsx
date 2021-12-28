export default App;

import React, {useState} from "react";
import { Link } from "react-router-dom";
import { useAuthStatus, authClear } from "../api/auth";
import { LockIcon } from "@primer/octicons-react";

function App() {
  const authenticated = useAuthStatus();
  const [message, setMessage] = useState(null);

  function signOut() {
    authClear();
    setMessage("Signed out!");
    setTimeout(() => {
      setMessage(null);
    }, 1000);
  }

  return (
    <div className='m-6 flex items-center justify-center sm:justify-start space-x-4 sm:space-x-6 md:space-x-8'>
      {message ? (
        <p>{message}</p>
      ) : (
        <>
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
            <button onClick={signOut}>
              <LockIcon />
            </button>
          }
        </>
      )}
    </div>
  );
}
