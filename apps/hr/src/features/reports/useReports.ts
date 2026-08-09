import { unwrap } from '@voxire/api';
import { useMutation } from '@tanstack/react-query';
import { hrApi } from '../../shared/services/hrApi';
import { reportPayloadSchema } from '../../shared/schemas/reports';
import type { ReportRequest, ReportPayload } from '../../shared/schemas/reports';

export function useGenerateReport(): ReturnType<
  typeof useMutation<ReportPayload, Error, ReportRequest>
> {
  return useMutation({
    mutationFn: async (request: ReportRequest): Promise<ReportPayload> => {
      const res = await unwrap(hrApi.generateReport(request));
      return reportPayloadSchema.parse(res);
    },
  });
}
