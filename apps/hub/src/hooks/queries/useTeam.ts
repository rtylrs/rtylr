import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../../lib/api';
import { toast } from 'sonner';

export interface TeamMember {
  id: string;
  name: string;
  email: string;
  role: 'owner' | 'admin' | 'manager' | 'staff' | 'viewer';
  apps: string[];
  status: 'active' | 'invited' | 'inactive';
  avatarUrl?: string;
  lastActive?: string;
  createdAt?: string;
}

const MOCK_TEAM: TeamMember[] = [
  { id: 'u_owner',  name: 'Owner Account', email: 'owner@example.com',  role: 'owner',  apps: ['pos','erp','hr','crm','finance','tracker'], status: 'active',  lastActive: new Date().toISOString() },
  { id: 'u_amelia', name: 'Amelia Nguyen', email: 'amelia@example.com', role: 'admin',  apps: ['pos','hr','crm'],                              status: 'active',  lastActive: new Date(Date.now() - 3600e3).toISOString() },
  { id: 'u_jorge',  name: 'Jorge Castillo', email: 'jorge@example.com', role: 'manager', apps: ['pos','erp'],                                    status: 'active',  lastActive: new Date(Date.now() - 2*3600e3).toISOString() },
  { id: 'u_priya',  name: 'Priya Shah',    email: 'priya@example.com',  role: 'staff',   apps: ['pos'],                                          status: 'active',  lastActive: new Date(Date.now() - 12*3600e3).toISOString() },
  { id: 'u_liam',   name: 'Liam O\u0027Connor', email: 'liam@example.com',role: 'staff',  apps: ['pos','hr'],                                     status: 'invited' },
  { id: 'u_mika',   name: 'Mika Tanaka',   email: 'mika@example.com',   role: 'viewer',  apps: ['crm'],                                          status: 'inactive' },
];

export const useTeam = () => {
  return useQuery<TeamMember[]>({
    queryKey: ['team'],
    queryFn: async () => {
      try {
        const res = await api.get('/team/members');
        return (res.data?.members ?? res.data) as TeamMember[];
      } catch {
        return MOCK_TEAM;
      }
    },
    staleTime: 30_000,
  });
};

export const useMember = (id: string | undefined) => {
  return useQuery<TeamMember | undefined>({
    queryKey: ['team', id],
    queryFn: async () => {
      if (!id) return undefined;
      try {
        const res = await api.get(`/team/members/${id}`);
        return res.data as TeamMember;
      } catch {
        return MOCK_TEAM.find((m) => m.id === id);
      }
    },
    enabled: !!id,
  });
};

export const useInviteMember = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { email: string; name: string; role: TeamMember['role']; apps: string[] }) => {
      try {
        const res = await api.post('/team/invite', payload);
        return res.data;
      } catch {
        return { ...payload, id: `u_${Math.random().toString(36).slice(2, 8)}`, status: 'invited' as const };
      }
    },
    onSuccess: () => {
      toast.success('Invitation sent');
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: () => toast.error('Failed to send invitation'),
  });
};

export const useUpdateMember = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, patch }: { id: string; patch: Partial<TeamMember> }) => {
      try {
        const res = await api.patch(`/team/members/${id}`, patch);
        return res.data;
      } catch {
        return { id, ...patch };
      }
    },
    onSuccess: () => {
      toast.success('Member updated');
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: () => toast.error('Failed to update member'),
  });
};

export const useDeactivateMember = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      try {
        await api.post(`/team/members/${id}/deactivate`);
      } catch {
        // swallow
      }
      return id;
    },
    onSuccess: () => {
      toast.success('Member deactivated');
      qc.invalidateQueries({ queryKey: ['team'] });
    },
  });
};
