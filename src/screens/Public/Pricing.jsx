const Pricing = () => {
  return (
    <div className="background">
      <div className="container">
        <div className="panel pricing-table">
          <div className="pricing-plan">
            <img
              src="https://s22.postimg.cc/8mv5gn7w1/paper-plane.png"
              alt=""
              className="pricing-img"
            />
            <h2 className="pricing-header">Personal</h2>
            <ul className="pricing-features">
              <li className="pricing-features-item">Purchase Products</li>
              <li className="pricing-features-item">Free Support</li>
            </ul>
            <span className="pricing-price">Free</span>
            <a href="#/" className="pricing-button">
              Sign up
            </a>
          </div>

          <div className="pricing-plan">
            <img
              src="https://s28.postimg.cc/ju5bnc3x9/plane.png"
              alt=""
              className="pricing-img"
            />
            <h2 className="pricing-header">Pro</h2>
            <ul className="pricing-features">
              <li className="pricing-features-item">Temp</li>
              <li className="pricing-features-item">Temp</li>
            </ul>
            <span className="pricing-price">$100</span>
            <a href="#/" className="pricing-button is-featured">
              Get started
            </a>
          </div>

          <div className="pricing-plan">
            <img
              src="https://s21.postimg.cc/tpm0cge4n/space-ship.png"
              alt=""
              className="pricing-img"
            />
            <h2 className="pricing-header">Enterprise</h2>
            <ul className="pricing-features">
              <li className="pricing-features-item">temp</li>
              <li className="pricing-features-item">temp</li>
            </ul>
            <span className="pricing-price">$100</span>
            <a href="#/" className="pricing-button">
              Get started
            </a>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Pricing;
