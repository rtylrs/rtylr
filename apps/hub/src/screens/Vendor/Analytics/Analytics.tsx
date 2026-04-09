import React, { useState } from "react";
import { useGetAnalytics, useGetRevenue, useGetOrders } from "../../../hooks/queries/useAnalytics";
import { useUserStore } from "../../../stores/userStore";
import { Line, Bar } from "react-chartjs-2";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from "chart.js";

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend
);

const Analytics: React.FC = () => {
  const { userId } = useUserStore();
  const [dateRange, setDateRange] = useState<{ start: string; end: string } | undefined>(undefined);

  const { data: analyticsData, isLoading: analyticsLoading } = useGetAnalytics(userId, dateRange);
  const { data: revenueData, isLoading: revenueLoading } = useGetRevenue(userId, dateRange);
  const { data: ordersData, isLoading: ordersLoading } = useGetOrders(userId, dateRange);

  const handleDateRangeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    const today = new Date();
    let start: Date;

    switch (value) {
      case "7d":
        start = new Date(today.getTime() - 7 * 24 * 60 * 60 * 1000);
        break;
      case "30d":
        start = new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000);
        break;
      case "90d":
        start = new Date(today.getTime() - 90 * 24 * 60 * 60 * 1000);
        break;
      default:
        setDateRange(undefined);
        return;
    }

    setDateRange({
      start: start.toISOString().split("T")[0],
      end: today.toISOString().split("T")[0],
    });
  };

  // Mock data for charts (replace with actual data from API)
  const revenueChartData = revenueData || [
    { date: "Mon", revenue: 1200 },
    { date: "Tue", revenue: 1900 },
    { date: "Wed", revenue: 3000 },
    { date: "Thu", revenue: 2780 },
    { date: "Fri", revenue: 1890 },
    { date: "Sat", revenue: 2390 },
    { date: "Sun", revenue: 3490 },
  ];

  const ordersChartData = ordersData || [
    { date: "Mon", orders: 12 },
    { date: "Tue", orders: 19 },
    { date: "Wed", orders: 30 },
    { date: "Thu", orders: 28 },
    { date: "Fri", orders: 18 },
    { date: "Sat", orders: 24 },
    { date: "Sun", orders: 35 },
  ];

  const metrics = analyticsData || {
    totalRevenue: 15670,
    totalOrders: 162,
    averageOrderValue: 96.73,
    totalVisits: 1240,
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-primary">Analytics</h1>
        <div>
          <select
            onChange={handleDateRangeChange}
            defaultValue=""
            className="border border-black/10 rounded-lg px-3 py-1.5 text-sm bg-white"
          >
            <option value="">All Time</option>
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
            <option value="90d">Last 90 Days</option>
          </select>
        </div>
      </div>

      {analyticsLoading ? (
        <div className="text-sm text-muted">Loading analytics...</div>
      ) : (
        <>
          <div className="grid grid-cols-4 gap-4">
            <div className="bg-white rounded-xl p-5 border border-black/5 shadow-card">
              <h3 className="text-xs font-semibold uppercase tracking-wide text-muted mb-2">Total Revenue</h3>
              <p className="text-3xl font-bold text-primary">${metrics.totalRevenue.toLocaleString()}</p>
            </div>
            <div className="bg-white rounded-xl p-5 border border-black/5 shadow-card">
              <h3 className="text-xs font-semibold uppercase tracking-wide text-muted mb-2">Total Orders</h3>
              <p className="text-3xl font-bold text-primary">{metrics.totalOrders}</p>
            </div>
            <div className="bg-white rounded-xl p-5 border border-black/5 shadow-card">
              <h3 className="text-xs font-semibold uppercase tracking-wide text-muted mb-2">Average Order Value</h3>
              <p className="text-3xl font-bold text-primary">${metrics.averageOrderValue.toFixed(2)}</p>
            </div>
            <div className="bg-white rounded-xl p-5 border border-black/5 shadow-card">
              <h3 className="text-xs font-semibold uppercase tracking-wide text-muted mb-2">Total Visits</h3>
              <p className="text-3xl font-bold text-primary">{metrics.totalVisits}</p>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="bg-white rounded-xl p-5 border border-black/5 shadow-card">
              <h2 className="text-base font-semibold text-primary mb-4">Revenue Over Time</h2>
              <div className="h-64">
                <Line
                  data={{
                    labels: revenueChartData.map((d) => d.date),
                    datasets: [
                      {
                        label: "Revenue",
                        data: revenueChartData.map((d) => d.revenue),
                        borderColor: "#4485c4",
                        backgroundColor: "rgba(68, 133, 196, 0.1)",
                        tension: 0.4,
                      },
                    ],
                  }}
                  options={{
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: { legend: { display: true } },
                    scales: { y: { beginAtZero: true } },
                  }}
                />
              </div>
            </div>

            <div className="bg-white rounded-xl p-5 border border-black/5 shadow-card">
              <h2 className="text-base font-semibold text-primary mb-4">Orders Over Time</h2>
              <div className="h-64">
                <Bar
                  data={{
                    labels: ordersChartData.map((d) => d.date),
                    datasets: [
                      {
                        label: "Orders",
                        data: ordersChartData.map((d) => d.orders),
                        backgroundColor: "#4485c4",
                      },
                    ],
                  }}
                  options={{
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: { legend: { display: true } },
                    scales: { y: { beginAtZero: true } },
                  }}
                />
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default Analytics;
