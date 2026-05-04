import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import type { UsersListResponse } from '../lib/types';

export default function Users() {
  const q = useQuery({
    queryKey: ['users'],
    queryFn: () => api<UsersListResponse>('/users?page_size=100'),
  });

  return (
    <div className="space-y-4">
      <header className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Users</h1>
        <span className="text-sm text-gray-500">
          {q.data?.users.length ?? 0} loaded
          {q.data?.next_page_token ? ' (more pages available)' : ''}
        </span>
      </header>

      {q.isLoading && <div className="text-gray-500 text-sm">Loading users…</div>}
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
                <th className="px-3 py-2 font-medium text-gray-700">Org Unit</th>
                <th className="px-3 py-2 font-medium text-gray-700">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {q.data.users.map((u) => (
                <tr key={u.id}>
                  <td className="px-3 py-2 font-mono text-xs text-gray-800">{u.primary_email}</td>
                  <td className="px-3 py-2">
                    {[u.given_name, u.family_name].filter(Boolean).join(' ') || (
                      <span className="text-gray-400">—</span>
                    )}
                  </td>
                  <td className="px-3 py-2 text-gray-500">{u.org_unit_path ?? '—'}</td>
                  <td className="px-3 py-2">
                    {u.suspended ? (
                      <span className="text-red-600">Suspended</span>
                    ) : (
                      <span className="text-green-600">Active</span>
                    )}
                  </td>
                </tr>
              ))}
              {q.data.users.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-3 py-8 text-center text-gray-500">
                    No users returned.
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
