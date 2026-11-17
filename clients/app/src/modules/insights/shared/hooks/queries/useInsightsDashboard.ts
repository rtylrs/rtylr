import { createOneKitClient, toFiniteNumber } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { DashboardServiceClient } from '@voxire/voxie/insights/dashboard/v1/client';
import type {
  DashboardProjectOption,
  HoursSummaryResponse,
  InsightsDashboardResponse,
  ProjectDashboardResponse,
  SalesSummaryResponse,
  WorkSummaryResponse,
} from './types';

const dashboardClient = createOneKitClient(DashboardServiceClient, 'insights');

export const useInsightsDashboard = (): UseQueryResult<InsightsDashboardResponse> => {
  return useQuery<InsightsDashboardResponse>({
    queryKey: ['insights-dashboard'],
    refetchInterval: 60000,
    queryFn: async (): Promise<InsightsDashboardResponse> => {
      const res = await dashboardClient.getDashboardSummary({});
      return {
        revenue: {
          value: toFiniteNumber(res.revenue?.value),
          deltaPct: toFiniteNumber(res.revenue?.deltaPct),
        },
        customers: {
          value: toFiniteNumber(res.customers?.value),
          deltaPct: toFiniteNumber(res.customers?.deltaPct),
        },
        orders: {
          value: toFiniteNumber(res.orders?.value),
          deltaPct: toFiniteNumber(res.orders?.deltaPct),
        },
        margin: {
          value: toFiniteNumber(res.margin?.value),
          deltaPct: toFiniteNumber(res.margin?.deltaPct),
        },
      };
    },
  });
};

// Business-wide Work (Flow) totals and per-project progress, so a service
// business sees its delivery on the overview, not only POS and invoice cards.
export const useWorkSummary = (): UseQueryResult<WorkSummaryResponse> => {
  return useQuery<WorkSummaryResponse>({
    queryKey: ['insights-dashboard-work'],
    refetchInterval: 60000,
    queryFn: async (): Promise<WorkSummaryResponse> => {
      const [stats, roadmap] = await Promise.all([
        dashboardClient.getInsights({}),
        dashboardClient.getRoadmap({}),
      ]);
      return {
        totalTasks: toFiniteNumber(stats.totalTasks),
        completedTasks: toFiniteNumber(stats.completedTasks),
        overdueTasks: toFiniteNumber(stats.overdueTasks),
        projects: (roadmap.projects ?? [])
          .map((p) => ({
            id: p.projectId ?? '',
            name: p.projectName ?? '',
            totalTasks: toFiniteNumber(p.totalTasks),
            completedTasks: toFiniteNumber(p.completedTasks),
            progressPct: toFiniteNumber(p.progressPercent),
          }))
          .filter((p) => p.totalTasks > 0),
      };
    },
  });
};

// CRM open pipeline and the deal value won this month, compared with the same
// span last month, plus recurring revenue from won monthly and yearly deals.
export const useSalesSummary = (): UseQueryResult<SalesSummaryResponse> => {
  return useQuery<SalesSummaryResponse>({
    queryKey: ['insights-dashboard-sales'],
    refetchInterval: 60000,
    queryFn: async (): Promise<SalesSummaryResponse> => {
      const res = await dashboardClient.getSalesSummary({});
      return {
        openPipelineValue: toFiniteNumber(res.openPipelineValue),
        openDealCount: toFiniteNumber(res.openDealCount),
        won: {
          value: toFiniteNumber(res.won?.value),
          deltaPct: toFiniteNumber(res.won?.deltaPct),
        },
        recentWonDealCount: toFiniteNumber(res.recentWonDealCount),
        mrr: {
          value: toFiniteNumber(res.mrr?.value),
          deltaPct: toFiniteNumber(res.mrr?.deltaPct),
        },
        activeContractCount: toFiniteNumber(res.activeContractCount),
      };
    },
  });
};

// Minutes logged in Work this month to date, compared with the same span last
// month, and how much of it went to each public project.
export const useHoursSummary = (): UseQueryResult<HoursSummaryResponse> => {
  return useQuery<HoursSummaryResponse>({
    queryKey: ['insights-dashboard-hours'],
    refetchInterval: 60000,
    queryFn: async (): Promise<HoursSummaryResponse> => {
      const res = await dashboardClient.getHoursSummary({});
      return {
        logged: {
          value: toFiniteNumber(res.logged?.value),
          deltaPct: toFiniteNumber(res.logged?.deltaPct),
        },
        priorMinutes: toFiniteNumber(res.priorMinutes),
        projects: (res.projects ?? [])
          .map((p) => ({ projectId: p.projectId ?? '', minutes: toFiniteNumber(p.minutes) }))
          .filter((p) => p.projectId !== '' && p.minutes > 0),
      };
    },
  });
};

// The HTTP status of a failed request, when the client reports one.
export function requestErrorStatus(error: unknown): number | undefined {
  if (!error || typeof error !== 'object') return undefined;
  const candidate = error as {
    statusCode?: unknown;
    status?: unknown;
    response?: { status?: unknown };
  };
  for (const value of [candidate.statusCode, candidate.status, candidate.response?.status]) {
    if (typeof value === 'number') return value;
  }
  return undefined;
}

// A 403 (no Insights access) or 404 (missing or private project) will not
// change on a retry.
const retryUnlessDenied = (failureCount: number, error: unknown): boolean => {
  const status = requestErrorStatus(error);
  return status !== 403 && status !== 404 && failureCount < 2;
};

// The Work projects the viewer may open a project dashboard for: every public
// project and the private ones they are a member of, by name.
export const useDashboardProjects = (): UseQueryResult<DashboardProjectOption[]> => {
  return useQuery<DashboardProjectOption[]>({
    queryKey: ['insights-dashboard-projects'],
    retry: retryUnlessDenied,
    queryFn: async (): Promise<DashboardProjectOption[]> => {
      const res = await dashboardClient.listProjects({});
      return (res.projects ?? [])
        .map((p) => ({
          id: p.projectId ?? '',
          name: p.projectName ?? '',
          status: p.status ?? '',
          visibility: p.visibility ?? '',
        }))
        .filter((p) => p.id !== '');
    },
  });
};

// One Work project's dashboard. The server answers 404 for a private project
// the viewer is not a member of, and leaves invoices out when the viewer may
// not view finance invoices.
export const useProjectDashboard = (
  projectId: string
): UseQueryResult<ProjectDashboardResponse> => {
  return useQuery<ProjectDashboardResponse>({
    queryKey: ['insights-dashboard-project', projectId],
    enabled: projectId !== '',
    refetchInterval: 60000,
    retry: retryUnlessDenied,
    queryFn: async (): Promise<ProjectDashboardResponse> => {
      const res = await dashboardClient.getProjectDashboard({ projectId });
      const invoices = res.invoices;
      return {
        project: {
          id: res.project?.projectId ?? projectId,
          name: res.project?.projectName ?? '',
          status: res.project?.status ?? '',
          visibility: res.project?.visibility ?? '',
        },
        totalTasks: toFiniteNumber(res.progress?.totalTasks),
        completedTasks: toFiniteNumber(res.progress?.completedTasks),
        progressPct: toFiniteNumber(res.progress?.percent),
        overdueTaskCount: toFiniteNumber(res.overdueTaskCount),
        overdueTasks: (res.overdueTasks ?? [])
          .map((task) => ({
            id: task.taskId ?? '',
            title: task.title ?? '',
            dueDate: task.dueDate ?? '',
            assigneeId: task.assigneeId ?? '',
          }))
          .filter((task) => task.id !== ''),
        hoursThisMonth: {
          value: toFiniteNumber(res.hours?.thisMonth?.value),
          deltaPct: toFiniteNumber(res.hours?.thisMonth?.deltaPct),
        },
        priorMinutes: toFiniteNumber(res.hours?.priorMinutes),
        allTimeMinutes: toFiniteNumber(res.hours?.allTimeMinutes),
        invoices: invoices
          ? {
              invoiceCount: toFiniteNumber(invoices.invoiceCount),
              invoiced: toFiniteNumber(invoices.invoiced),
              paid: toFiniteNumber(invoices.paid),
              outstanding: toFiniteNumber(invoices.outstanding),
              overdue: toFiniteNumber(invoices.overdue),
              overdueCount: toFiniteNumber(invoices.overdueCount),
            }
          : undefined,
      };
    },
  });
};
