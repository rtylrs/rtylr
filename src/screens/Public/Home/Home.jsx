import React from "react";

import Hero from "./Components/Hero";
import Features from "./Components/Features";
import Ticker from "../../../components/Ticker";
import WhyUs from "./Components/WhyUs";
import Price from "./Components/Price";
import Demo from "./Components/Demo";
import HelmetContainer from "../../../components/HelmetContainer";

function Home() {
  return (
    <React.Fragment>
      <HelmetContainer />
      <Hero />
      <Ticker />
      <Features />
      <WhyUs />
      <Price />
      <Demo />
    </React.Fragment>
  );
}

export default Home;
