const VendorDashboard = (): JSX.Element => (
  <div className="space-y-6">
    <div>
      <h1 className="text-2xl font-bold text-primary">Dashboard</h1>
      <p className="text-sm text-muted mt-1">Welcome back.</p>
    </div>
    <div className="grid grid-cols-3 gap-4">
      {['Orders Today', 'Revenue', 'Customers'].map((label) => (
        <div key={label} className="bg-white rounded-xl p-5 border border-black/5 shadow-card">
          <p className="text-xs font-semibold uppercase tracking-wide text-muted mb-2">{label}</p>
          <p className="text-3xl font-bold text-primary">—</p>
        </div>
      ))}
    </div>
  </div>
);

export default VendorDashboard;
