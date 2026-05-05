import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import type { AuditEntry, AuditListResponse } from '../lib/types';

const ACTION_OPTIONS = ['', 'create', 'update', 'delete', 'suspend', 'restore', 'export'];
const RESOURCE_OPTIONS = ['', 'user', 'group', 'group_member', 'user_alias', 'org_unit', 'workspace_credentials'];
const PAGE_SIZE = 100;

export default function Audit() {
  const [actor, setActor] = useState('');
  const [action, setAction] = useState('');
  const [resourceType, setResourceType] = useState('');
  const [onlyFailures, setOnlyFailures] = useState(false);
  const [offset, setOffset] = useState(0);
  const [expanded, setExpanded] = useState<number | null>(null);

  const buildPath = () => {
    const params = new URLSearchParams();
    params.set('limit', String(PAGE_SIZE));
    params.set('offset', String(offset));
    if (actor) params.set('actor', actor);
    if (action) params.set('action', action);
    if (resourceType) params.set('resource_type', resourceType);
    if (onlyFailures) params.set('only_failures', 'true');
    return `/audit?${params.toString()}`;
  };

  const q = useQuery({
    queryKey: ['audit', actor, action, resourceType, onlyFailures, offset],
    queryFn: () => api<AuditListResponse>(buildPath()),
  });

  const onFilterChange = () => setOffset(0);

  return (
    <div className="space-y-4">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Audit log</h1>
          <p className="text-sm text-gray-500">
            Every mutation Goldy performs, with actor, action, target, and a JSON snapshot.
          </p>
        </div>
        <button
          onClick={() => q.refetch()}
          disabled={q.isFetching}
          className="px-3 py-1.5 text-sm border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50"
        >
          {q.isFetching ? 'Refreshing…' : 'Refresh'}
        </button>
      </header>

      <section className="bg-white border border-gray-200 rounded-lg p-3 grid grid-cols-1 sm:grid-cols-4 gap-3">
        <label className="text-xs">
          <span className="text-gray-600 font-medium">Actor</span>
          <input
            value={actor}
            onChange={(e) => {
              setActor(e.target.value);
              onFilterChange();
            }}
            placeholder="email@..."
            className="mt-0.5 w-full border border-gray-300 rounded px-2 py-1 text-sm"
          />
        </label>
        <label className="text-xs">
          <span className="text-gray-600 font-medium">Action</span>
          <select
            value={action}
            onChange={(e) => {
              setAction(e.target.value);
              onFilterChange();
            }}
            className="mt-0.5 w-full border border-gray-300 rounded px-2 py-1 text-sm bg-white"
          >
            {ACTION_OPTIONS.map((a) => (
              <option key={a} value={a}>
                {a || '— any —'}
              </option>
            ))}
          </select>
        </label>
        <label className="text-xs">
          <span className="text-gray-600 font-medium">Resource type</span>
          <select
            value={resourceType}
            onChange={(e) => {
              setResourceType(e.target.value);
              onFilterChange();
            }}
            className="mt-0.5 w-full border border-gray-300 rounded px-2 py-1 text-sm bg-white"
          >
            {RESOURCE_OPTIONS.map((r) => (
              <option key={r} value={r}>
                {r || '— any —'}
              </option>
            ))}
          </select>
        </label>
        <label className="text-xs flex items-end gap-2 pb-1">
          <input
            type="checkbox"
            checked={onlyFailures}
            onChange={(e) => {
              setOnlyFailures(e.target.checked);
              onFilterChange();
            }}
            className="rounded"
          />
          <span className="text-gray-700">Only failures</span>
        </label>
      </section>

      {q.isLoading && <div className="text-gray-500 text-sm">Loading audit entries…</div>}
      {q.isError && (
        <div className="rounded border border-red-300 bg-red-50 p-3 text-red-700 text-sm break-all">
          {(q.error as Error).message}
        </div>
      )}

      {q.data && (
        <>
          <div className="text-xs text-gray-500">
            {q.data.total.toLocaleString()} entr{q.data.total === 1 ? 'y' : 'ies'} · showing {offset + 1}–
            {Math.min(offset + q.data.entries.length, q.data.total)}
          </div>

          <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead className="bg-gray-50 text-left">
                <tr>
                  <th className="px-3 py-2 font-medium text-gray-700">Time</th>
                  <th className="px-3 py-2 font-medium text-gray-700">Actor</th>
                  <th className="px-3 py-2 font-medium text-gray-700">Action</th>
                  <th className="px-3 py-2 font-medium text-gray-700">Resource</th>
                  <th className="px-3 py-2 font-medium text-gray-700"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {q.data.entries.map((e) => (
                  <Row
                    key={e.id}
                    entry={e}
                    open={expanded === e.id}
                    onToggle={() => setExpanded(expanded === e.id ? null : e.id)}
                  />
                ))}
                {q.data.entries.length === 0 && (
                  <tr>
                    <td colSpan={5} className="px-3 py-8 text-center text-gray-500">
                      No audit entries match the current filters.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          <div className="flex justify-between text-sm">
            <button
              onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
              disabled={offset === 0}
              className="px-3 py-1.5 border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50"
            >
              ← Previous
            </button>
            <button
              onClick={() => setOffset(offset + PAGE_SIZE)}
              disabled={offset + PAGE_SIZE >= q.data.total}
              className="px-3 py-1.5 border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50"
            >
              Next →
            </button>
          </div>
        </>
      )}
    </div>
  );
}

function Row({ entry, open, onToggle }: { entry: AuditEntry; open: boolean; onToggle: () => void }) {
  return (
    <>
      <tr className={`hover:bg-gray-50 cursor-pointer ${entry.ok ? '' : 'bg-red-50'}`} onClick={onToggle}>
        <td className="px-3 py-1.5 text-xs text-gray-600 whitespace-nowrap">
          {new Date(entry.occurred_at).toLocaleString()}
        </td>
        <td className="px-3 py-1.5 text-xs font-mono">{entry.actor}</td>
        <td className="px-3 py-1.5">
          <span
            className={`text-xs px-1.5 py-0.5 rounded ${
              entry.action === 'delete'
                ? 'bg-red-100 text-red-800'
                : entry.action === 'suspend'
                ? 'bg-amber-100 text-amber-800'
                : entry.action === 'create'
                ? 'bg-green-100 text-green-800'
                : 'bg-blue-100 text-blue-800'
            }`}
          >
            {entry.action}
          </span>
          {!entry.ok && <span className="ml-1 text-xs text-red-700">⚠</span>}
        </td>
        <td className="px-3 py-1.5 text-xs">
          <span className="text-gray-500">{entry.resource_type}/</span>
          <span className="font-mono">{entry.resource_id}</span>
        </td>
        <td className="px-3 py-1.5 text-xs text-blue-600">{open ? '▾' : '▸'}</td>
      </tr>
      {open && (
        <tr className="bg-gray-50">
          <td colSpan={5} className="px-3 py-2">
            <div className="text-xs space-y-1">
              {entry.error_message && (
                <div className="text-red-700">
                  <strong>Error:</strong> {entry.error_message}
                </div>
              )}
              {entry.request_id && (
                <div>
                  <strong>Request ID:</strong> <code className="font-mono">{entry.request_id}</code>
                </div>
              )}
              {entry.before !== undefined && entry.before !== null && (
                <details>
                  <summary className="cursor-pointer text-gray-700">Before</summary>
                  <pre className="bg-white p-2 rounded border border-gray-200 mt-1 overflow-auto text-[11px]">
                    {JSON.stringify(entry.before, null, 2)}
                  </pre>
                </details>
              )}
              {entry.after !== undefined && entry.after !== null && (
                <details open>
                  <summary className="cursor-pointer text-gray-700">After</summary>
                  <pre className="bg-white p-2 rounded border border-gray-200 mt-1 overflow-auto text-[11px]">
                    {JSON.stringify(entry.after, null, 2)}
                  </pre>
                </details>
              )}
            </div>
          </td>
        </tr>
      )}
    </>
  );
}
