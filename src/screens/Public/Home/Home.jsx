import React from "react";

import Hero from "./Components/Hero";
import Features from "./Components/Features";
import Ticker from "../../../components/Ticker";
import WhyUs from "./Components/WhyUs";
import Goals from "./Components/Goals";
import Price from "./Components/Price";
function Home() {
  return (
    <React.Fragment>
      <Hero />
      <Ticker />
      <Features />
      <WhyUs />
<Price/>
      <Goals />
    </React.Fragment>
  );
}

export default Home;
