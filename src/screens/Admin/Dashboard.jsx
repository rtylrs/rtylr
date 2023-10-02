import React from "react";
import { auth, db } from "../../firebase";

function Dashboard() {
  // eslint-disable-next-line no-unused-vars
  //const [allUsers, setAllUsers] = React.useState([]); //TODO: add the users to the dashboard
  const [allEvents, setAllEvents] = React.useState([]);
  const [users, setUsers] = React.useState(0);
  const [vendors, setVendors] = React.useState(0);
  const [reviews, setReviews] = React.useState(0);
  const [events, setEvents] = React.useState(0);
  const [clicks, setClicks] = React.useState(0);
  const [reviewsData, setReviewsData] = React.useState([]);

  React.useEffect(() => {
    (async () => {
      const users = await db
        .collection("Users")
        .where("isVendor", "==", false)
        .get();
      setUsers(users.size);

      // const Allusers = await db
      //   .collection("Users")
      //   .where("isVendor", "==", false)
      //   .get();
      // setAllUsers(Allusers.docs.map((doc) => doc.data()));

      const vendors = await db
        .collection("Users")
        .where("isVendor", "==", true)
        .get();
      setVendors(vendors.size);

      const reviews = await db.collection("Reviews").get();
      setReviewsData(reviews.docs.map((doc) => doc.data()));
      setReviews(reviews.size);

      const events = await db.collection("Events").get();
      setEvents(events.size);
      setAllEvents(events.docs.map((doc) => doc.data()));

      const clicks = await db.collection("Config").doc("WebsiteData").get();
      setClicks(clicks.data().visits);
    })();
  }, []);

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
            <p>Welcome to the admin dashboard</p>
          </div>
        </div>
        <div className="main__cards">
          <div className="card">
            <i className="fa fa-user"></i>
            <div className="card_inner">
              <p className="text-primary-p">Users</p>
              <span className="font-bold text-title">{users}</span>
            </div>
          </div>

          <div className="card">
            <i className="fa-solid fa-building"></i>
            <div className="card_inner">
              <p className="text-primary-p">Orders</p>
              <span className="font-bold text-title">{vendors}</span>
            </div>
          </div>

          <div className="card">
            <i className="fa-solid fa-calendar-days"></i>
            <div className="card_inner">
              <p className="text-primary-p">Pending Orders</p>
              <span className="font-bold text-title">{events}</span>
            </div>
          </div>

          <div className="card">
            <i className="fa-solid fa-heart"></i>
            <div className="card_inner">
              <p className="text-primary-p">Feedback</p>
              <span className="font-bold text-title">{reviews}</span>
            </div>
          </div>
        </div>

        <div className="charts">
          <div className="charts__left">
            <div className="charts__left__title">
              <div className="InnerContainerAdminCharts">
                <h1>Latest Reviews</h1>
                {reviewsData.map((review, index) => (
                  <div key={index}>
                    <h3>
                      {review.customerName}-{review.review}
                    </h3>
                    <p>Vendor: {review.vendorName}</p>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <div className="charts__right">
            <div className="charts__right__title">
              <div>
                <h1>Stats Reports</h1>
              </div>
            </div>

            <div className="charts__right__cards">
              <div className="card1">
                <h1>Website Clicks</h1>
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
