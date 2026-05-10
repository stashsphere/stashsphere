import { Axios } from 'axios';

export type ExportStatus = {
  id: string;
  status: string;
  createdAt: string;
  expiresAt: string | null;
  error: string | null;
};

export const triggerExport = async (axios: Axios): Promise<ExportStatus> => {
  const response = await axios.post(
    '/export',
    {},
    {
      headers: { 'Content-Type': 'application/json' },
    }
  );
  return response.data as ExportStatus;
};

export const getExportStatus = async (axios: Axios): Promise<ExportStatus | null> => {
  const response = await axios.get('/export', {
    headers: { 'Content-Type': 'application/json' },
    validateStatus: (status) => status === 200 || status === 404,
  });
  if (response.status === 404) return null;
  return response.data as ExportStatus;
};
