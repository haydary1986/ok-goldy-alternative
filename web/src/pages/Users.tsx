import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import type { User, UsersListResponse } from '../lib/types';
import UserEditDrawer from '../components/UserEditDrawer';
import CreateUserModal from '../components/CreateUserModal';

const PAGE_SIZE = 100;

export default function Users() {
  const [pageStack, setPageStack] = useState<string[]>([]); // tokens we've seen, including current
  const currentToken = pageStack[pageStack.length - 1] ?? '';
  const [search, setSearch] = useState('');
  const [showSuspended, setShowSuspended] = useState<'all' | 'active' | 'suspended'>('all');
  const [editing, setEditing] = useState<User | null>(null);
  const [creating, setCreating] = useState(false);

  const q = useQuery({
    queryKey: ['users', currentToken],
    queryFn: () =>
      api<UsersListResponse>(
        `/users?page_size=${PAGE_SIZE}${currentToken ? `&page_token=${encodeURIComponent(currentToken)}` : ''}`,
      ),
  });

  const filtered = useMemo(() => {
    if (!q.data) return [];
    const s = search.trim().toLowerCase();
    return q.data.users.filter((u) => {
      if (showSuspended === 'active' && u.suspended) return false;
      if (showSuspended === 'suspended' && !u.suspended) return false;
      if (!s) return true;
      const hay = [u.primary_email, u.given_name, u.family_name, u.org_unit_path]
        .filter(Boolean)
        .join(' ')
        .toLowerCase();
      return hay.includes(s);
    });
  }, [q.data, search, showSuspended]);

  const goNext = () => {
    if (q.data?.next_page_token) {
      setPageStack((s) => [...s, q.data.next_page_token!]);
    }
  };
  const goPrev = () => {
    setPageStack((s) => (s.length > 0 ? s.slice(0, -1) : s));
  };
  const isFirstPage = pageStack.length === 0;

  return (
    <div className="space-y-4">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Users</h1>
          <p className="text-sm text-gray-500">
            {q.data
              ? `${filtered.length} shown of ${q.data.users.length} on this page${q.data.next_page_token ? ' (more pages →)' : ''}`
              : '…'}
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setCreating(true)}
            className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            + Add user
          </button>
        </div>
      </header>

      <div className="flex flex-wrap items-center gap-2">
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="🔍 Search email / name / OU on this page…"
          className="flex-1 min-w-[240px] border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-blue-500"
        />
        <select
          value={showSuspended}
          onChange={(e) => setShowSuspended(e.target.value as 'all' | 'active' | 'suspended')}
          className="border border-gray-300 rounded px-3 py-1.5 text-sm bg-white"
        >
          <option value="all">All</option>
          <option value="active">Active only</option>
          <option value="suspended">Suspended only</option>
        </select>
      </div>

      {q.isLoading && <div className="text-gray-500 text-sm">Loading users…</div>}
      {q.isError && (
        <div className="rounded border border-red-300 bg-red-50 p-3 text-red-700 text-sm whitespace-pre-wrap break-all">
          {(q.error as Error).message}
        </div>
      )}

      {q.data && (
        <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50 text-left">
              <tr>
                <th className="px-3 py-2 font-medium text-gray-700">Email</th>
                <th className="px-3 py-2 font-medium text-gray-700">Name</th>
                <th className="px-3 py-2 font-medium text-gray-700">Org Unit</th>
                <th className="px-3 py-2 font-medium text-gray-700">Status</th>
                <th className="px-3 py-2"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {filtered.map((u) => (
                <tr key={u.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => setEditing(u)}>
                  <td className="px-3 py-2 font-mono text-xs text-gray-800">{u.primary_email}</td>
                  <td className="px-3 py-2">
                    {[u.given_name, u.family_name].filter(Boolean).join(' ') || (
                      <span className="text-gray-400">—</span>
                    )}
                  </td>
                  <td className="px-3 py-2 text-gray-500 font-mono text-xs">{u.org_unit_path ?? '—'}</td>
                  <td className="px-3 py-2">
                    {u.suspended ? (
                      <span className="inline-flex items-center gap-1 text-red-700 text-xs">
                        <span className="w-1.5 h-1.5 rounded-full bg-red-500" />
                        Suspended
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 text-green-700 text-xs">
                        <span className="w-1.5 h-1.5 rounded-full bg-green-500" />
                        Active
                      </span>
                    )}
                    {u.is_admin && (
                      <span className="ml-2 text-xs text-blue-700 bg-blue-50 px-1.5 py-0.5 rounded">
                        admin
                      </span>
                    )}
                  </td>
                  <td className="px-3 py-2 text-right">
                    <span className="text-blue-600 hover:underline text-xs">Edit ↗</span>
                  </td>
                </tr>
              ))}
              {filtered.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-3 py-8 text-center text-gray-500">
                    No users match the current filters.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Pagination */}
      {q.data && (
        <div className="flex items-center justify-between text-sm">
          <button
            onClick={goPrev}
            disabled={isFirstPage}
            className="px-3 py-1.5 border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50"
          >
            ← Previous
          </button>
          <span className="text-gray-500">Page {pageStack.length + 1}</span>
          <button
            onClick={goNext}
            disabled={!q.data.next_page_token}
            className="px-3 py-1.5 border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50"
          >
            Next →
          </button>
        </div>
      )}

      <UserEditDrawer user={editing} onClose={() => setEditing(null)} />
      <CreateUserModal open={creating} onClose={() => setCreating(false)} />
    </div>
  );
}
