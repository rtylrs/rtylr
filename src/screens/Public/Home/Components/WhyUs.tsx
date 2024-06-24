import {
  faUsers,
  faComments,
  faLanguage,
  faEdit,
} from "@fortawesome/free-solid-svg-icons";
import React from "react";
import PropTypes from "prop-types";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

const whyUsFeatures = [
  {
    icon: faUsers,
    title: "Productized Service",
    description:
      "Ditch developer management hassle. We deliver a tailored solution, covering everything from ideation to maintenance.",
  },
  {
    icon: faComments,
    title: "Transparent Communication",
    description:
      "Stay informed every step of the way with clear communication and progress updates.",
  },
  {
    icon: faLanguage,
    title: "24/7 Support & Unlimited Requests",
    description:
      "Get the help you need, whenever you need it, without worrying about limits.",
  },
  {
    icon: faEdit,
    title: "Scalable Solutions",
    description:
      "We adapt to your growing needs, ensuring a seamless team expansion.",
  },
];

export const WhyUs: React.FC = () => {
  return (
    <div className="why-us">
      <h2>Why Choose Us</h2>
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

const WhyUsCard: React.FC<{
  icon: any;
  title: string;
  description: string;
}> = ({ icon, title, description }) => {
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
