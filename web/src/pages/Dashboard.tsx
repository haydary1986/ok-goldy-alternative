import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import type { StatsOverview } from '../lib/types';

export default function Dashboard() {
  const q = useQuery({
    queryKey: ['stats-overview'],
    queryFn: () => api<StatsOverview>('/stats/overview'),
  });

  return (
    <div className="space-y-5">
      <header className="flex items-end justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-2xl font-semibold">Workspace overview</h1>
          <p className="text-sm text-gray-500">
            {q.data
              ? `Generated ${new Date(q.data.generated_at).toLocaleTimeString()} · ${q.data.duration_ms} ms`
              : 'Walking your Workspace tenant…'}
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

      {q.isLoading && <div className="text-gray-500 text-sm">Loading metrics… (first load walks the entire tenant, can take 5–15s)</div>}
      {q.isError && (
        <div className="rounded border border-red-300 bg-red-50 p-3 text-red-700 text-sm break-all">
          {(q.error as Error).message}
        </div>
      )}

      {q.data && (
        <>
          {/* Top row: Users summary */}
          <section className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <Card label="Total users" value={q.data.total_users} accent="blue" />
            <Card label="Active" value={q.data.active_users} accent="green" />
            <Card label="Suspended" value={q.data.suspended_users} accent="red" />
            <Card label="Super admins" value={q.data.admin_users} accent="purple" />
          </section>

          {/* Inactivity */}
          <section>
            <h2 className="text-sm font-semibold text-gray-700 mb-2">Inactive accounts (last login)</h2>
            <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
              <Card label="Never logged in" value={q.data.never_logged_in} accent="amber" link="/inactive?days=999999" />
              <Card label="30d+ inactive"  value={q.data.inactive_30d}  accent="amber" link="/inactive?days=30" />
              <Card label="90d+ inactive"  value={q.data.inactive_90d}  accent="amber" link="/inactive?days=90" />
              <Card label="180d+ inactive" value={q.data.inactive_180d} accent="orange" link="/inactive?days=180" />
              <Card label="365d+ inactive" value={q.data.inactive_365d} accent="orange" link="/inactive?days=365" />
            </div>
          </section>

          {/* Recent creations */}
          <section>
            <h2 className="text-sm font-semibold text-gray-700 mb-2">Recently created</h2>
            <div className="grid grid-cols-3 gap-3">
              <Card label="Last 7 days"  value={q.data.created_last_7d}  accent="blue" />
              <Card label="Last 30 days" value={q.data.created_last_30d} accent="blue" />
              <Card label="Last 90 days" value={q.data.created_last_90d} accent="blue" />
            </div>
          </section>

          {/* Groups */}
          <section>
            <h2 className="text-sm font-semibold text-gray-700 mb-2">Groups</h2>
            <div className="grid grid-cols-3 gap-3">
              <Card label="Total groups"   value={q.data.total_groups}        accent="indigo" />
              <Card label="Empty groups"   value={q.data.empty_groups}        accent="gray" />
              <Card label="Total memberships" value={q.data.total_group_members} accent="indigo" />
            </div>
          </section>

          {/* OU breakdown */}
          {q.data.users_by_ou.length > 0 && (
            <section>
              <h2 className="text-sm font-semibold text-gray-700 mb-2">Users per organisational unit</h2>
              <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
                <table className="w-full text-sm">
                  <thead className="bg-gray-50 text-left">
                    <tr>
                      <th className="px-3 py-2 font-medium text-gray-700">Org Unit</th>
                      <th className="px-3 py-2 font-medium text-gray-700 text-right">Total</th>
                      <th className="px-3 py-2 font-medium text-gray-700 text-right">Active</th>
                      <th className="px-3 py-2 font-medium text-gray-700 text-right">Suspended</th>
                      <th className="px-3 py-2 font-medium text-gray-700">Distribution</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    {q.data.users_by_ou.slice(0, 25).map((ou) => {
                      const pct = q.data!.total_users > 0 ? (ou.total / q.data!.total_users) * 100 : 0;
                      return (
                        <tr key={ou.org_unit_path} className="hover:bg-gray-50">
                          <td className="px-3 py-1.5 font-mono text-xs">{ou.org_unit_path}</td>
                          <td className="px-3 py-1.5 text-right font-medium">{ou.total.toLocaleString()}</td>
                          <td className="px-3 py-1.5 text-right text-green-700">{ou.active.toLocaleString()}</td>
                          <td className="px-3 py-1.5 text-right text-red-700">{ou.suspended.toLocaleString()}</td>
                          <td className="px-3 py-1.5">
                            <div className="bg-gray-100 rounded h-2 w-32 overflow-hidden">
                              <div className="bg-blue-500 h-full" style={{ width: `${Math.max(pct, 1)}%` }} />
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
                {q.data.users_by_ou.length > 25 && (
                  <div className="px-3 py-2 text-xs text-gray-500 bg-gray-50 border-t border-gray-100">
                    Showing top 25 of {q.data.users_by_ou.length} OUs.
                  </div>
                )}
              </div>
            </section>
          )}

          <section className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <ActionCard title="Find inactive accounts" description="Filter and bulk-suspend stale users" to="/inactive" />
            <ActionCard title="Audit log" description="Every mutation, recorded" to="/audit" />
            <ActionCard title="Settings" description="Workspace credentials + scopes" to="/settings" />
          </section>
        </>
      )}
    </div>
  );
}

function Card({
  label,
  value,
  accent,
  link,
}: {
  label: string;
  value: number;
  accent: 'blue' | 'green' | 'red' | 'amber' | 'orange' | 'purple' | 'indigo' | 'gray';
  link?: string;
}) {
  const colors: Record<string, string> = {
    blue: 'border-blue-200 bg-blue-50 text-blue-900',
    green: 'border-green-200 bg-green-50 text-green-900',
    red: 'border-red-200 bg-red-50 text-red-900',
    amber: 'border-amber-200 bg-amber-50 text-amber-900',
    orange: 'border-orange-200 bg-orange-50 text-orange-900',
    purple: 'border-purple-200 bg-purple-50 text-purple-900',
    indigo: 'border-indigo-200 bg-indigo-50 text-indigo-900',
    gray: 'border-gray-200 bg-gray-50 text-gray-900',
  };
  const inner = (
    <div className={`border rounded-lg px-3 py-2.5 ${colors[accent]}`}>
      <div className="text-2xl font-semibold">{value.toLocaleString()}</div>
      <div className="text-xs uppercase tracking-wide opacity-75 mt-0.5">{label}</div>
    </div>
  );
  if (link) return <Link to={link} className="block hover:opacity-80 transition">{inner}</Link>;
  return inner;
}

function ActionCard({ title, description, to }: { title: string; description: string; to: string }) {
  return (
    <Link
      to={to}
      className="block bg-white border border-gray-200 rounded-lg p-4 hover:shadow-sm hover:border-gray-300 transition"
    >
      <div className="text-base font-medium text-gray-900">{title}</div>
      <div className="text-sm text-gray-500 mt-1">{description}</div>
    </Link>
  );
}
