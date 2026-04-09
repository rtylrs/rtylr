import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { DotsThreeOutline, GearSix, SignOut } from 'phosphor-react';

const SidebarUser = () => {
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();
  const ref = useRef<HTMLDivElement>(null);

  const authUser = (() => {
    try { return JSON.parse(localStorage.getItem('authUser') || 'null'); } catch { return null; }
  })();
  const initial = String(authUser?.name || authUser?.email || '?').trim().charAt(0).toUpperCase();

  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => { if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false); };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open]);

  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('authUser');
    navigate('/auth/login', { replace: true });
  };

  return (
    <div className="flex items-center gap-2.5 px-3 py-3 border-t border-white/10 bg-white/5 shrink-0 relative" ref={ref}>
      <div className="w-9 h-9 rounded-full bg-accent text-white flex items-center justify-center font-bold text-base shrink-0">
        {initial}
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-[13px] font-semibold text-white truncate">{authUser?.name || 'User'}</p>
        <p className="text-[11px] text-white/60 truncate">{authUser?.email || 'user@rtylr.com'}</p>
      </div>
      <button
        className="p-1.5 rounded-md text-white/60 hover:bg-white/10 hover:text-white transition-colors"
        onClick={() => setOpen((v) => !v)}
      >
        <DotsThreeOutline size={16} />
      </button>

      {open && (
        <div className="absolute bottom-full right-2 mb-2 w-48 bg-admin border border-white/10 rounded-xl shadow-dropdown overflow-hidden z-50 animate-dropdown-in">
          <button
            className="flex items-center gap-2.5 w-full px-4 py-2.5 text-sm text-white/85 hover:bg-white/8 transition-colors"
            onClick={() => { navigate('/settings'); setOpen(false); }}
          >
            <GearSix size={15} />
            Settings
          </button>
          <div className="h-px bg-white/10" />
          <button
            className="flex items-center gap-2.5 w-full px-4 py-2.5 text-sm text-red-400 hover:bg-red-400/10 transition-colors"
            onClick={logout}
          >
            <SignOut size={15} />
            Log out
          </button>
        </div>
      )}
    </div>
  );
};

export default SidebarUser;
