export default Packs;

import React from 'react';
import { Link } from "react-router-dom";

function Packs() {
  const [packs, setPacks] = React.useState([]);

  React.useEffect(async () => {
    let response = await fetch('/api/packs')
    let data = await response.json()
    setPacks(data);
  }, []);

  function NumericDisplay(props) {
    return <div>
      <span  className='text-xl font-bold'>
        {props.value}
      </span>
      <br/>
      {props.label}
    </div>
  }

  const packElements = packs.map(pack =>
    <Link to={'/packs/'+pack.id}>
      <div key={pack.name} className='h-32 rounded-lg text-center border border-stone-200 hover:bg-stone-200 transition-colors'>
        <div className='font-bold my-4'>
          {pack.name}
        </div>
        <div className='flex justify-center space-x-16'>
          <NumericDisplay value={pack.roleCount} label='Roles' />
          <NumericDisplay value={pack.stringCount} label='Strings' />
        </div>
      </div>
    </Link>
  );

  return (
    <div className='grid gap-4 sm:gap-6 md:gap-8 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 transition-colors'>
      {packElements}
      <Link to='/packs#new'>
        <div key='add' className='h-32 rounded-lg border border-stone-200 hover:bg-stone-200 text-center flex items-center justify-center'>
          <p className='text-xl text-stone-400'>+</p>
        </div>
      </Link>
    </div>
  );
}
