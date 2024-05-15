/*
  File Created By: Mohamad Homsi
  Last Edited: Thursday 16 May 2024
  Last Edited by: Mohamad Homsi
*/

import { useEffect } from "react";
import { useParams } from "react-router-dom";

function Verified() {
  const { token } = useParams();

  useEffect(() => {
    console.log("Token:", token);

    // TODO: Verify Email
  }, []);

  return <div>Your Account Has Been Verified</div>;
}

export default Verified;
