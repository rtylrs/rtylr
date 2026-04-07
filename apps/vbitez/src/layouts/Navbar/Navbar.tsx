import { SearchBar } from "./components";

const Navbar = () => {
  return (
    <div className="navbar">
      <div className="navbar__left" />

      {/* Centered Global Search */}
      <SearchBar />

      <div className="navbar__right" />
    </div>
  );
};

export default Navbar;