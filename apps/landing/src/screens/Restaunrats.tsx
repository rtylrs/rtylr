import React from "react";
import useAsync from "../../hooks/useAsync";
import { businessApi } from "../../api";

const Restaunrats = (): JSX.Element => {
  const [restaurants, setRestaurants] = React.useState<any>([]);

  const getAllRestaurants = useAsync({
    fn: businessApi.getAll,
    onSuccess: () => {
      setRestaurants(getAllRestaurants.result);
    },
    onError: () => {
      alert(getAllRestaurants.error);
    },
    callOnMount: true,
  });

  return (
    <div>
      <h1>Restaurants</h1>
      {restaurants?.map((restaurant) => (
        <div key={restaurant._id}>
          <h2>
            {restaurant.name} - {restaurant._id}
          </h2>
        </div>
      ))}
    </div>
  );
};

export default Restaunrats;
