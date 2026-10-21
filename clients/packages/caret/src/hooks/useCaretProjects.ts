import { useQuery } from "@tanstack/react-query";
import type { CaretProject } from "../types/chat";
import { createMemoryClient } from "../lib/memoryClient";

interface UseCaretProjectsParams {
  userId: number;
  apiBaseUrl: string;
}

export function useCaretProjects({
  userId,
  apiBaseUrl,
}: UseCaretProjectsParams) {
  const { data, isLoading, error, refetch } = useQuery<CaretProject[]>({
    queryKey: ["caret", "projects", userId],
    queryFn: async (): Promise<CaretProject[]> => {
      const response = await createMemoryClient(apiBaseUrl).listProjects({});
      return (response.projects ?? []).flatMap((project) => {
        if (!project.id || !project.name) return [];
        return [
          {
            id: project.id,
            name: project.name,
            icon: project.icon || "folder",
            color: project.color || "neutral",
            context: project.context || "",
            updatedAt: project.updatedAt || new Date(0).toISOString(),
          },
        ];
      });
    },
    enabled: userId > 0,
  });

  return { projects: data ?? [], isLoading, error: error ?? null, refetch };
}
