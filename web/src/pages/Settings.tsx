import { useState, useRef } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api, apiUpload } from '../lib/api';
import type { WorkspaceStatus } from '../lib/types';

const ADMIN_PATH = '/admin/workspace';

export default function Settings() {
  const qc = useQueryClient();
  const status = useQuery({
    queryKey: ['ws-status'],
    queryFn: () => api<WorkspaceStatus>(`${ADMIN_PATH}/status`),
  });

  const [delegatedAdmin, setDelegatedAdmin] = useState('');
  const [customerID, setCustomerID] = useState('my_customer');
  const [file, setFile] = useState<File | null>(null);
  const [feedback, setFeedback] = useState<{ kind: 'ok' | 'err'; msg: string } | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const upload = useMutation({
    mutationFn: async () => {
      if (!file) throw new Error('Please choose your service-account JSON file.');
      if (!delegatedAdmin) throw new Error('Delegated admin email is required.');
      const fd = new FormData();
      fd.append('file', file);
      fd.append('delegated_admin', delegatedAdmin);
      fd.append('customer_id', customerID || 'my_customer');
      return apiUpload<WorkspaceStatus>(`${ADMIN_PATH}/credentials`, fd);
    },
    onSuccess: () => {
      setFeedback({ kind: 'ok', msg: 'Credentials saved. Workspace endpoints are live.' });
      setFile(null);
      if (fileInputRef.current) fileInputRef.current.value = '';
      qc.invalidateQueries({ queryKey: ['ws-status'] });
    },
    onError: (err: Error) => setFeedback({ kind: 'err', msg: err.message }),
  });

  const test = useMutation({
    mutationFn: () => api<{ ok: boolean }>(`${ADMIN_PATH}/test`, { method: 'POST' }),
    onSuccess: () => setFeedback({ kind: 'ok', msg: 'Connection test passed — Workspace API is reachable.' }),
    onError: (err: Error) => setFeedback({ kind: 'err', msg: err.message }),
  });

  const remove = useMutation({
    mutationFn: () => api<{ deleted: boolean }>(`${ADMIN_PATH}/credentials`, { method: 'DELETE' }),
    onSuccess: () => {
      setFeedback({ kind: 'ok', msg: 'Credentials removed.' });
      qc.invalidateQueries({ queryKey: ['ws-status'] });
    },
    onError: (err: Error) => setFeedback({ kind: 'err', msg: err.message }),
  });

  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-semibold">Settings — Google Workspace integration</h1>
        <p className="text-gray-600 mt-1">
          Upload the service-account JSON you downloaded from Google Cloud Console. Goldy stores it in the
          database and uses it to authenticate against the Admin SDK.
        </p>
      </header>

      {/* Status */}
      <section
        className={`border rounded-lg p-4 ${
          status.data?.configured
            ? 'border-green-300 bg-green-50'
            : 'border-amber-300 bg-amber-50'
        }`}
      >
        {status.isLoading && <div className="text-sm text-gray-600">Loading status…</div>}
        {status.isError && (
          <div className="text-sm text-red-700">Failed to load status: {(status.error as Error).message}</div>
        )}
        {status.data && (
          <div className="text-sm">
            {status.data.configured ? (
              <>
                <div className="font-medium text-green-900">
                  Workspace integration is active{' '}
                  <span className="font-normal text-green-700">
                    (loaded from {status.data.source === 'db' ? 'database' : 'environment variables'})
                  </span>
                </div>
                <dl className="mt-2 grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-1 text-gray-700">
                  {status.data.delegated_admin && (
                    <div><dt className="inline font-medium">Delegated admin: </dt><dd className="inline font-mono text-xs">{status.data.delegated_admin}</dd></div>
                  )}
                  {status.data.customer_id && (
                    <div><dt className="inline font-medium">Customer ID: </dt><dd className="inline font-mono text-xs">{status.data.customer_id}</dd></div>
                  )}
                  {status.data.sa_email && (
                    <div><dt className="inline font-medium">Service account: </dt><dd className="inline font-mono text-xs">{status.data.sa_email}</dd></div>
                  )}
                  {status.data.project_id && (
                    <div><dt className="inline font-medium">Project ID: </dt><dd className="inline font-mono text-xs">{status.data.project_id}</dd></div>
                  )}
                  {status.data.updated_at && (
                    <div><dt className="inline font-medium">Updated: </dt><dd className="inline">{new Date(status.data.updated_at).toLocaleString()}</dd></div>
                  )}
                </dl>

                {status.data.sa_client_id && (
                  <div className="mt-3 rounded border border-blue-300 bg-blue-50 p-3">
                    <div className="text-sm font-medium text-blue-900">Client ID for Domain-Wide Delegation</div>
                    <p className="text-xs text-blue-800 mt-0.5">
                      Paste this exact value into Workspace Admin → Domain-Wide Delegation → Client ID, then add the four
                      OAuth scopes below and click <strong>Authorize</strong>.
                    </p>
                    <div className="mt-2 flex items-center gap-2">
                      <code className="flex-1 px-2 py-1 bg-white border border-blue-200 rounded font-mono text-sm select-all">
                        {status.data.sa_client_id}
                      </code>
                      <button
                        onClick={() => navigator.clipboard.writeText(status.data!.sa_client_id!)}
                        className="px-2 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700"
                      >
                        Copy
                      </button>
                      <a
                        href="https://admin.google.com/ac/owl/domainwidedelegation"
                        target="_blank"
                        rel="noreferrer"
                        className="px-2 py-1 text-xs bg-white border border-blue-300 text-blue-700 rounded hover:bg-blue-50"
                      >
                        Open DWD ↗
                      </a>
                    </div>
                  </div>
                )}
                <div className="mt-3 flex gap-2">
                  <button
                    onClick={() => test.mutate()}
                    disabled={test.isPending}
                    className="px-3 py-1.5 text-sm bg-white border border-green-400 text-green-800 rounded hover:bg-green-100 disabled:opacity-50"
                  >
                    {test.isPending ? 'Testing…' : 'Test connection'}
                  </button>
                  <button
                    onClick={() => {
                      if (confirm('Remove the stored Workspace credentials? Goldy will fall back to environment variables (if set) or stop serving Workspace endpoints.')) {
                        remove.mutate();
                      }
                    }}
                    disabled={remove.isPending}
                    className="px-3 py-1.5 text-sm bg-white border border-red-300 text-red-800 rounded hover:bg-red-50 disabled:opacity-50"
                  >
                    {remove.isPending ? 'Removing…' : 'Remove credentials'}
                  </button>
                </div>
              </>
            ) : (
              <div>
                <div className="font-medium text-amber-900">Workspace integration is not configured</div>
                <p className="text-amber-800 mt-1">
                  Until you upload a service-account JSON below, every endpoint under <code className="bg-amber-100 px-1 rounded">/api/v1/users</code> and <code className="bg-amber-100 px-1 rounded">/api/v1/groups</code> returns <code className="bg-amber-100 px-1 rounded">503 workspace_unavailable</code>.
                </p>
              </div>
            )}
          </div>
        )}
      </section>

      {/* Setup guide */}
      <details className="bg-white border border-gray-200 rounded-lg p-4 group" open>
        <summary className="cursor-pointer font-medium text-gray-900 select-none">
          📘 How to obtain the service-account JSON from Google Cloud Console
        </summary>
        <ol className="mt-3 ml-5 list-decimal space-y-2 text-sm text-gray-700">
          <li>
            Go to <a className="text-blue-600 hover:underline" href="https://console.cloud.google.com" target="_blank" rel="noreferrer">Google Cloud Console</a> and sign in with an account that can manage your Workspace organisation.
          </li>
          <li>
            Create a new project (or select an existing one) — for example <code className="bg-gray-100 px-1 rounded">goldy-workspace</code>.
          </li>
          <li>
            Enable the <strong>Admin SDK API</strong>: open{' '}
            <a className="text-blue-600 hover:underline" href="https://console.cloud.google.com/apis/library/admin.googleapis.com" target="_blank" rel="noreferrer">
              Admin SDK API → Enable
            </a>.
          </li>
          <li>
            Create a service account: <em>IAM &amp; Admin → Service Accounts → Create service account</em>. Skip the
            optional role assignment (Workspace permissions are granted via Domain-Wide Delegation, not IAM roles).
          </li>
          <li>
            Open the new service account → <em>Keys → Add Key → Create new key → JSON</em>. The JSON file
            downloads automatically — that's the file you upload below.
          </li>
          <li>
            On the same service-account page, copy the <strong>Unique ID</strong> (also called Client ID, a long
            number).
          </li>
          <li>
            In a new tab, open{' '}
            <a className="text-blue-600 hover:underline" href="https://admin.google.com/ac/owl/domainwidedelegation" target="_blank" rel="noreferrer">
              Workspace Admin Console → Security → Access &amp; data control → API controls → Manage Domain-Wide Delegation
            </a>.
          </li>
          <li>
            Click <strong>Add new</strong>, paste the Client ID, then paste the OAuth scopes below
            (comma-separated, all on one line). Click <strong>Authorize</strong>.
          </li>
          <li>
            Come back here, choose the JSON file, enter your super-admin email, and submit.
          </li>
        </ol>

        <div className="mt-4">
          <div className="text-xs font-medium text-gray-600 mb-1">Required OAuth scopes (paste these, comma-separated):</div>
          <textarea
            readOnly
            rows={4}
            className="w-full text-xs font-mono bg-gray-50 border border-gray-200 rounded p-2"
            value={(status.data?.required_scopes ?? []).join(',\n')}
            onClick={(e) => (e.target as HTMLTextAreaElement).select()}
          />
          <button
            onClick={() => navigator.clipboard.writeText((status.data?.required_scopes ?? []).join(','))}
            className="mt-1 text-xs text-blue-600 hover:underline"
          >
            Copy scopes to clipboard
          </button>
        </div>
      </details>

      {/* Upload form */}
      <section className="bg-white border border-gray-200 rounded-lg p-4 space-y-3">
        <h2 className="font-medium text-gray-900">Upload service-account JSON</h2>

        <label className="block text-sm">
          <span className="text-gray-700">Service-account JSON file</span>
          <input
            ref={fileInputRef}
            type="file"
            accept="application/json,.json"
            onChange={(e) => setFile(e.target.files?.[0] ?? null)}
            className="mt-1 block w-full text-sm file:mr-3 file:py-1.5 file:px-3 file:rounded file:border-0 file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
          />
          {file && <span className="text-xs text-gray-500 mt-1 block">Selected: {file.name} ({Math.round(file.size / 1024)} KB)</span>}
        </label>

        <label className="block text-sm">
          <span className="text-gray-700">Delegated super-admin email</span>
          <input
            type="email"
            placeholder="superadmin@yourdomain.com"
            value={delegatedAdmin}
            onChange={(e) => setDelegatedAdmin(e.target.value)}
            className="mt-1 block w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-blue-500"
            required
          />
          <span className="text-xs text-gray-500 mt-1 block">
            The service account will impersonate this Workspace admin when calling the Admin SDK.
          </span>
        </label>

        <label className="block text-sm">
          <span className="text-gray-700">Customer ID</span>
          <input
            type="text"
            value={customerID}
            onChange={(e) => setCustomerID(e.target.value)}
            placeholder="my_customer"
            className="mt-1 block w-full border border-gray-300 rounded px-3 py-1.5 text-sm font-mono focus:outline-none focus:border-blue-500"
          />
          <span className="text-xs text-gray-500 mt-1 block">
            Use <code className="bg-gray-100 px-1 rounded">my_customer</code> for a single-tenant Workspace, or your customer ID for multi-tenant.
          </span>
        </label>

        <button
          onClick={() => {
            setFeedback(null);
            upload.mutate();
          }}
          disabled={upload.isPending || !file || !delegatedAdmin}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium"
        >
          {upload.isPending ? 'Uploading…' : 'Save credentials'}
        </button>
      </section>

      {feedback && (
        <div
          className={`rounded border p-3 text-sm ${
            feedback.kind === 'ok'
              ? 'border-green-300 bg-green-50 text-green-800'
              : 'border-red-300 bg-red-50 text-red-800'
          }`}
        >
          <div className="whitespace-pre-wrap break-words">{feedback.msg}</div>
          {feedback.kind === 'err' && /unauthorized_client|DWD not authorized/i.test(feedback.msg) && status.data?.sa_client_id && (
            <div className="mt-2 pt-2 border-t border-red-200 text-xs text-red-900">
              <div>👉 Open <a className="underline" href="https://admin.google.com/ac/owl/domainwidedelegation" target="_blank" rel="noreferrer">Workspace Admin → Domain-Wide Delegation</a></div>
              <div>👉 Click <strong>Add new</strong> (or edit existing entry)</div>
              <div>👉 Client ID: <code className="bg-white px-1 rounded font-mono">{status.data.sa_client_id}</code></div>
              <div>👉 Paste all four scopes from the guide above (one line, comma-separated)</div>
              <div>👉 Click <strong>Authorize</strong>, wait 1–2 min, then Test again</div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
