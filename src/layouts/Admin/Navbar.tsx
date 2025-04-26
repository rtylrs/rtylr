import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faBars } from "@fortawesome/free-solid-svg-icons";
import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { logout } from "../../redux/reducers/userSlice";
import { RootState } from "../../redux/store";

const Navbar = ({ openSidebar }) => {
  const dispatch = useDispatch();
  const user = useSelector((state: RootState) => state.user);

  useEffect(() => {}, []);

  return (
    <div className="navbar">
      <div className="nav_icon" onClick={() => openSidebar("navbar")}>
        <FontAwesomeIcon icon={faBars} />
      </div>

      <div className="navbar__left">
        <p>{user?.user?.name}</p>
      </div>
      <div className="navbar__right">
        <div
          onClick={() => {
            dispatch(logout());
          }}
        >
          <p>Log out</p>
        </div>
      </div>
    </div>
  );
};

export default Navbar;
