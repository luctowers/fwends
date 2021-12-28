export default App;

import React from "react";
import { Routes, Route, Navigate, useLocation } from "react-router-dom";
import { TransitionGroup, CSSTransition } from "react-transition-group";
import Header from "./Header";
import Home from "./Home";
import Pieces from "./Pieces";
import Packs from "./Packs";
import NotFound from "./NotFound";
import AuthPrompt from "./AuthPrompt";

function App() {
  let location = useLocation();
  return (
    <div className='px-4 sm:px-8 mx-auto max-w-screen-2xl'>
      <Header />
      <div className="overlay">
        <TransitionGroup>
          <CSSTransition key={location.pathname} classNames="fade" timeout={500}>
            <Routes location={location}>
              <Route path="/" element={<Home />} />
              <Route path="/pieces" element={<Pieces />} />
              <Route path="/packs" element={<Packs />} />
              <Route path='/404' element={<NotFound />} />
              <Route path="*" element={<Navigate replace to="/404" />} />
            </Routes>
          </CSSTransition>
        </TransitionGroup>
      </div>
      <AuthPrompt />
    </div>
  );
}
