import React from "react";
import { Line } from "react-chartjs-2";
import { useSelector } from "react-redux";
import { RootState } from "../../../../redux/store";

const Visits: React.FC<{}> = ({}) => {
  const visitData = useSelector((state: RootState) => state.user.visits);

  const chartData = {
    labels: visitData.map((visit) => new Date(visit.date).toLocaleDateString()),
    datasets: [
      {
        label: "Number of Visits",
        data: visitData.map((visit) => visit.visits),
        fill: false,
        backgroundColor: "rgba(75,192,192,0.4)",
        borderColor: "rgba(75,192,192,1)",
      },
    ],
  };

  const options = {
    responsive: true,
    scales: {
      y: {
        beginAtZero: true,
        ticks: {
          stepSize: 1,
        },
      },
    },
    plugins: {
      legend: {
        display: false,
      },
    },
  };

  return (
    <div className="visits-container">
      <div>
        <Line data={chartData} options={options} />
      </div>
    </div>
  );
};

export default Visits;
