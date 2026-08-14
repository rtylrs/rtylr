import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  decodeEmployeeDocumentPayload,
  type EmployeeDocumentPayload,
} from '@voxire/voxie/hr/document/v1/types';
import { safe } from '../../lib/api';
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

interface DocumentListWire {
  documents?: unknown;
  total?: unknown;
  limit?: unknown;
  offset?: unknown;
}

const decodeDocumentList = (value: unknown): ListEmployeeDocumentsResponse => {
  const body = (value && typeof value === 'object' ? value : {}) as DocumentListWire;
  const documents = Array.isArray(body.documents) ? body.documents : [];
  return {
    documents: documents.map((document) => decodeEmployeeDocumentPayload(document)),
    total: Number(body.total ?? documents.length),
    limit: Number(body.limit ?? 20),
    offset: Number(body.offset ?? 0),
  };
};

export const useListEmployeeDocuments = (employeeId?: number) => {
  return useQuery<ListEmployeeDocumentsResponse>({
    queryKey: ['employee-documents', employeeId],
    queryFn: async (): Promise<ListEmployeeDocumentsResponse> => {
      const res = await unwrap(
        safe.get<unknown>('/hr/documents', {
          params: employeeId ? { employee_id: employeeId } : undefined,
        })
      );
      return decodeDocumentList(res);
    },
    enabled: !!employeeId,
  });
};

export const useUploadDocument = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: UploadDocumentPayload): Promise<UploadDocumentResponse> => {
      const res = await unwrap(
        safe.post<unknown>('/hr/documents', {
          employee_id: data.employeeId,
          document_type: data.documentType,
          title: data.title,
          storage_mode: data.storageMode,
          file_url: data.fileUrl,
          external_url: data.externalUrl,
          expiry_date: data.expiryDate,
        })
      );
      const body = (res && typeof res === 'object' ? res : {}) as { document?: unknown };
      return { document: decodeEmployeeDocumentPayload(body.document) };
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
