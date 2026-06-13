import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Lock, UserCog } from 'lucide-react';
import { useEmployeeSessionStore } from '../stores/employeeSessionStore';

interface Props {
  /** Short verb phrase, e.g. "apply discounts", "issue refunds". */
  action: string;
}

/**
 * Shown when the signed-in operator lacks the capability for an action. In the
 * unified identity model there is no separate manager PIN — a manager signs in
 * via the roster — so the remedy is to switch operators.
 */
const ManagerRequired: React.FC<Props> = ({ action }) => {
  const navigate = useNavigate();
  const endSession = useEmployeeSessionStore((s) => s.endSession);

  const switchOperator = (): void => {
    endSession();
    navigate('/');
  };

  return (
    <div className="flex flex-col items-center text-center gap-4 py-10 px-6 max-w-sm mx-auto">
      <div className="w-14 h-14 rounded-2xl bg-amber-500/10 flex items-center justify-center">
        <Lock className="w-6 h-6 text-amber-500" />
      </div>
      <div>
        <p className="text-sm font-semibold text-pos-text">Manager approval needed</p>
        <p className="text-xs text-pos-muted mt-1">
          You don't have permission to {action}. A manager or store manager must sign in.
        </p>
      </div>
      <button
        onClick={switchOperator}
        className="inline-flex items-center gap-2 h-11 px-5 rounded-xl bg-pos-text text-pos-bg font-semibold text-sm"
      >
        <UserCog className="w-4 h-4" />
        Switch operator
      </button>
    </div>
  );
};

export default ManagerRequired;
