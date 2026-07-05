import type {
  HeadcountResponse,
  OperationsKPIResponse,
  ProductKPIResponse,
  ShiftMetricsResponse,
  TimeOffResponse,
  UtilizationResponse,
  RevenueSeriesResponse,
  RevenueByAppResponse,
  RevenueComparisonResponse,
  CustomerKPIResponse,
  TopCustomersResponse,
  CustomerLTVResponse,
  CustomerChurnResponse,
  CashFlowResponse,
  ExpensesByCategoryResponse,
  InvoiceAgingResponse,
} from '../hooks/queries/types';

const majorToMinor = (value: unknown): number => Math.round(Number(value ?? 0) * 100);

export const mapProductPerformance = (body: Record<string, unknown>): ProductKPIResponse => ({
  totalSkus: Number(body.totalSkus ?? 0),
  activeSkus: Number(body.activeSkus ?? 0),
  outOfStock: Number(body.outOfStock ?? 0),
  avgMarginPct: Number(body.avgMarginPct ?? 0),
});

export const mapOperationsSummary = (body: Record<string, unknown>): OperationsKPIResponse => ({
  activeStaff: Number(body.activeStaff ?? 0),
  shiftsThisWeek: Number(body.shiftsThisWeek ?? 0),
  pendingTimeOff: Number(body.pendingTimeOff ?? 0),
  openLeaveRequests: Number(body.openLeaveRequests ?? 0),
});

export const mapHeadcountMetrics = (body: Record<string, unknown>): HeadcountResponse => {
  const metrics = (body.metrics ?? {}) as Record<string, unknown>;
  const hired = Number(metrics.hiredThisPeriod ?? 0);
  const departed = Number(metrics.departedThisPeriod ?? 0);
  return {
    activeStaff: Number(metrics.current ?? 0),
    momDelta: hired - departed,
    momDeltaPct: Number(metrics.turnoverRate ?? 0),
  };
};

export const mapShiftMetrics = (body: Record<string, unknown>): ShiftMetricsResponse => {
  const metrics = (body.metrics ?? {}) as Record<string, unknown>;
  return {
    scheduled: Number(metrics.totalShifts ?? 0),
    completed: Number(metrics.covered ?? 0),
    noShows: Number(metrics.missed ?? 0),
    completionRatePct: Number(metrics.coverageRate ?? 0),
  };
};

export const mapTimeOffMetrics = (body: Record<string, unknown>): TimeOffResponse => {
  const metrics = (body.metrics ?? {}) as Record<string, unknown>;
  return {
    pending: Number(metrics.pending ?? 0),
    approved: Number(metrics.approved ?? 0),
  };
};

export const mapUtilizationMetrics = (body: Record<string, unknown>): UtilizationResponse => {
  const metrics = (body.metrics ?? {}) as Record<string, unknown>;
  return {
    scheduledHours: Number(metrics.totalHours ?? 0),
    availableHours: Number(metrics.availableHours ?? 0),
    utilizationPct: Number(metrics.avgUtilization ?? 0),
  };
};

export const mapRevenueOverview = (body: Record<string, unknown>): RevenueSeriesResponse => {
  const overview = (body.overview ?? {}) as Record<string, unknown>;
  const trend = Array.isArray(overview.trend) ? overview.trend : [];
  return {
    series: trend.map((point) => {
      const row = point as Record<string, unknown>;
      return {
        date: String(row.date ?? ''),
        amountMinorUnits: majorToMinor(row.amount),
      };
    }),
  };
};

export const mapRevenueByApp = (body: Record<string, unknown>): RevenueByAppResponse => {
  const apps = Array.isArray(body.apps) ? body.apps : [];
  return {
    byApp: apps.map((entry) => {
      const row = entry as Record<string, unknown>;
      return {
        app: String(row.app ?? ''),
        amountMinorUnits: majorToMinor(row.amount),
        pct: Number(row.percentage ?? 0),
      };
    }),
  };
};

export const mapRevenueComparison = (body: Record<string, unknown>): RevenueComparisonResponse => {
  const comparisons = Array.isArray(body.comparisons) ? body.comparisons : [];
  const first = (comparisons[0] ?? {}) as Record<string, unknown>;
  return {
    thisPeriod: majorToMinor(first.current),
    lastPeriod: majorToMinor(first.previous),
    deltaPct: Number(first.change ?? 0),
  };
};

export const mapCustomerAnalytics = (body: Record<string, unknown>): CustomerKPIResponse => {
  const analytics = (body.analytics ?? {}) as Record<string, unknown>;
  const total = Number(analytics.total ?? 0);
  const churnRate = Number(analytics.churnRate ?? 0);
  const retentionRate = Number(analytics.retentionRate ?? 0);
  return {
    total,
    newThisMonth: Number(analytics.newThisPeriod ?? 0),
    active: Math.round((total * retentionRate) / 100),
    churned: Math.round((total * churnRate) / 100),
  };
};

export const mapTopCustomers = (body: Record<string, unknown>): TopCustomersResponse => {
  const customers = Array.isArray(body.customers) ? body.customers : [];
  return {
    customers: customers.map((entry) => {
      const row = entry as Record<string, unknown>;
      return {
        customerId: Number(row.id ?? 0),
        name: String(row.name ?? ''),
        email: '',
        lifetimeSpendMinorUnits: majorToMinor(row.totalSpent),
        orderCount: Number(row.orderCount ?? 0),
      };
    }),
  };
};

export const mapCustomerLTV = (body: Record<string, unknown>): CustomerLTVResponse => {
  const analysis = (body.analysis ?? {}) as Record<string, unknown>;
  const segments = Array.isArray(analysis.segments) ? analysis.segments : [];
  return {
    buckets: segments.map((entry) => {
      const row = entry as Record<string, unknown>;
      const avg = majorToMinor(row.avgLtv);
      return {
        label: String(row.segment ?? ''),
        lowerMinorUnits: avg,
        upperMinorUnits: avg,
        customerCount: Number(row.customerCount ?? 0),
      };
    }),
  };
};

export const mapCustomerChurn = (body: Record<string, unknown>): CustomerChurnResponse => {
  const analysis = (body.analysis ?? {}) as Record<string, unknown>;
  return {
    windows: [
      {
        windowDays: 30,
        churnedCount: Number(analysis.churnedCount ?? 0),
        activeCount: 0,
        churnPct: Number(analysis.churnRate ?? 0),
      },
    ],
  };
};

export const mapCashFlowAnalytics = (body: Record<string, unknown>): CashFlowResponse => {
  const analytics = (body.analytics ?? {}) as Record<string, unknown>;
  const trend = Array.isArray(analytics.trend) ? analytics.trend : [];
  return {
    series: trend.map((entry) => {
      const row = entry as Record<string, unknown>;
      return {
        date: String(row.date ?? ''),
        inflowMinorUnits: majorToMinor(row.inflow),
        outflowMinorUnits: majorToMinor(row.outflow),
      };
    }),
  };
};

export const mapExpenseAnalytics = (body: Record<string, unknown>): ExpensesByCategoryResponse => {
  const analytics = (body.analytics ?? {}) as Record<string, unknown>;
  const categories = Array.isArray(analytics.byCategory) ? analytics.byCategory : [];
  return {
    categories: categories.map((entry, index) => {
      const row = entry as Record<string, unknown>;
      return {
        categoryId: index + 1,
        categoryName: String(row.category ?? ''),
        amountMinorUnits: majorToMinor(row.amount),
      };
    }),
  };
};

export const mapInvoiceAging = (body: Record<string, unknown>): InvoiceAgingResponse => {
  const aging = (body.aging ?? {}) as Record<string, unknown>;
  const buckets = Array.isArray(aging.buckets) ? aging.buckets : [];
  return {
    buckets: buckets.map((entry) => {
      const row = entry as Record<string, unknown>;
      return {
        bucket: String(row.bucket ?? ''),
        invoiceCount: Number(row.count ?? 0),
        amountMinorUnits: majorToMinor(row.amount),
      };
    }),
  };
};
