import { getPageTitle } from './routeLabels';

describe('getPageTitle', () => {
  it('prefers the most specific nested route label', () => {
    expect(getPageTitle('/sales/leads/42')).toBe('Leads');
    expect(getPageTitle('/sales/contracts/42')).toBe('Contracts');
  });

  it('does not treat similarly named paths as nested routes', () => {
    expect(getPageTitle('/salesforce')).toBe('CRM');
  });
});
