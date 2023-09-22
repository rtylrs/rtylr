function Home() {
  return (
    <div>
      <div className="Hero">
        <div className="hero-text">
          <h1>Welcome to Sanity!</h1>
          <h2>
            Discover what you're seeking or explore the endless possibilities!
            Your ultimate destination for unforgettable experiences.
          </h2>
          <div className="searchBarHome">
            <input
              type="text"
              placeholder="Search for products..."
              className="searchBar"
            />
          </div>
        </div>
      </div>
    </div>
  );
}

export default Home;
