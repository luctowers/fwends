export default App;

import React from "react";
import { Link } from "react-router-dom";

function App() {
    return (
        <div className='m-6 flex items-center justify-center sm:justify-start space-x-4 sm:space-x-6 md:space-x-8'>
            <Link className='font-bold' to='/'>
                FWENDS
            </Link>
            <Link to='/pieces'>
                PIECES
            </Link>
            <Link to='/packs'>
                PACKS
            </Link>
        </div>
    );
}
