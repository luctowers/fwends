import React, { useEffect, useRef } from "react";
import { useParams } from "react-router-dom";

export default function PackView() {
  const {packId} = useParams();
  const titleRef = useRef();
  useEffect(() => {
    if (packId == "new" && titleRef.current) {
      titleRef.current.focus();
    }
  }, [packId]);
  if (packId == "new") {
    return <input ref={titleRef} placeholder='Pack Title' type='text' className='appearance-none focus:outline-none text-xl' />;
  }
}
