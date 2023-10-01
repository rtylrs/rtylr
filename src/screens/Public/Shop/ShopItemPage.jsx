import React from "react";
import Slider from "react-slick";
import "slick-carousel/slick/slick.css";
import "slick-carousel/slick/slick-theme.css";
import { db } from "../../../firebase"; // Import your Firebase configuration
import { useParams } from "react-router-dom";

function ItemPage() {
  const { itemId } = useParams();
  const [item, setItem] = React.useState({});
  const settings = {
    dots: true,
    infinite: true,
    speed: 500,
    slidesToShow: 1,
    slidesToScroll: 1,
    nextArrow: <div></div>,
    prevArrow: <div></div>,
    //
  };

  React.useEffect(() => {
    db.collection("Products")
      .doc(itemId)
      .get()
      .then((doc) => {
        if (doc.exists) {
          setItem(doc.data());
        } else {
          console.log("No such document!");
        }
      })
      .catch((error) => {
        console.log("Error getting document:", error);
      });
  }, [itemId]);

  return (
    <div className="item-page">
      <h1>{item.name}</h1>
      <div className="item-images">
        <Slider {...settings}>
          {item.images &&
            item.images.map((image, index) => (
              <div key={index}>
                <img src={image} alt={`Image ${index + 1}`} />
              </div>
            ))}
        </Slider>
      </div>
      <div className="item-info">
        <p>{item.creator}</p>
        <p>{item.price} KD</p>
      </div>
    </div>
  );
}

export default ItemPage;
