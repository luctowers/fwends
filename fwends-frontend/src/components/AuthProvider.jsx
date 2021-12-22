export { AuthContext, AuthProvider };

import React, { useState, createContext } from "react";
import PropTypes from "prop-types";
import { authenticate } from "../api/auth";

let AuthContext = createContext({
  login: () => Promise.resolve(),
  logout: () => Promise.resolve(),
  authenticated: false,
  pending: false,
  error: null,
});

AuthProvider.propTypes = {
  children: PropTypes.any
};

function AuthProvider(props) {
  let [authenticated, setAuthenticated] = useState(false);
  let [pending, setPending] = useState(false);
  let [error, setError] = useState(false);

  async function login() {
    if (!authenticated) {
      setPending(true);
      setError(null);
      try {
        await authenticate(); 
        setPending(false);
        setAuthenticated(true);
      } catch (error) {
        setPending(false);
        setError(error);
      }
    }
  }

  async function logout() {
    setAuthenticated(false);
  }

  return (
    <AuthContext.Provider value={{login,logout,authenticated,pending,error}}>
      {props.children}
    </AuthContext.Provider>
  );
}
