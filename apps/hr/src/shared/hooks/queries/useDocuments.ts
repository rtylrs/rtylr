import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ListEmployeeDocumentsResponse } from '@voxire/contracts/hr/document/v1/list_employee_documents';
import type { UploadDocumentResponse } from '@voxire/contracts/hr/document/v1/upload_document';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

export interface UploadDocumentPayload {
  employeeId: number;
  documentType: string;
  title: string;
  storageMode: string;
  fileUrl: string;
  externalUrl?: string;
}

export const useListEmployeeDocuments = (employeeId?: number) => {
  return useQuery<ListEmployeeDocumentsResponse>({
    queryKey: ['employee-documents', employeeId],
    queryFn: async (): Promise<ListEmployeeDocumentsResponse> => {
      const res = await unwrap(
        safe.get<ListEmployeeDocumentsResponse>('/hr/documents', {
          params: employeeId ? { employeeId } : undefined,
        })
      );
      return res as ListEmployeeDocumentsResponse;
    },
    enabled: !!employeeId,
  });
};

export const useUploadDocument = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: UploadDocumentPayload): Promise<UploadDocumentResponse> => {
      const res = await unwrap(safe.post<UploadDocumentResponse>('/hr/documents', data));
      return res as UploadDocumentResponse;
    },
    onSuccess: (_data: UploadDocumentResponse, variables: UploadDocumentPayload) => {
      queryClient.invalidateQueries({ queryKey: ['employee-documents', variables.employeeId] });
      queryClient.invalidateQueries({ queryKey: ['employee-documents'] });
      toast.success(t('documents.uploaded'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
