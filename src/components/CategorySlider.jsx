import { useState } from "react";
import { Categories } from "../data/categories";
import { Link } from "react-router-dom";

function CategorySlider() {
  const [currentIndex, setCurrentIndex] = useState(0);

  const nextSlide = () => {
    const nextIndex =
      currentIndex + 1 >= Categories.length ? 0 : currentIndex + 1;
    setCurrentIndex(nextIndex);
  };

  const prevSlide = () => {
    const prevIndex =
      currentIndex - 1 < 0 ? Categories.length - 1 : currentIndex - 1;
    setCurrentIndex(prevIndex);
  };

  return (
    <div
      className="vertical-slider"
      style={{
        height: "100%",
      }}
    >
      <button className="slider-btn up-btn" onClick={prevSlide}>
        &lt;
      </button>
      <div
        className="slider-container"
        style={{
          height: "100%",
        }}
      >
        <div
          className="slider-content"
          style={{
            width: `calc(100% / 5 * ${Categories.length})`,
            transform: `translateX(-${
              currentIndex * (100 / Categories.length)
            }%)`,
          }}
        >
          {Categories.map((category, index) => (
            <div
              key={index}
              className="slider-item"
              style={{
                backgroundImage: `url(${category.imageUrl})`,
                backgroundSize: "cover",
                backgroundPosition: "center",
              }}
            >
              <Link
                to={
                  category.title.toLowerCase() === "brands"
                    ? `/brands`
                    : `/collection/${category.title.toLowerCase()}`
                }
              >
                <p>{category.title}</p>
              </Link>
            </div>
          ))}
        </div>
      </div>
      <button className="slider-btn down-btn" onClick={nextSlide}>
        &gt;
      </button>
    </div>
  );
}

export default CategorySlider;
