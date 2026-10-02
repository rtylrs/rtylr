import { useUserStore } from '../../../../shared/stores/userStore';

export function useFlowBusinessName() {
  const name = useUserStore((state) => state.user?.businessName);
  return { data: name ?? null };
}

export function useFlowBusinessId() {
  const id = useUserStore((state) => state.user?.businessId);
  return { data: id === undefined ? null : String(id) };
}
