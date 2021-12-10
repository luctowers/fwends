export default App;

import React from 'react';
import { Routes, Route } from "react-router-dom";
import Header from './Header';
import Home from './Home';
import Pieces from './Pieces';
import Packs from './Packs';

function App() {
  return (
    <div className='m-4 sm:m-8'>
      <Header />
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/pieces" element={<Pieces />} />
        <Route path="/packs" element={<Packs />} />
      </Routes>
    </div>
  );
}
