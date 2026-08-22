import { unwrap } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { z } from 'zod';
import { safe } from '../../lib/api';
import { unwrapList } from '../../lib/unwrapList';
import { automationSchema } from '../../schemas/automations';
import type { Automation } from '../../schemas/automations';

interface UpdateAutomationPayload {
  id: number;
  data: Partial<Automation>;
}

interface BackendAutomationRequest {
  name: string;
  triggerType: string;
  triggerConfig: Record<string, unknown>;
  actions: unknown[];
  isActive: boolean;
}

export interface AutomationMutationResponse {
  automationId?: number;
  automation_id?: number;
  message?: string;
}

export function triggerTypeFor(value?: string): string {
  const map: Record<string, string> = {
    'New customer': 'customer_signup',
    Birthday: 'date_birthday',
    'Inactive 60 days': 'inactivity',
    'Inactive customer': 'inactivity',
    'Order created': 'order_created',
    'Post visit': 'order_created',
    'Tier upgrade': 'loyalty_tier_up',
    'Tier downgrade': 'loyalty_tier_down',
    'Tag added': 'tag_added',
    'Points threshold': 'points_threshold',
  };
  return map[value ?? ''] ?? value ?? 'customer_signup';
}

export function actionTypeFor(value: unknown): unknown {
  if (typeof value !== 'string') return value;
  const map: Record<string, string> = {
    'Send email': 'send_email',
    'Send SMS': 'send_sms',
    'Award points': 'award_points',
    'Add tag': 'tag_add',
  };
  return map[value] ?? value;
}

export function toBackendRequest(data: Partial<Automation>): BackendAutomationRequest {
  const actions = (
    Array.isArray(data.actions)
      ? data.actions
      : (data.nodes ?? []).slice(1).map((node) => ({ type: node.label }))
  ).map((action) =>
    typeof action === 'object' && action !== null
      ? {
          ...(action as Record<string, unknown>),
          type: actionTypeFor((action as Record<string, unknown>).type),
        }
      : actionTypeFor(action)
  );
  const triggerType = data.triggerType ?? triggerTypeFor(data.trigger);
  const triggerConfig =
    data.triggerConfig && typeof data.triggerConfig === 'object'
      ? (data.triggerConfig as Record<string, unknown>)
      : {};
  return {
    name: data.name ?? '',
    triggerType,
    triggerConfig,
    actions,
    isActive: data.isActive ?? data.status === 'active',
  };
}

export function mergeAutomationData(
  current: Automation | undefined,
  data: Partial<Automation>
): Partial<Automation> {
  return { ...current, ...data };
}

export const useListAutomations = (): UseQueryResult<Automation[]> =>
  useQuery<Automation[]>({
    queryKey: ['crm-automations'],
    queryFn: async (): Promise<Automation[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/automations'));
      return z.array(automationSchema).parse(unwrapList(res, 'automations'));
    },
  });

export const useGetAutomation = (id: number | undefined): UseQueryResult<Automation> => {
  const qc = useQueryClient();
  return useQuery<Automation>({
    queryKey: ['crm-automation', id],
    queryFn: async (): Promise<Automation> => {
      if (id == null || id <= 0) throw new Error('Invalid automation id');
      const cached = qc
        .getQueryData<Automation[]>(['crm-automations'])
        ?.find((automation) => automation.id === id);
      if (cached) return cached;

      // The automations API exposes list/create/update/delete routes, but no
      // dedicated GET /{id} route. Reuse the list endpoint for deep links.
      const res = await unwrap(safe.get<unknown>('/crm/automations'));
      const automations = z.array(automationSchema).parse(unwrapList(res, 'automations'));
      const automation = automations.find((item) => item.id === id);
      if (!automation) throw new Error('Automation not found');
      return automation;
    },
    enabled: id != null && Number.isInteger(id) && id > 0,
  });
};

export const useCreateAutomation = (): UseMutationResult<
  AutomationMutationResponse,
  unknown,
  Partial<Automation>
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<AutomationMutationResponse, unknown, Partial<Automation>>({
    mutationFn: async (data: Partial<Automation>): Promise<AutomationMutationResponse> => {
      const res = await unwrap(safe.post<unknown>('/crm/automations', toBackendRequest(data)));
      const raw = res && typeof res === 'object' ? (res as Record<string, unknown>) : {};
      const automationId = Number(raw.automationId ?? raw.automation_id);
      return {
        automationId: Number.isFinite(automationId) && automationId > 0 ? automationId : undefined,
        message: typeof raw.message === 'string' ? raw.message : undefined,
      };
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-automations'] });
      toast.success(t('crm.automations.created'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdateAutomation = (): UseMutationResult<
  AutomationMutationResponse,
  unknown,
  UpdateAutomationPayload
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<AutomationMutationResponse, unknown, UpdateAutomationPayload>({
    mutationFn: async (payload: UpdateAutomationPayload): Promise<AutomationMutationResponse> => {
      const cached =
        qc.getQueryData<Automation>(['crm-automation', payload.id]) ??
        qc.getQueryData<Automation[]>(['crm-automations'])?.find((item) => item.id === payload.id);
      const res = await unwrap(
        safe.put<unknown>(
          `/crm/automations/${payload.id}`,
          toBackendRequest(mergeAutomationData(cached, payload.data))
        )
      );
      const raw = res && typeof res === 'object' ? (res as Record<string, unknown>) : {};
      return { message: typeof raw.message === 'string' ? raw.message : undefined };
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-automations'] });
      toast.success(t('crm.automations.updated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useDeleteAutomation = (): UseMutationResult<unknown, unknown, number> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, number>({
    mutationFn: (id: number): Promise<unknown> => unwrap(safe.delete(`/crm/automations/${id}`)),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-automations'] });
      toast.success(t('crm.automations.deleted'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export type { Automation } from '../../schemas/automations';
