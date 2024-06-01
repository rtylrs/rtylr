// src/components/WhyUsCard.jsx
import React from 'react';
import PropTypes from 'prop-types';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

const WhyUsCard = ({ icon, title, description }) => {
  return (
    <div className="why-us-card">
      <div className="icon-container">
        <FontAwesomeIcon icon={icon} size="2x" />
      </div>
      <div className="whyus-text-container">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
};

WhyUsCard.propTypes = {
  icon: PropTypes.object.isRequired,
  title: PropTypes.string.isRequired,
  description: PropTypes.string.isRequired,
};

export default WhyUsCard;
