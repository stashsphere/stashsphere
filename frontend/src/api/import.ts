import { Axios } from 'axios';

export type ImportStatus = {
  id: string;
  status: string;
  createdAt: string;
  completedAt: string | null;
  error: string | null;
  thingsImported: number | null;
  listsImported: number | null;
  imagesImported: number | null;
};

export const queueImport = async (axios: Axios, file: File): Promise<ImportStatus> => {
  const formData = new FormData();
  formData.append('file', file);
  const response = await axios.post('/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return response.data as ImportStatus;
};

export const getImportStatus = async (axios: Axios): Promise<ImportStatus | null> => {
  const response = await axios.get('/import', {
    validateStatus: (status) => status === 200 || status === 404,
  });
  if (response.status === 404) return null;
  return response.data as ImportStatus;
};
