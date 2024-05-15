import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import useColors from "../../../../hooks/useColors";


function WhyUs (){
    const { t } = useTranslation();
    const navigate = useNavigate();
    const { accent } = useColors();

return(
<div className="whyus-container">
  <section className="whyUs-container-1">
    <div className="whyUs-text1">
      <h1>Enhance your promotions <br />and elevate your brand</h1>
      <h2>Transforming your restaurant's vision into captivating digital solutions since 1998. <br /> We conceptualize, design, and develop data-driven products to overcome your business challenges.</h2>
      <br />
      <h2>Experience our comprehensive services designed to guide you seamlessly in crafting digital masterpieces, on budget and on time.</h2>
    </div>
    <div className="whyUs-image">
      <img src="/images/test-image.png" alt="" />
    </div>
  </section>

  <section className="whyUs-container-2">
    <div className="whyUs-image2">
      <img src="/images/test2.png" alt="" />
    </div>
    <div className="whyUs-text2">
      <h1>Stay ahead with innovative digital strategies</h1>
      <h2>Our team of experts crafts cutting-edge websites infused with modern technologies and forward-thinking design, ensuring your restaurant stands out in the digital landscape.</h2>
      <br />
      <h2>Benefit from our holistic services tailored to bring your digital vision to life, every step of the way.</h2>
    </div>
  </section>
</div>

);



}

export default WhyUs;

/*
  File Created By: Abed El-Fattah Amouneh
  Last Edited: Wednesday 15 May 2024
  Last Edited by: Abed El-Fattah Amouneh
*/