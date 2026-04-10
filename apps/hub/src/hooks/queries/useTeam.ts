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

export const useTeam = () => {
  return useQuery<TeamMember[]>({
    queryKey: ['team'],
    queryFn: async () => {
      const res = await api.get('/team/members');
      return (res.data?.members ?? res.data) as TeamMember[];
    },
    staleTime: 30_000,
  });
};

export const useMember = (id: string | undefined) => {
  return useQuery<TeamMember | undefined>({
    queryKey: ['team', id],
    queryFn: async () => {
      if (!id) return undefined;
      const res = await api.get(`/team/members/${id}`);
      return res.data as TeamMember;
    },
    enabled: !!id,
  });
};

export const useInviteMember = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { email: string; fullName: string; role: TeamMember['role']; apps: string[] }) => {
      const res = await api.post('/team/invite', payload);
      return res.data;
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
      const res = await api.patch(`/team/members/${id}`, patch);
      return res.data;
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
      await api.post(`/team/members/${id}/deactivate`);
      return id;
    },
    onSuccess: () => {
      toast.success('Member deactivated');
      qc.invalidateQueries({ queryKey: ['team'] });
    },
  });
};
