import { Scheduler } from "@aldabil/react-scheduler";

const events = [
  {
    id: 1,
    title: "Meeting",
    start: new Date("2023-10-15T10:00:00"),
    end: new Date("2023-10-15T11:00:00"),
  },
  {
    id: 2,
    title: "Lunch",
    start: new Date("2023-10-15T12:00:00"),
    end: new Date("2023-10-15T13:00:00"),
  },
];

const Events = () => {
  const yesterday = new Date();
  yesterday.setDate(yesterday.getDate() - 1);

  return (
    <div>
      <div className="Events">
        <div className="Events-text">
          <h1>Championing Eco-Friendly Initiatives</h1>
          <h2>Embark on Your Path to a Sustainable Future</h2>
        </div>
      </div>
      <div className="SchedulerContainer">
        <Scheduler
          events={events}
          initialView="week"
          initialDate={yesterday}
          onEventClick={(e) => console.log(e)}
          editable={false}
          deletable={false}
          draggable={false}
          disableViewNavigator
        />
      </div>
    </div>
  );
};

export default Events;
