import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import type { GroupsListResponse } from '../lib/types';

export default function Groups() {
  const q = useQuery({
    queryKey: ['groups'],
    queryFn: () => api<GroupsListResponse>('/groups?page_size=100'),
  });

  return (
    <div className="space-y-4">
      <header className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Groups</h1>
        <span className="text-sm text-gray-500">
          {q.data?.groups.length ?? 0} loaded
          {q.data?.next_page_token ? ' (more pages available)' : ''}
        </span>
      </header>

      {q.isLoading && <div className="text-gray-500 text-sm">Loading groups…</div>}
      {q.isError && (
        <div className="rounded border border-red-300 bg-red-50 p-3 text-red-700 text-sm">
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
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {q.data.groups.map((g) => (
                <tr key={g.id}>
                  <td className="px-3 py-2 font-mono text-xs text-gray-800">{g.email}</td>
                  <td className="px-3 py-2">{g.name ?? <span className="text-gray-400">—</span>}</td>
                  <td className="px-3 py-2 text-gray-500">{g.description ?? '—'}</td>
                  <td className="px-3 py-2 text-right">{g.direct_members_count ?? 0}</td>
                </tr>
              ))}
              {q.data.groups.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-3 py-8 text-center text-gray-500">
                    No groups returned.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
