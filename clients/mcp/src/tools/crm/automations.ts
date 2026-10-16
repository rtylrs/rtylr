import type { Tool } from '../types.js';
import { ok, apiErr } from '../types.js';

export const listAutomations: Tool = {
  name: 'crm_list_automations',
  description: 'List CRM automations (triggered workflows like welcome emails, re-engagement sequences).',
  inputSchema: { type: 'object', properties: {} },
  handler: async (_input, api) => {
    try {
      const res = await api.get('/crm/automations');
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const toggleAutomation: Tool = {
  name: 'crm_toggle_automation',
  description: 'Enable or disable an automation.',
  inputSchema: {
    type: 'object',
    properties: {
      automation_id: { type: 'string', description: 'Automation ID' },
      enabled: { type: 'string', description: 'true to enable, false to disable' },
    },
    required: ['automation_id', 'enabled'],
  },
  handler: async (input, api) => {
    try {
      const enabled = String(input.enabled) === 'true';
      const idStr = String(input.automation_id);
      const listRes = await api.get('/crm/automations');
      const automations: Array<Record<string, unknown>> = listRes.data?.automations ?? [];
      const existing = automations.find((a) => String(a.id) === idStr);
      if (!existing) {
        return apiErr(new Error(`Automation ${idStr} not found`));
      }
      const res = await api.put(`/crm/automations/${encodeURIComponent(idStr)}`, {
        name: existing.name,
        trigger_type: existing.trigger_type,
        trigger_config: existing.trigger_config ?? {},
        actions: existing.actions ?? [],
        is_active: enabled,
      });
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
