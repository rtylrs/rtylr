import { ChevronDown, ChevronUp } from 'lucide-react';
import React from 'react';
import { useNavigate } from 'react-router-dom';

const FreeAccount: React.FC<{}> = () => {
  const [open, setOpen] = React.useState<boolean>(true);
  const navigate = useNavigate();

  return (
    <div
      className="free-account-popup"
      style={{
        height: open ? 'auto' : '50px',
      }}
    >
      <div className="free-account-header">
        <button className="dropdown-button" onClick={() => setOpen(!open)}>
          {open ? <ChevronDown size={16} /> : <ChevronUp size={16} />}
        </button>
      </div>
      <div
        style={{
          display: open ? 'block' : 'none',
        }}
      >
        <div className="free-account-content">
          <p>You are currently using a free tier.</p>
        </div>
        <button className="upgrade-button" onClick={() => navigate('/settings/billing')}>
          Select a Plan
        </button>
      </div>
    </div>
  );
};

export default FreeAccount;
