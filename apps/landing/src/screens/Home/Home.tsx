import React from 'react';
import Hero from './Components/Hero';
import Features from './Components/Features';
// import Ticker from '../../components/Ticker/Ticker';
import Price from './Components/Price';
import Demo from './Components/Demo';
import HelmetContainer from '../../components/HelmetContainer';
import Faq from './Components/FAQ';

const Home = (): JSX.Element => {
  return (
    <React.Fragment>
      <HelmetContainer />
      <Hero />
      {/* <Ticker /> */}
      <Features />
      <Demo />
      <Price />
      <Faq />
    </React.Fragment>
  );
};

export default Home;
