export default PackList;

import { PlusIcon } from "@primer/octicons-react";
import React, {useState, useEffect} from "react";
import { Link } from "react-router-dom";
import { authPrompt } from "../api/auth";
import PropTypes from "prop-types";

function PackList() {
  const [packs, setPacks] = useState([]);

  useEffect(async () => {
    let response = await fetch("/api/packs");
    let data = await response.json();
    setPacks(data);
  }, []);

  function NumericDisplay({value, label}) {
    return <div>
      <span  className='text-xl font-bold'>
        {value}
      </span>
      <br/>
      {label}
    </div>;
  }

  NumericDisplay.propTypes = {
    value: PropTypes.string,
    label: PropTypes.string,
  };

  const packElements = packs.map(pack =>
    <Link to={"/packs/"+pack.id} key={pack.id}>
      <div className='h-32 button-frost'>
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
    <div className='grid gap-4 sm:gap-6 md:gap-8 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'>
      {packElements}
      <button onClick={authPrompt}>
        <div key='add' className='h-32 button-frost'>
          <PlusIcon className="w-5 h-5" />
        </div>
      </button>
    </div>
  );
}
