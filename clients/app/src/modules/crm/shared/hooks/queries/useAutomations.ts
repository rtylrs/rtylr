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
import { crmApi } from '../../services/crmApi';
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

function toBackendRequest(data: Partial<Automation>): BackendAutomationRequest {
  const rawActions = Array.isArray(data.actions) ? data.actions : (data.nodes ?? []).slice(1);
  const actions = rawActions.map((action) => {
    const value =
      action && typeof action === 'object' ? (action as Record<string, unknown>) : { type: action };
    const rawType = value.automationType ?? value.type ?? value.label;
    const canonicalType = actionTypeFor(rawType);
    const {
      id: _id,
      label: _label,
      x: _x,
      y: _y,
      automationType: _automationType,
      ...config
    } = value;
    return {
      ...config,
      type: String(canonicalType ?? ''),
    };
  });
  const triggerType =
    data.triggerType ?? data.nodes?.[0]?.automationType ?? triggerTypeFor(data.trigger);
  const triggerConfig =
    typeof data.triggerConfig === 'string'
      ? z.record(z.string(), z.unknown()).parse(JSON.parse(data.triggerConfig))
      : (data.triggerConfig ?? {});
  return {
    name: data.name ?? '',
    triggerType,
    triggerConfig,
    actions,
    isActive: data.isActive ?? data.status === 'active',
  };
}

export const useListAutomations = (): UseQueryResult<Automation[]> =>
  useQuery<Automation[]>({
    queryKey: ['crm-automations'],
    queryFn: async (): Promise<Automation[]> => {
      const res = await unwrap(crmApi.getAutomations());
      return z.array(automationSchema).parse(unwrapList(res, 'automations'));
    },
  });

export const useGetAutomation = (id: number | undefined): UseQueryResult<Automation> =>
  useQuery<Automation>({
    queryKey: ['crm-automation', id],
    queryFn: async (): Promise<Automation> => {
      const res = await unwrap(crmApi.getAutomation(id ?? 0));
      const envelope = (res ?? {}) as { automation?: unknown };
      return automationSchema.parse(envelope.automation ?? res);
    },
    enabled: !!id,
  });

export const useCreateAutomation = (): UseMutationResult<
  Automation,
  unknown,
  Partial<Automation>
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Automation, unknown, Partial<Automation>>({
    mutationFn: async (data: Partial<Automation>): Promise<Automation> => {
      const res = await unwrap(crmApi.createAutomation(toBackendRequest(data)));
      return res as Automation;
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-automations'] });
      toast.success(t('crm.automations.created'));
    },
    onError: (): void => {
      toast.error(t('crmErrors.generic'));
    },
  });
};

export const useUpdateAutomation = (): UseMutationResult<
  Automation,
  unknown,
  UpdateAutomationPayload
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Automation, unknown, UpdateAutomationPayload>({
    mutationFn: async (payload: UpdateAutomationPayload): Promise<Automation> => {
      const cached = qc
        .getQueryData<Automation[]>(['crm-automations'])
        ?.find((item) => item.id === payload.id);
      const res = await unwrap(
        crmApi.updateAutomation(payload.id, toBackendRequest({ ...cached, ...payload.data }))
      );
      return res as Automation;
    },
    onSuccess: (_data: Automation, payload: UpdateAutomationPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-automations'] });
      void qc.invalidateQueries({ queryKey: ['crm-automation', payload.id] });
      toast.success(t('crm.automations.updated'));
    },
    onError: (): void => {
      toast.error(t('crmErrors.generic'));
    },
  });
};

export const useDeleteAutomation = (): UseMutationResult<unknown, unknown, number> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, number>({
    mutationFn: (id: number): Promise<unknown> => unwrap(crmApi.deleteAutomation(id)),
    onSuccess: (_data: unknown, id: number): void => {
      void qc.invalidateQueries({ queryKey: ['crm-automations'] });
      void qc.invalidateQueries({ queryKey: ['crm-automation', id] });
      toast.success(t('crm.automations.deleted'));
    },
    onError: (): void => {
      toast.error(t('crmErrors.generic'));
    },
  });
};

export type { Automation } from '../../schemas/automations';
