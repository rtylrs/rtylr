import { describe, expect, it } from 'vitest';
import { dashboardWidgetSchema, hrDashboardStatsSchema } from './dashboard';

describe('HR dashboard wire numbers', () => {
  it('coerces numeric strings and replaces invalid values', () => {
    const result = hrDashboardStatsSchema.parse({
      totalStaff: '12',
      onShiftToday: '7',
      openPositions: 'invalid',
      monthlyPayroll: '25000.50',
      todayAtGlance: {
        shiftsScheduled: '9',
        newApplications: '2',
        pendingApprovals: '3',
      },
    });

    expect(result).toMatchObject({
      totalStaff: 12,
      onShiftToday: 7,
      openPositions: 0,
      monthlyPayroll: 25000.5,
      todayAtGlance: { shiftsScheduled: 9, newApplications: 2, pendingApprovals: 3 },
    });
  });

  it('coerces persisted widget coordinates and identifiers', () => {
    const result = dashboardWidgetSchema.parse({
      id: '1',
      ownerId: '2',
      roleId: '3',
      widgetType: 'headcount',
      positionX: '0',
      positionY: '1',
      width: '4',
      height: '2',
      createdAt: '100',
      updatedAt: '200',
    });

    expect(result).toMatchObject({ id: 1, ownerId: 2, width: 4, height: 2 });
  });
});
