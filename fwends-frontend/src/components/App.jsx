export default App;

import React from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import Header from "./Header";
import Home from "./Home";
import Pieces from "./Pieces";
import Packs from "./Packs";
import NotFound from "./NotFound";
import AuthPrompt from "./AuthPrompt";

function App() {
  return (
    <div className='m-4 sm:m-8'>
      <Header />
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/pieces" element={<Pieces />} />
        <Route path="/packs" element={<Packs />} />
        <Route path='/404' element={<NotFound />} />
        <Route path="*" element={<Navigate replace to="/404" />} />
      </Routes>
      <AuthPrompt />
    </div>
  );
}
