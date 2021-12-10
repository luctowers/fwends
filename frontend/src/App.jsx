export default App;

import React from 'react';
import { Routes, Route } from "react-router-dom";
import Header from './Header';
import Home from './Home';
import Pieces from './Pieces';
import Packs from './Packs';

function App() {
  return (
    <div>
      <Header />
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/pieces" element={<Pieces />} />
        <Route path="/packs" element={<Packs />} />
      </Routes>
    </div>
  );
}
