export default PackList;

import React, {useState, useEffect, useContext} from "react";
import PropTypes from "prop-types";
import { Link } from "react-router-dom";
import { AuthContext } from "./AuthProvider";

function PackList() {
    const [packs, setPacks] = useState([]);
    const {login,authenticated,pending,error} = useContext(AuthContext);

    useEffect(async () => {
        let response = await fetch("/api/packs");
        let data = await response.json();
        setPacks(data);
    }, []);

    function NumericDisplay(props) {
        return <div>
            <span  className='text-xl font-bold'>
                {props.value}
            </span>
            <br/>
            {props.label}
        </div>;
    }

    NumericDisplay.propTypes = {
        value: PropTypes.number,
        label: PropTypes.string
    };

    const packElements = packs.map(pack =>
        <Link to={"/packs/"+pack.id} key={pack.id}>
            <div className='h-32 rounded-lg text-center border border-stone-200 hover:bg-stone-200 transition-colors'>
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
            <h1>{authenticated ? "LOGGED IN" : "NOT LOGGED IN"}</h1>
            <h1>{pending ? "PENDING" : "NOT PENDING"}</h1>
            <h1>{error ? error.message : "NO ERROR"}</h1>
            {packElements}
            <button onClick={login}>
                <div key='add' className='h-32 rounded-lg border border-stone-200 hover:bg-stone-200 text-center flex items-center justify-center'>
                    <p className='text-xl text-stone-400'>+</p>
                </div>
            </button>
        </div>
    );
}
