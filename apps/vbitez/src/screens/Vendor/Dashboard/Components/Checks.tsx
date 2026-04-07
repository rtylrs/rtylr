import { Bomb, TriangleAlert } from "lucide-react";
import React from "react";

const Checks: React.FC = () => {
  const [majorIssue, setMajorIssue] = React.useState<number>(0);
  const [warnings, setWarnings] = React.useState<number>(0);
  return (
    <div className="checks-container">
      <div className="check-item">
        <Bomb size={16} />
        <p>
          You have{" "}
          <span className="major-issue">
            {majorIssue} major {majorIssue === 1 ? "issue" : "issues"}
          </span>
        </p>
      </div>
      <div className="check-item">
        <TriangleAlert size={16} />
        <p>
          You have{" "}
          <span className="warning">
            {warnings} {warnings === 1 ? "warning" : "warnings"}
          </span>
        </p>
      </div>
    </div>
  );
};

export default Checks;
