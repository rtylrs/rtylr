import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { EmployeeDocumentPayload } from '@voxire/voxie/hr/document/v1/types';
import { hrTransport } from '../../lib/clients';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

interface ListEmployeeDocumentsResponse {
  documents: EmployeeDocumentPayload[];
  total: number;
  limit: number;
  offset: number;
}

interface UploadDocumentResponse {
  document: EmployeeDocumentPayload;
}

export interface UploadDocumentPayload {
  employeeId: number;
  documentType: string;
  title: string;
  storageMode: string;
  fileUrl: string;
  externalUrl?: string;
  expiryDate?: string;
}

export const useListEmployeeDocuments = (employeeId?: number) => {
  return useQuery<ListEmployeeDocumentsResponse>({
    queryKey: ['hr', 'employee-documents', employeeId],
    queryFn: async (): Promise<ListEmployeeDocumentsResponse> => {
      const res = await unwrap(
        hrTransport.get<ListEmployeeDocumentsResponse>('/hr/documents', {
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
  const { t } = useTranslation('hrCommon');

  return useMutation({
    mutationFn: async (data: UploadDocumentPayload): Promise<UploadDocumentResponse> => {
      const res = await unwrap(hrTransport.post<UploadDocumentResponse>('/hr/documents', data));
      return res as UploadDocumentResponse;
    },
    onSuccess: (_data: UploadDocumentResponse, variables: UploadDocumentPayload) => {
      queryClient.invalidateQueries({
        queryKey: ['hr', 'employee-documents', variables.employeeId],
      });
      queryClient.invalidateQueries({ queryKey: ['hr', 'employee-documents'] });
      toast.success(t('hrDocuments.uploaded'));
    },
    onError: () => {
      toast.error(t('hrErrors.generic'));
    },
  });
};
