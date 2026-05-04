import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import type { Group, GroupsListResponse } from '../lib/types';
import GroupEditDrawer from '../components/GroupEditDrawer';
import CreateGroupModal from '../components/CreateGroupModal';

const PAGE_SIZE = 100;

export default function Groups() {
  const [pageStack, setPageStack] = useState<string[]>([]);
  const currentToken = pageStack[pageStack.length - 1] ?? '';
  const [search, setSearch] = useState('');
  const [editing, setEditing] = useState<Group | null>(null);
  const [creating, setCreating] = useState(false);

  const q = useQuery({
    queryKey: ['groups', currentToken],
    queryFn: () =>
      api<GroupsListResponse>(
        `/groups?page_size=${PAGE_SIZE}${currentToken ? `&page_token=${encodeURIComponent(currentToken)}` : ''}`,
      ),
  });

  const filtered = useMemo(() => {
    if (!q.data) return [];
    const s = search.trim().toLowerCase();
    return q.data.groups.filter((g) => {
      if (!s) return true;
      const hay = [g.email, g.name, g.description].filter(Boolean).join(' ').toLowerCase();
      return hay.includes(s);
    });
  }, [q.data, search]);

  const goNext = () => {
    if (q.data?.next_page_token) {
      setPageStack((s) => [...s, q.data.next_page_token!]);
    }
  };
  const goPrev = () => setPageStack((s) => (s.length > 0 ? s.slice(0, -1) : s));
  const isFirstPage = pageStack.length === 0;

  return (
    <div className="space-y-4">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Groups</h1>
          <p className="text-sm text-gray-500">
            {q.data
              ? `${filtered.length} shown of ${q.data.groups.length} on this page${q.data.next_page_token ? ' (more pages →)' : ''}`
              : '…'}
          </p>
        </div>
        <button
          onClick={() => setCreating(true)}
          className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          + Add group
        </button>
      </header>

      <input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="🔍 Search email / name / description on this page…"
        className="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-blue-500"
      />

      {q.isLoading && <div className="text-gray-500 text-sm">Loading groups…</div>}
      {q.isError && (
        <div className="rounded border border-red-300 bg-red-50 p-3 text-red-700 text-sm break-all">
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
                <th className="px-3 py-2 font-medium text-gray-700">Description</th>
                <th className="px-3 py-2 font-medium text-gray-700 text-right">Members</th>
                <th className="px-3 py-2"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {filtered.map((g) => (
                <tr key={g.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => setEditing(g)}>
                  <td className="px-3 py-2 font-mono text-xs text-gray-800">{g.email}</td>
                  <td className="px-3 py-2">{g.name ?? <span className="text-gray-400">—</span>}</td>
                  <td className="px-3 py-2 text-gray-500 truncate max-w-md">{g.description ?? '—'}</td>
                  <td className="px-3 py-2 text-right">{g.direct_members_count ?? 0}</td>
                  <td className="px-3 py-2 text-right">
                    <span className="text-blue-600 hover:underline text-xs">Manage ↗</span>
                  </td>
                </tr>
              ))}
              {filtered.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-3 py-8 text-center text-gray-500">
                    No groups match the current search.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

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

      <GroupEditDrawer group={editing} onClose={() => setEditing(null)} />
      <CreateGroupModal open={creating} onClose={() => setCreating(false)} />
    </div>
  );
}
