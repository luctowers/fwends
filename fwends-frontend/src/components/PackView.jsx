import React, { useEffect, useRef } from "react";
import { Link, useParams } from "react-router-dom";
import SpanInput from "./SpanInput";

export default function PackView() {
  const {packId} = useParams();
  const titleRef = useRef();
  useEffect(() => {
    if (packId == "new" && titleRef.current) {
      titleRef.current.focus();
    }
  }, [packId]);
  if (packId == "new") {
    return (
      <p className="text-xl tracking-wide">
        <Link to='/packs' className="hover:underline decoration-1">packs</Link>
        {" / "}
        <SpanInput ref={titleRef} placeholder='Title' />
      </p>
    );
  }
}
