import { useState, useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api, apiPost } from '../lib/api';
import { useToast } from '../components/Toast';
import type {
  BulkSuspendRequest,
  BulkSuspendResponse,
  InactiveListResponse,
  User,
} from '../lib/types';

const PRESETS = [
  { days: 30, label: '30 days' },
  { days: 60, label: '60 days' },
  { days: 90, label: '90 days' },
  { days: 180, label: '180 days' },
  { days: 365, label: '1 year' },
];

export default function Inactive() {
  const qc = useQueryClient();
  const toast = useToast();
  const [params, setParams] = useSearchParams();
  const initialDays = parseInt(params.get('days') ?? '90', 10);
  const [days, setDays] = useState<number>(isNaN(initialDays) ? 90 : initialDays);
  const [includeAdmins, setIncludeAdmins] = useState(false);
  const [includeSuspended, setIncludeSuspended] = useState(false);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [search, setSearch] = useState('');

  const q = useQuery({
    queryKey: ['inactive', days, includeAdmins, includeSuspended],
    queryFn: () =>
      api<InactiveListResponse>(
        `/users/inactive?days=${days}&include_admins=${includeAdmins}&include_suspended=${includeSuspended}`,
      ),
  });

  const filtered = useMemo<User[]>(() => {
    if (!q.data) return [];
    const s = search.trim().toLowerCase();
    if (!s) return q.data.users;
    return q.data.users.filter((u) =>
      [u.primary_email, u.given_name, u.family_name, u.org_unit_path]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
        .includes(s),
    );
  }, [q.data, search]);

  const allFilteredSelected =
    filtered.length > 0 && filtered.every((u) => selected.has(u.id));

  const toggleAll = () => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (allFilteredSelected) {
        filtered.forEach((u) => next.delete(u.id));
      } else {
        filtered.forEach((u) => next.add(u.id));
      }
      return next;
    });
  };
  const toggle = (id: string) =>
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  const bulkSuspend = useMutation({
    mutationFn: () => {
      const ids = Array.from(selected);
      const body: BulkSuspendRequest = { user_ids: ids, suspended: true };
      return apiPost<BulkSuspendResponse>('/users/bulk/suspend', body);
    },
    onSuccess: (resp) => {
      if (resp.failed === 0) {
        toast.success(`Suspended ${resp.successful} users.`);
      } else {
        toast.error(`${resp.successful} suspended, ${resp.failed} failed.`);
      }
      setSelected(new Set());
      qc.invalidateQueries({ queryKey: ['inactive'] });
      qc.invalidateQueries({ queryKey: ['users'] });
      qc.invalidateQueries({ queryKey: ['stats-overview'] });
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const updateDays = (newDays: number) => {
    setDays(newDays);
    setSelected(new Set());
    setParams({ days: String(newDays) });
  };

  return (
    <div className="space-y-4">
      <header className="flex items-start justify-between gap-3 flex-wrap">
        <div>
          <h1 className="text-2xl font-semibold">Inactive accounts</h1>
          <p className="text-sm text-gray-500">
            Users who haven't signed in for the chosen window — or never have.
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

      {/* Filters */}
      <section className="bg-white border border-gray-200 rounded-lg p-4 space-y-3">
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-sm text-gray-700 mr-2">Last login older than:</span>
          {PRESETS.map((p) => (
            <button
              key={p.days}
              onClick={() => updateDays(p.days)}
              className={`px-3 py-1.5 text-sm rounded border ${
                days === p.days
                  ? 'bg-blue-600 text-white border-blue-600'
                  : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
              }`}
            >
              {p.label}
            </button>
          ))}
          <input
            type="number"
            value={days}
            onChange={(e) => updateDays(parseInt(e.target.value, 10) || 0)}
            className="w-24 border border-gray-300 rounded px-2 py-1 text-sm"
            min={1}
          />
          <span className="text-sm text-gray-500">days</span>
        </div>
        <div className="flex items-center gap-4 text-sm">
          <label className="inline-flex items-center gap-2">
            <input
              type="checkbox"
              checked={includeAdmins}
              onChange={(e) => setIncludeAdmins(e.target.checked)}
              className="rounded"
            />
            <span>Include super admins</span>
          </label>
          <label className="inline-flex items-center gap-2">
            <input
              type="checkbox"
              checked={includeSuspended}
              onChange={(e) => setIncludeSuspended(e.target.checked)}
              className="rounded"
            />
            <span>Include already-suspended users</span>
          </label>
        </div>
      </section>

      {q.isLoading && (
        <div className="text-gray-500 text-sm">
          Walking the full Workspace user list to compute inactivity… (5–15s on first run)
        </div>
      )}
      {q.isError && (
        <div className="rounded border border-red-300 bg-red-50 p-3 text-red-700 text-sm break-all">
          {(q.error as Error).message}
        </div>
      )}

      {q.data && (
        <>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div className="text-sm">
              Found <strong>{q.data.total.toLocaleString()}</strong> inactive users
              <span className="text-gray-500"> (cutoff: {new Date(q.data.cutoff).toLocaleDateString()})</span>
            </div>
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="🔍 Filter loaded list…"
              className="border border-gray-300 rounded px-3 py-1.5 text-sm w-56"
            />
          </div>

          {selected.size > 0 && (
            <div className="bg-amber-50 border border-amber-300 rounded-lg p-3 flex items-center justify-between gap-3 flex-wrap">
              <div className="text-sm text-amber-900">
                <strong>{selected.size.toLocaleString()}</strong> user{selected.size === 1 ? '' : 's'} selected
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => setSelected(new Set())}
                  className="px-3 py-1.5 text-sm bg-white border border-gray-300 rounded hover:bg-gray-50"
                >
                  Clear
                </button>
                <button
                  onClick={() => {
                    if (
                      confirm(
                        `Suspend ${selected.size} user${selected.size === 1 ? '' : 's'}?\n\nThey lose Workspace access until restored, but their data is preserved.`,
                      )
                    ) {
                      bulkSuspend.mutate();
                    }
                  }}
                  disabled={bulkSuspend.isPending}
                  className="px-3 py-1.5 text-sm bg-red-600 text-white rounded hover:bg-red-700 disabled:opacity-50"
                >
                  {bulkSuspend.isPending ? 'Suspending…' : `Suspend ${selected.size}`}
                </button>
              </div>
            </div>
          )}

          <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead className="bg-gray-50 text-left">
                <tr>
                  <th className="px-3 py-2 w-8">
                    <input
                      type="checkbox"
                      checked={allFilteredSelected}
                      onChange={toggleAll}
                      className="rounded"
                      aria-label="Select all"
                    />
                  </th>
                  <th className="px-3 py-2 font-medium text-gray-700">Email</th>
                  <th className="px-3 py-2 font-medium text-gray-700">Last login</th>
                  <th className="px-3 py-2 font-medium text-gray-700">Days inactive</th>
                  <th className="px-3 py-2 font-medium text-gray-700">OU</th>
                  <th className="px-3 py-2 font-medium text-gray-700">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filtered.map((u) => (
                  <tr key={u.id} className={`hover:bg-gray-50 ${selected.has(u.id) ? 'bg-blue-50' : ''}`}>
                    <td className="px-3 py-1.5">
                      <input
                        type="checkbox"
                        checked={selected.has(u.id)}
                        onChange={() => toggle(u.id)}
                        className="rounded"
                        aria-label={`Select ${u.primary_email}`}
                      />
                    </td>
                    <td className="px-3 py-1.5 font-mono text-xs">{u.primary_email}</td>
                    <td className="px-3 py-1.5 text-xs text-gray-600">
                      {formatLastLogin(u.last_login_time)}
                    </td>
                    <td className="px-3 py-1.5 text-xs">{daysSince(u.last_login_time)}</td>
                    <td className="px-3 py-1.5 font-mono text-xs text-gray-500">
                      {u.org_unit_path ?? '—'}
                    </td>
                    <td className="px-3 py-1.5 text-xs">
                      {u.suspended ? (
                        <span className="text-red-700">Suspended</span>
                      ) : (
                        <span className="text-green-700">Active</span>
                      )}
                      {u.is_admin && (
                        <span className="ml-1 text-blue-700 bg-blue-50 px-1 rounded text-[10px]">admin</span>
                      )}
                    </td>
                  </tr>
                ))}
                {filtered.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-3 py-8 text-center text-gray-500">
                      🎉 No inactive users for this window.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  );
}

function formatLastLogin(iso?: string): string {
  if (!iso) return '—';
  const t = new Date(iso);
  if (t.getFullYear() < 1971 || t.getFullYear() === 1) return 'Never';
  return t.toLocaleDateString();
}

function daysSince(iso?: string): string {
  if (!iso) return 'Never';
  const t = new Date(iso);
  if (t.getFullYear() < 1971 || t.getFullYear() === 1) return 'Never';
  const diff = Math.floor((Date.now() - t.getTime()) / (1000 * 60 * 60 * 24));
  return diff.toLocaleString() + 'd';
}
