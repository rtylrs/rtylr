import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createMemoryClient } from "../lib/memoryClient";

export interface CaretMemoryItem {
  id: string;
  value: string;
  updatedAt?: string;
  /** Distilled from the person's feedback on replies, not told directly. */
  learned?: boolean;
}

interface UseCaretMemoriesParams {
  userId: number;
  apiBaseUrl: string;
}

/** What Caret remembers for the signed-in user, with edit and delete. */
export function useCaretMemories({
  userId,
  apiBaseUrl,
}: UseCaretMemoriesParams) {
  const queryClient = useQueryClient();
  const queryKey = ["caret", "memories", userId];
  const client = () => createMemoryClient(apiBaseUrl);

  const query = useQuery({
    queryKey,
    enabled: userId > 0,
    queryFn: async (): Promise<CaretMemoryItem[]> => {
      const res = await client().listMemories({});
      return (res.memories ?? [])
        .filter((memory) => memory.id && memory.value)
        .map((memory) => ({
          id: memory.id as string,
          value: memory.value as string,
          updatedAt: memory.updatedAt,
          learned: memory.type === "lesson",
        }));
    },
  });

  const refresh = () => queryClient.invalidateQueries({ queryKey });

  const save = useMutation({
    mutationFn: ({ id, value }: { id?: string; value: string }) =>
      client().saveMemory({ id: id ?? "", value }),
    onSettled: refresh,
  });

  const remove = useMutation({
    mutationFn: (id: string) => client().deleteMemory({ id }),
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<CaretMemoryItem[]>(queryKey);
      queryClient.setQueryData<CaretMemoryItem[]>(queryKey, (current) =>
        current?.filter((memory) => memory.id !== id),
      );
      return { previous };
    },
    onError: (_error, _id, context) => {
      if (context?.previous)
        queryClient.setQueryData(queryKey, context.previous);
    },
    onSettled: refresh,
  });

  return {
    memories: query.data ?? [],
    isLoading: query.isLoading,
    error: query.error,
    save: (value: string, id?: string) => save.mutateAsync({ id, value }),
    remove: (id: string) => remove.mutateAsync(id),
    isSaving: save.isPending,
  };
}
