import { useContext, useEffect, useRef, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { PrimaryButton, SecondaryButton } from '../../components/shared';
import { triggerExport, getExportStatus, downloadExport, ExportStatus } from '../../api/export';
import { queueImport, getImportStatus, ImportStatus } from '../../api/import';

const POLL_INTERVAL_MS = 3000;

const ExportSection = () => {
  const axiosInstance = useContext(AxiosContext);
  const [status, setStatus] = useState<ExportStatus | null | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [triggering, setTriggering] = useState(false);
  const [error, setError] = useState<string | undefined>(undefined);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const stopPolling = () => {
    if (pollRef.current !== null) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
  };

  const fetchStatus = async () => {
    if (axiosInstance === null) return;
    try {
      const s = await getExportStatus(axiosInstance);
      setStatus(s);
      if (s?.status !== 'pending' && s?.status !== 'processing') stopPolling();
    } catch {
      stopPolling();
    }
  };

  useEffect(() => {
    fetchStatus().finally(() => setLoading(false));
    return () => stopPolling();
  }, [axiosInstance]);

  useEffect(() => {
    const active = status?.status === 'pending' || status?.status === 'processing';
    if (active && pollRef.current === null) {
      pollRef.current = setInterval(fetchStatus, POLL_INTERVAL_MS);
    }
    if (!active) stopPolling();
  }, [status?.status]);

  const handleTrigger = async () => {
    if (axiosInstance === null) return;
    setTriggering(true);
    setError(undefined);
    try {
      setStatus(await triggerExport(axiosInstance));
    } catch (e: unknown) {
      const err = e as { response?: { status?: number } };
      if (err?.response?.status === 409) {
        setError('An export is already in progress.');
        await fetchStatus();
      } else {
        setError('Failed to start export. Please try again.');
      }
    } finally {
      setTriggering(false);
    }
  };

  const handleDownload = async () => {
    if (axiosInstance === null) return;
    const { blob, filename } = await downloadExport(axiosInstance);
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
  };

  const formatDate = (iso: string) =>
    new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });

  if (loading) {
    return <p className="text-display">Loading…</p>;
  }

  if (status === null) {
    return (
      <div>
        {error && <p className="text-warning mb-4">{error}</p>}
        <p className="text-display mb-4">No export has been created yet.</p>
        <PrimaryButton onClick={handleTrigger} disabled={triggering}>
          {triggering ? 'Starting…' : 'Create Export'}
        </PrimaryButton>
      </div>
    );
  }

  if (status?.status === 'pending') {
    return (
      <div className="bg-neutral-900 border border-neutral-700 rounded-sm p-4">
        <p className="text-display">Export in progress… This page will update automatically.</p>
      </div>
    );
  }

  if (status?.status === 'done') {
    return (
      <div>
        <div className="bg-success-900 border border-success-700 rounded-sm p-4 mb-4">
          <p className="text-success-200 font-semibold">Export ready</p>
          <p className="text-success-200 text-sm mt-1">Created: {formatDate(status.createdAt)}</p>
          {status.expiresAt && (
            <p className="text-success-200 text-sm">Expires: {formatDate(status.expiresAt)}</p>
          )}
        </div>
        <div className="flex gap-3">
          <PrimaryButton onClick={handleDownload}>Download</PrimaryButton>
          <SecondaryButton onClick={handleTrigger} disabled={triggering}>
            {triggering ? 'Starting…' : 'Create New Export'}
          </SecondaryButton>
        </div>
      </div>
    );
  }

  if (status?.status === 'error') {
    return (
      <div>
        <div className="bg-danger-900 border border-danger-700 rounded-sm p-4 mb-4">
          <p className="text-danger-200 font-semibold">Export failed</p>
          {status.error && <p className="text-danger-200 text-sm mt-1">{status.error}</p>}
        </div>
        {error && <p className="text-warning mb-4">{error}</p>}
        <PrimaryButton onClick={handleTrigger} disabled={triggering}>
          {triggering ? 'Starting…' : 'Create New Export'}
        </PrimaryButton>
      </div>
    );
  }

  return null;
};

const ImportSection = () => {
  const axiosInstance = useContext(AxiosContext);
  const [status, setStatus] = useState<ImportStatus | null | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [file, setFile] = useState<File | null>(null);
  const [importing, setImporting] = useState(false);
  const [error, setError] = useState<string | undefined>(undefined);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const stopPolling = () => {
    if (pollRef.current !== null) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
  };

  const fetchStatus = async () => {
    if (axiosInstance === null) return;
    try {
      const s = await getImportStatus(axiosInstance);
      setStatus(s);
      if (s?.status !== 'pending' && s?.status !== 'processing') stopPolling();
    } catch {
      stopPolling();
    }
  };

  useEffect(() => {
    fetchStatus().finally(() => setLoading(false));
    return () => stopPolling();
  }, [axiosInstance]);

  useEffect(() => {
    const active = status?.status === 'pending' || status?.status === 'processing';
    if (active && pollRef.current === null) {
      pollRef.current = setInterval(fetchStatus, POLL_INTERVAL_MS);
    }
    if (!active) stopPolling();
  }, [status?.status]);

  const handleImport = async () => {
    if (axiosInstance === null || file === null) return;
    setImporting(true);
    setError(undefined);
    try {
      setStatus(await queueImport(axiosInstance, file));
      setFile(null);
      if (fileInputRef.current) fileInputRef.current.value = '';
    } catch (e: unknown) {
      const err = e as { response?: { status?: number; data?: { message?: string } } };
      if (err?.response?.status === 409) {
        setError('An import is already in progress.');
        await fetchStatus();
      } else if (err?.response?.status === 413) {
        setError('File is too large.');
      } else if (err?.response?.status === 400) {
        setError(err?.response?.data?.message ?? 'Invalid import file.');
      } else {
        setError('Failed to start import. Please try again.');
      }
    } finally {
      setImporting(false);
    }
  };

  if (loading) {
    return <p className="text-display">Loading…</p>;
  }

  const active = status?.status === 'pending' || status?.status === 'processing';

  return (
    <div>
      {status?.status === 'done' && (
        <div className="bg-success-900 border border-success-700 rounded-sm p-4 mb-4">
          <p className="text-success-200 font-semibold">Import complete</p>
          <p className="text-success-200 text-sm mt-1">
            {status.thingsImported ?? 0} things, {status.listsImported ?? 0} lists,{' '}
            {status.imagesImported ?? 0} images imported.
          </p>
        </div>
      )}
      {status?.status === 'error' && (
        <div className="bg-danger-900 border border-danger-700 rounded-sm p-4 mb-4">
          <p className="text-danger-200 font-semibold">Import failed</p>
          {status.error && <p className="text-danger-200 text-sm mt-1">{status.error}</p>}
        </div>
      )}
      {active && (
        <div className="bg-neutral-900 border border-neutral-700 rounded-sm p-4 mb-4">
          <p className="text-display">Import in progress… This page will update automatically.</p>
        </div>
      )}
      {!active && (
        <div className="flex items-center gap-3">
          <input
            ref={fileInputRef}
            type="file"
            accept=".zip"
            disabled={importing}
            onChange={(e) => setFile(e.currentTarget.files?.[0] ?? null)}
            className="text-display text-sm"
          />
          <PrimaryButton onClick={handleImport} disabled={file === null || importing}>
            {importing ? 'Uploading…' : 'Import'}
          </PrimaryButton>
        </div>
      )}
      {error && <p className="text-warning text-sm mt-2">{error}</p>}
    </div>
  );
};

export const ImportExport = () => (
  <div className="max-w-2xl mx-auto px-4">
    <div className="mt-8">
      <h2 className="text-xl font-semibold text-display mb-2">Export</h2>
      <p className="text-display text-sm mb-4">
        Export all your things, lists, and images as a ZIP file.
      </p>
      <ExportSection />
    </div>

    <div className="mt-8 pt-8 border-t border-gray-200">
      <h2 className="text-xl font-semibold text-display mb-2">Import</h2>
      <p className="text-display text-sm mb-4">
        Import a previously exported ZIP file to restore your data.
      </p>
      <ImportSection />
    </div>
  </div>
);
