import React from "react";
import { auth } from "../../firebase";
import {
  FaSuitcase,
  FaShoppingBag,
  FaStarHalfAlt,
  FaUser,
} from "react-icons/fa";
import { adminApi } from "../../api";
import useAsync from "../../hooks/useAsync";

function Dashboard() {
  const [allEvents, setAllEvents] = React.useState([]);
  const [users, setUsers] = React.useState(0);
  const [business, setBusiness] = React.useState(0);
  const [reviews, setReviews] = React.useState(0);
  const [events, setEvents] = React.useState(0);
  const [clicks, setClicks] = React.useState(0);
  const [reviewsData, setReviewsData] = React.useState([]);

  const getUsersCount = useAsync({
    fn: adminApi.getUserCount,
    onSuccess: () => {
      setUsers(getUsersCount.result?.count);
    },
    onError: () => {
      alert(getUsersCount.error);
    },
    callOnMount: true,
  });

  const getBusinessCount = useAsync({
    fn: adminApi.getBusinessCount,
    onSuccess: () => {
      setBusiness(getBusinessCount.result?.count);
    },
    onError: () => {
      alert(getBusinessCount.error);
    },
    callOnMount: true,
  });

  return (
    <div className="AdminDash">
      <div className="main__container">
        <div className="main__title">
          <div className="main__greeting">
            <h1>
              Hello{" "}
              {auth.currentUser?.displayName &&
                auth.currentUser.displayName.split(" ")[0]}
              ,
            </h1>
            <p>Welcome to your admin dashboard</p>
          </div>
        </div>
        <div className="main__cards">
          <div className="card">
            <i className="fa fa-user"></i>
            <div className="card_inner">
              <div>
                <FaUser />
                <p className="text-primary-p">Users</p>
              </div>
              <span className="font-bold text-title">{users}</span>
            </div>
          </div>

          <div className="card">
            <i className="fa-solid fa-building"></i>
            <div className="card_inner">
              <div>
                <FaShoppingBag />
                <p className="text-primary-p">Restaurants</p>
              </div>

              <span className="font-bold text-title">{business}</span>
            </div>
          </div>

          <div className="card">
            <i className="fa-solid fa-calendar-days"></i>
            <div className="card_inner">
              <div>
                <FaSuitcase />
                <p className="text-primary-p">Pending Orders</p>
              </div>
              <span className="font-bold text-title">{events}</span>
            </div>
          </div>

          <div className="card">
            <i className="fa-solid fa-heart"></i>
            <div className="card_inner">
              <div>
                <FaStarHalfAlt />
                <p className="text-primary-p">Avg Rating</p>
              </div>
              <span className="font-bold text-title">{reviews}</span>
            </div>
          </div>
        </div>

        <div className="charts">
          <div className="charts__left">
            <div className="charts__left__title">
              <div>
                <h1>
                  Latest
                  <FaStarHalfAlt
                    style={{ marginLeft: "10px" }}
                    size="xs"
                  />
                </h1>
                {reviewsData.map((review, index) => (
                  <div key={index}>
                    <h3>
                      By {review.name} - {review.rating}/5 Stars
                    </h3>
                    <p>{review.email}</p>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <div className="charts__right">
            <div className="charts__right__title">
              <div>
                <h1>Highlights</h1>
              </div>
            </div>

            <div className="charts__right__cards">
              <div className="card1">
                <h1>Warnings</h1>
                <p>{clicks}</p>
              </div>
              <div className="card2">
                <h1>Order's today</h1>
                <p>{clicks}</p>
              </div>
              <div className="card3">
                <h1>Petitions</h1>
                <p>{clicks}</p>
              </div>

              <div className="card4">
                <h1>Pending Events</h1>
                <p>
                  {
                    allEvents.filter(
                      (event) =>
                        event.date &&
                        event.date.toDate() >
                          new Date(new Date().getTime() - 1000 * 60 * 60 * 24)
                    ).length
                  }
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Dashboard;
