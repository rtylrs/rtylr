import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { bookService } from '../services/bookService';

export function useBookSettings() {
  return useQuery({
    queryKey: ['book-settings'],
    queryFn: async () => {
      const res = await bookService.getSettings();
      return res.data?.data ?? res.data ?? null;
    },
  });
}

export function useUpdateBookSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => bookService.updateSettings(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['book-settings'] }),
  });
}
