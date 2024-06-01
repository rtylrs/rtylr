// src/components/WhyUs.jsx
import React from 'react';
import WhyUsCard from './WhyUsCard';
import { faUsers, faComments, faSearch, faLanguage, faEdit, faMoneyBillWave, faStar, faCalendarAlt } from '@fortawesome/free-solid-svg-icons';

const whyUsFeatures = [
  {
    icon: faUsers,
    title: 'Productized Service',
    description: 'Ditch developer management hassle. We deliver a tailored solution, covering everything from ideation to maintenance.',
  },
  {
    icon: faComments,
    title: 'Transparent Communication',
    description: 'Stay informed every step of the way with clear communication and progress updates.',
  },
  
  {
    icon: faLanguage,
    title: '24/7 Support & Unlimited Requests',
    description: 'Get the help you need, whenever you need it, without worrying about limits.',
  },
  {
    icon: faEdit,
    title: 'Scalable Solutions',
    description: 'We adapt to your growing needs, ensuring a seamless team expansion.',
  },
];

const WhyUs = () => {
  return (
    <div className="why-us">
      
      <h2>Why Choose Us </h2>
      <h1>What Sets Us Apart?</h1>

      <div className="why-us-cards">
        {whyUsFeatures.map((feature, index) => (
          <WhyUsCard
            key={index}
            icon={feature.icon}
            title={feature.title}
            description={feature.description}
          />
        ))}
      </div>
    </div>
  );
};

export default WhyUs;
