const Pricing = () => {
  return (
    <div className="pricing-page">
      <h1>Pricing Plans</h1>
      <div className="pricing-table">
        <table>
          <thead>
            <tr>
              <th>Products</th>
              <th className="free-plan">Free Plan</th>
              <th className="pro-plan">Basic Plan</th>
              <th className="enterprise-plan">Ultimate Plan</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>A/B Testing</td>
              <td>No-cost</td>
              <td>No-cost</td>
              <td>No-cost</td>
            </tr>
            <tr>
              <td>Analytics</td>
              <td>No-cost</td>
              <td>No-cost</td>
              <td>No-cost</td>
            </tr>
            <tr>
              <td>App Check</td>
              <td>No-cost, subject to quotas and limits</td>
              <td>No-cost, subject to quotas and limits</td>
              <td>No-cost, subject to quotas and limits</td>
            </tr>
            <tr>
              <td>App Distribution</td>
              <td>No-cost</td>
              <td>No-cost</td>
              <td>No-cost</td>
            </tr>
            <tr>
              <td>App Hosting</td>
              <td>Not applicable</td>
              <td>
                No cost up to 10GB/month
                <br />
                Then $0.10/GB
              </td>
              <td>
                No cost up to 50GB/month
                <br />
                Then $0.05/GB
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default Pricing;
