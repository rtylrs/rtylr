import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router-dom";
import { Categories } from "../../data/categories";
import { useEffect, useState } from "react";
import { db } from "../../firebase";
import PetitionModal from "../../components/PetitionModal";

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faRecycle,
  faLeaf,
  faTint,
  faShoppingCart,
  faHandHoldingHeart,
  faUsers,
} from "@fortawesome/free-solid-svg-icons";
import ProductCard from "../../components/ProductCard";
import { useSelector } from "react-redux";

function Home() {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [blogs, setBlogs] = useState([]);
  const [isPetitionModalOpen, setIsPetitionModalOpen] = useState(false);
  const [petitions, setPetitions] = useState([]);
  const [petitonId, setPetitionId] = useState("");
  const products = useSelector((state) => state.products.products);

  useEffect(() => {
    const petitionsUnsubscribe = db
      .collection("petitions")
      .limit(3)
      .onSnapshot((petitionsSnapshot) => {
        const petitionsData = petitionsSnapshot.docs.map((doc) => ({
          id: doc.id,
          ...doc.data(),
        }));
        setPetitions(petitionsData);
      });

    db.collection("Blogs")
      .limit(3)
      .get()
      .then((blogsSnapshot) => {
        const blogsData = blogsSnapshot.docs.map((doc) => ({
          id: doc.id,
          ...doc.data(),
        }));
        setBlogs(blogsData);
      })
      .catch((error) => {
        console.error("Error fetching blogs data:", error);
      });

    return () => {
      petitionsUnsubscribe();
    };
  }, []);

  return (
    <div>
      <PetitionModal
        isOpen={isPetitionModalOpen}
        onClose={() => {
          setIsPetitionModalOpen(false);
        }}
        petitionId={petitonId}
      />
      <div className="Hero">
        <div className="hero-text">
          <h1>{t("Hero.welcome")}</h1>
          <h2>{t("Hero.discover")}</h2>
          <button className="hero-button" onClick={() => navigate("/shop")}>
            {t("Hero.shop")}
          </button>
        </div>
      </div>
      <section className="product-section">
        <h2 className="section-title">Our Featured Selections</h2>
        <div className="product-grid">
          {products &&
            products
              .slice(0, 4)
              .map((product) => (
                <ProductCard key={product.id} product={product} />
              ))}
        </div>
        <button
          className="view-more-button"
          onClick={() => {
            navigate("/shop");
          }}
        >
          View More
        </button>
      </section>

      <section className="home-section">
        <h2 className="section-title">Our Green Selection</h2>
        <div className="grid-container">
          {Categories.map((category, index) => (
            <Link
              key={index}
              className="grid-item"
              style={{
                backgroundImage: `url(${category.imageUrl})`,
              }}
              to={
                category.title.toLowerCase() === "brands"
                  ? `/brands`
                  : `/collection/${category.title.toLowerCase()}`
              }
            >
              <div className="item-content">
                <p>{category.title}</p>
              </div>
            </Link>
          ))}
        </div>
      </section>

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

      <section className="blog-section">
        <h2 className="section-title">Latest Blog Posts</h2>
        <div className="blog-grid">
          {blogs &&
            blogs.map((post, index) => (
              <Link className="blog-post" key={index} to={`/blog/${post.id}`}>
                <h3 className="post-title">
                  {post.title.length > 50
                    ? post.title.substring(0, 50) + "..."
                    : post.title}
                </h3>
                <p className="post-content">
                  {post.content.length > 250
                    ? post.content.substring(0, 250) + "..."
                    : post.content}
                </p>
                <p className="post-author">By {post?.ownerName}</p>
                <p className="post-date">
                  {post.timestamp.toDate().toDateString()}
                </p>
              </Link>
            ))}
        </div>
        <button
          className="view-more-blogs-button"
          onClick={() => {
            navigate("/all-blogs");
          }}
        >
          View More
        </button>
      </section>

      <section className="petition-section">
        <h2>Latest Petitions</h2>
        <div className="petition-list">
          {petitions.map((petition, index) => (
            <div className="petition-item" key={index}>
              <h3>{petition.title}</h3>
              <div className="progress-container">
                <div className="progress-section-one">
                  <div className="signature-progress">
                    {petition.signatures} / {petition.goal}
                  </div>
                  <div className="progress-bar">
                    <div
                      className="progress"
                      style={{
                        width: `${
                          petition.signatures >= petition.goal
                            ? "100%"
                            : `${Math.min(
                                (petition.signatures / petition.goal) * 100,
                                100
                              )}%`
                        }`,
                      }}
                    ></div>
                  </div>
                </div>
                <div className="progress-section-two">
                  <button
                    className="sign-button"
                    onClick={() => {
                      setPetitionId(petition.id);
                      setIsPetitionModalOpen(true);
                    }}
                  >
                    Sign Now
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
        <button
          className="view-more-blogs-button"
          onClick={() => {
            navigate("/all-petitions");
          }}
        >
          View All
        </button>
      </section>
    </div>
  );
}

export default Home;
