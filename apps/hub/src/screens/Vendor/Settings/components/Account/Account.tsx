import React, { useState, useEffect } from "react";
import { useUserStore } from "../../../../../stores/userStore";

interface LoginHistoryItem {
  email: string;
  method: string;
  ip: string;
  timeAgo: string;
}

const Account: React.FC = () => {
  const { user } = useUserStore();
  const [formData, setFormData] = useState({ firstName: "", lastName: "", email: "" });
  const [isEditing, setIsEditing] = useState(false);
  const [loginHistory] = useState<LoginHistoryItem[]>([
    { email: "mohamad.homsi@anghami.com", method: "Logged in via Google", ip: "from 141.105.87.106", timeAgo: "3 days ago" },
    { email: "mohamad.homsi@anghami.com", method: "Logged in via Google", ip: "from 141.105.87.106", timeAgo: "2 weeks ago" },
    { email: "mohamad.homsi@anghami.com", method: "Logged in via Google", ip: "from 141.105.87.106", timeAgo: "2 weeks ago" },
  ]);
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 3;

  useEffect(() => {
    if (user) {
      setFormData({
        firstName: user.firstName || user.name?.split(" ")[0] || "",
        lastName: user.lastName || user.name?.split(" ").slice(1).join(" ") || "",
        email: user.email || "",
      });
    }
  }, [user]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    setIsEditing(true);
  };

  const handleSave = () => {
    console.log("Saving:", formData);
    setIsEditing(false);
  };

  const handleCancel = () => {
    if (user) {
      setFormData({
        firstName: user.firstName || user.name?.split(" ")[0] || "",
        lastName: user.lastName || user.name?.split(" ").slice(1).join(" ") || "",
        email: user.email || "",
      });
    }
    setIsEditing(false);
  };

  const totalPages = Math.ceil(loginHistory.length / itemsPerPage);
  const currentHistory = loginHistory.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

  const inputCls = "w-full border border-black/10 rounded-lg px-3 py-2 text-sm text-primary bg-white focus:outline-none focus:border-accent transition-colors";
  const labelCls = "block text-sm font-medium text-primary mb-1.5";

  return (
    <div className="max-w-2xl space-y-8">
      <div>
        <h1 className="text-xl font-bold text-primary">Account</h1>
      </div>

      {/* Profile */}
      <div className="bg-white rounded-xl border border-black/8 p-6 space-y-5">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label htmlFor="firstName" className={labelCls}>First name</label>
            <input type="text" id="firstName" name="firstName" value={formData.firstName} onChange={handleInputChange} className={inputCls} />
          </div>
          <div>
            <label htmlFor="lastName" className={labelCls}>Last name</label>
            <input type="text" id="lastName" name="lastName" value={formData.lastName} onChange={handleInputChange} className={inputCls} />
          </div>
        </div>
        <div>
          <label htmlFor="email" className={labelCls}>Email</label>
          <input type="email" id="email" name="email" value={formData.email} onChange={handleInputChange} className={inputCls} />
        </div>
        <div className="flex justify-end gap-3 pt-1">
          <button type="button" onClick={handleCancel} disabled={!isEditing} className="px-4 py-2 text-sm rounded-lg border border-black/10 text-primary hover:bg-black/4 transition-colors disabled:opacity-40">
            Cancel
          </button>
          <button type="button" onClick={handleSave} disabled={!isEditing} className="px-4 py-2 text-sm rounded-lg bg-admin text-white hover:opacity-90 transition-opacity disabled:opacity-40">
            Save changes
          </button>
        </div>
      </div>

      {/* Invalidate sessions */}
      <div className="bg-white rounded-xl border border-black/8 p-6 space-y-3">
        <h2 className="text-base font-semibold text-primary">Invalidate sessions</h2>
        <p className="text-sm text-muted">Invalidate all sessions for your account. This will log you out of your current session.</p>
        <button type="button" className="px-4 py-2 text-sm rounded-lg bg-destructive text-white hover:opacity-90 transition-opacity">
          Invalidate all account sessions
        </button>
      </div>

      {/* Login history */}
      <div className="bg-white rounded-xl border border-black/8 overflow-hidden">
        <div className="px-6 py-4 border-b border-black/8">
          <h2 className="text-base font-semibold text-primary">Login history</h2>
        </div>
        <table className="w-full text-sm">
          <thead className="bg-black/2 border-b border-black/8">
            <tr>
              {["Email", "Method", "IP Address", "Time"].map((h) => (
                <th key={h} className="px-6 py-3 text-left text-xs font-semibold text-muted uppercase tracking-wider">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-black/5">
            {currentHistory.map((item, i) => (
              <tr key={i} className="hover:bg-black/2 transition-colors">
                <td className="px-6 py-3 text-primary">{item.email}</td>
                <td className="px-6 py-3 text-muted">{item.method}</td>
                <td className="px-6 py-3 text-muted">{item.ip}</td>
                <td className="px-6 py-3 text-muted">{item.timeAgo}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-1.5 px-6 py-3 border-t border-black/8">
            <button onClick={() => setCurrentPage((p) => Math.max(1, p - 1))} disabled={currentPage === 1} className="px-3 py-1.5 text-sm rounded-lg border border-black/10 disabled:opacity-40 hover:bg-black/5">←</button>
            {Array.from({ length: totalPages }, (_, i) => i + 1).map((page) => (
              <button key={page} onClick={() => setCurrentPage(page)} className={`px-3 py-1.5 text-sm rounded-lg border ${currentPage === page ? 'border-admin bg-admin text-white' : 'border-black/10 hover:bg-black/5'}`}>{page}</button>
            ))}
            <button onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))} disabled={currentPage === totalPages} className="px-3 py-1.5 text-sm rounded-lg border border-black/10 disabled:opacity-40 hover:bg-black/5">→</button>
          </div>
        )}
      </div>
    </div>
  );
};

export default Account;
