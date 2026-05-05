import { useContext, useEffect, useRef, useState } from 'react';
import { AxiosContext } from '../../context/axios';
import { PrimaryButton, SecondaryButton } from '../../components/shared';
import { triggerExport, getExportStatus, downloadExport, ExportStatus } from '../../api/export';

const POLL_INTERVAL_MS = 3000;

export const ImportExport = () => {
  const axiosInstance = useContext(AxiosContext);
  const [exportStatus, setExportStatus] = useState<ExportStatus | null | undefined>(undefined);
  const [initialLoading, setInitialLoading] = useState(true);
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
      const status = await getExportStatus(axiosInstance);
      setExportStatus(status);
      if (status?.status !== 'pending') {
        stopPolling();
      }
    } catch {
      stopPolling();
    }
  };

  useEffect(() => {
    fetchStatus().finally(() => setInitialLoading(false));
    return () => stopPolling();
  }, [axiosInstance]);

  useEffect(() => {
    if (exportStatus?.status === 'pending' && pollRef.current === null) {
      pollRef.current = setInterval(fetchStatus, POLL_INTERVAL_MS);
    }
    if (exportStatus?.status !== 'pending') {
      stopPolling();
    }
  }, [exportStatus?.status]);

  const handleTriggerExport = async () => {
    if (axiosInstance === null) return;
    setTriggering(true);
    setError(undefined);
    try {
      const status = await triggerExport(axiosInstance);
      setExportStatus(status);
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

  const formatDate = (iso: string) =>
    new Date(iso).toLocaleString(undefined, {
      dateStyle: 'medium',
      timeStyle: 'short',
    });

  const renderExportBody = () => {
    if (initialLoading) {
      return <p className="text-display">Loading…</p>;
    }

    if (exportStatus === null) {
      return (
        <div>
          <p className="text-display mb-4">No export has been created yet.</p>
          <PrimaryButton onClick={handleTriggerExport} disabled={triggering}>
            {triggering ? 'Starting…' : 'Create Export'}
          </PrimaryButton>
        </div>
      );
    }

    if (exportStatus?.status === 'pending') {
      return (
        <div className="bg-neutral-900 border border-neutral-700 rounded-sm p-4">
          <p className="text-display">Export in progress… This page will update automatically.</p>
        </div>
      );
    }

    if (exportStatus?.status === 'done') {
      return (
        <div>
          <div className="bg-success-900 border border-success-700 rounded-sm p-4 mb-4">
            <p className="text-success-200 font-semibold">Export ready</p>
            <p className="text-success-200 text-sm mt-1">
              Created: {formatDate(exportStatus.createdAt)}
            </p>
            {exportStatus.expiresAt && (
              <p className="text-success-200 text-sm">
                Expires: {formatDate(exportStatus.expiresAt)}
              </p>
            )}
          </div>
          <div className="flex gap-3">
            <PrimaryButton
              onClick={async () => {
                if (!axiosInstance) return;
                const { blob, filename } = await downloadExport(axiosInstance);
                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = filename;
                a.click();
                URL.revokeObjectURL(url);
              }}
            >
              Download
            </PrimaryButton>
            <SecondaryButton onClick={handleTriggerExport} disabled={triggering}>
              {triggering ? 'Starting…' : 'Create New Export'}
            </SecondaryButton>
          </div>
        </div>
      );
    }

    if (exportStatus?.status === 'error') {
      return (
        <div>
          <div className="bg-danger-900 border border-danger-700 rounded-sm p-4 mb-4">
            <p className="text-danger-200 font-semibold">Export failed</p>
            {exportStatus.error && (
              <p className="text-danger-200 text-sm mt-1">{exportStatus.error}</p>
            )}
          </div>
          <PrimaryButton onClick={handleTriggerExport} disabled={triggering}>
            {triggering ? 'Starting…' : 'Create New Export'}
          </PrimaryButton>
        </div>
      );
    }

    return null;
  };

  return (
    <div className="max-w-2xl mx-auto px-4">
      <div className="mt-8">
        <h2 className="text-xl font-semibold text-display mb-2">Export</h2>
        <p className="text-display text-sm mb-4">
          Export all your things, lists, and images as a ZIP file.
        </p>
        {error && <p className="text-warning mb-4">{error}</p>}
        {renderExportBody()}
      </div>

      <div className="mt-8 pt-8 border-t border-gray-200">
        <h2 className="text-xl font-semibold text-display mb-2">Import</h2>
        <p className="text-display text-sm mb-4">
          Import a previously exported ZIP file to restore your data.
        </p>
        <div className="flex items-center gap-3">
          <input type="file" accept=".zip" disabled className="text-display text-sm" />
          <PrimaryButton disabled>Import</PrimaryButton>
        </div>
        <p className="text-warning text-sm mt-2">Import is not yet available.</p>
      </div>
    </div>
  );
};
