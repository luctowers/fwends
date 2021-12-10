export default App;

import React from 'react';
import { Link } from "react-router-dom";

function App() {
  return (
    <div className="py-6 lg:px-8 mx-4 lg:mx-0 flex items-center space-x-8">
      <Link to='/'>
        <h1 className="pl-2 font-bold">FWENDS</h1>
      </Link>
      <Link to='/pieces'>
        <h1>PIECES</h1>
      </Link>
      <Link to='/packs'>
        <h1>PACKS</h1>
      </Link>
    </div>
  );
}
