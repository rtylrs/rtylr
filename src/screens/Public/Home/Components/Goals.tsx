import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faRecycle,
  faLeaf,
  faTint,
  faShoppingCart,
  faHandHoldingHeart,
  faUsers,
} from "@fortawesome/free-solid-svg-icons";
import React from "react";

const Goals: React.FC = () => {
  return (
    <div>
      <section className="home-goals">
        <div className="home-goals-content">
          <div className="goal">
            <FontAwesomeIcon icon={faRecycle} />
            <p>Promote Waste Reduction</p>
          </div>
          <div className="goal">
            <FontAwesomeIcon icon={faLeaf} />
            <p>Advance Sustainable Living</p>
          </div>
          <div className="goal">
            <FontAwesomeIcon icon={faTint} />
            <p>Eliminate Single-Use Plastics</p>
          </div>
          <div className="goal">
            <FontAwesomeIcon icon={faShoppingCart} />
            <p>Introduce Waste-Free Shopping Culture</p>
          </div>
          <div className="goal">
            <FontAwesomeIcon icon={faHandHoldingHeart} />
            <p>Support the Community</p>
          </div>
          <div className="goal">
            <FontAwesomeIcon icon={faUsers} />
            <p>Inspire Sustainable Awareness</p>
          </div>
        </div>
      </section>
    </div>
  );
};

export default Goals;
