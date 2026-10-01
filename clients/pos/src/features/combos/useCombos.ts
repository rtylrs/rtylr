import { createOneKitClient } from "@voxire/api";
import {
  ComboServiceClient,
  type ComboServiceClientOptions,
} from "@voxire/voxie/pos/combo/v1/client";
import { useQuery } from "@tanstack/react-query";
import type { Combo } from "../../shared/types/pos";
import { resolveApiBaseUrl } from "../../shared/services/server/serverConfig";

class ServerScopedComboServiceClient extends ComboServiceClient {
  constructor(_configuredBaseUrl: string, options?: ComboServiceClientOptions) {
    super(resolveApiBaseUrl(), options);
  }
}

function comboItems(value: string): Combo["items"] {
  if (!value) return [];
  try {
    const parsed = JSON.parse(value) as unknown;
    if (!Array.isArray(parsed)) return [];
    return parsed.flatMap((item) => {
      if (typeof item === "number" || typeof item === "string") {
        const itemId = String(item);
        return itemId ? [{ itemId, qty: 1 }] : [];
      }
      if (!item || typeof item !== "object") return [];
      const record = item as Record<string, unknown>;
      const itemId = String(
        record.itemId ??
          record.item_id ??
          record.productId ??
          record.product_id ??
          "",
      );
      if (!itemId) return [];
      return [
        {
          itemId,
          qty:
            typeof record.qty === "number"
              ? record.qty
              : typeof record.quantity === "number"
                ? record.quantity
                : 1,
          name: typeof record.name === "string" ? record.name : undefined,
        },
      ];
    });
  } catch {
    return [];
  }
}

export function useCombos() {
  const query = useQuery({
    queryKey: ["pos-combos"],
    queryFn: async (): Promise<Combo[]> => {
      const comboClient = createOneKitClient(
        ServerScopedComboServiceClient,
        "pos",
      );
      const res = await comboClient.listCombos({});
      return (res.combos ?? [])
        .filter((combo) => combo.isActive !== false)
        .map((c) => ({
          id: c.id ?? "",
          name: c.name ?? "",
          price: c.price ?? 0,
          description: c.description ?? undefined,
          items: comboItems(c.items ?? ""),
        }));
    },
    staleTime: 5 * 60 * 1000,
  });

  return {
    combos: query.data ?? [],
    loading: query.isLoading,
    error: query.error
      ? query.error instanceof Error
        ? query.error.message
        : String(query.error)
      : null,
    reload: query.refetch,
  };
}
