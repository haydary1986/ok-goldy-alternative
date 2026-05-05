import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import type {
  StatsOverview,
  UsageSnapshot,
  UsersListResponse,
  User,
} from '../lib/types';

// Reports page goes deeper than Dashboard: it joins per-user usage data
// (Drive / Gmail) with the directory list to produce ranked tables and
// per-OU storage rollups that Workspace Admin Console can't easily show.

export default function Reports() {
  const [topN, setTopN] = useState(50);

  const stats = useQuery({
    queryKey: ['stats-overview'],
    queryFn: () => api<StatsOverview>('/stats/overview'),
  });

  const usage = useQuery({
    queryKey: ['usage-snapshot'],
    queryFn: () => api<UsageSnapshot>('/usage/users'),
    staleTime: 30 * 60_000,
    retry: false,
  });

  // Walk the first page of users — for top consumer rankings we need user
  // name + OU joined with usage. Reports page is best-effort, so we use
  // the same /users page already cached by the Users page.
  const users = useQuery({
    queryKey: ['users-all-snapshot'],
    queryFn: async () => {
      // Fetch up to 5 pages so the top-N rankings have enough range.
      let token: string | undefined = undefined;
      const all: User[] = [];
      for (let i = 0; i < 60; i++) {
        const r: UsersListResponse = await api<UsersListResponse>(
          `/users?page_size=500${token ? `&page_token=${encodeURIComponent(token)}` : ''}`,
        );
        all.push(...r.users);
        if (!r.next_page_token) break;
        token = r.next_page_token;
      }
      return all;
    },
    staleTime: 5 * 60_000,
  });

  const topConsumers = useMemo(() => {
    if (!usage.data) return [];
    const list = Object.values(usage.data.users)
      .filter((u) => u.drive_used_mb > 0)
      .sort((a, b) => b.drive_used_mb - a.drive_used_mb)
      .slice(0, topN);
    return list;
  }, [usage.data, topN]);

  const topGmailReceivers = useMemo(() => {
    if (!usage.data) return [];
    return Object.values(usage.data.users)
      .filter((u) => u.gmail_num_received > 0)
      .sort((a, b) => b.gmail_num_received - a.gmail_num_received)
      .slice(0, topN);
  }, [usage.data, topN]);

  const ouStorage = useMemo(() => {
    if (!usage.data || !users.data) return [];
    const byOU: Record<string, { ou: string; usedMB: number; users: number }> = {};
    for (const u of users.data) {
      const usg = usage.data.users[u.primary_email];
      if (!usg) continue;
      const ou = u.org_unit_path || '/';
      if (!byOU[ou]) byOU[ou] = { ou, usedMB: 0, users: 0 };
      byOU[ou].usedMB += usg.drive_used_mb;
      byOU[ou].users++;
    }
    return Object.values(byOU)
      .sort((a, b) => b.usedMB - a.usedMB)
      .slice(0, 25);
  }, [usage.data, users.data]);

  const userByEmail = useMemo(() => {
    const m: Record<string, User> = {};
    for (const u of users.data ?? []) m[u.primary_email] = u;
    return m;
  }, [users.data]);

  const totalDriveTB = useMemo(() => {
    if (!usage.data) return 0;
    let totalMB = 0;
    for (const u of Object.values(usage.data.users)) totalMB += u.drive_used_mb;
    return totalMB / 1024 / 1024;
  }, [usage.data]);

  return (
    <div className="space-y-6">
      <header className="flex items-end justify-between gap-3 flex-wrap">
        <div>
          <h1 className="text-2xl font-semibold">Reports</h1>
          <p className="text-sm text-gray-500">
            Storage and Gmail-activity rankings sourced from Admin Reports API.
            {usage.data && (
              <>
                {' '}Snapshot date: {new Date(usage.data.date).toLocaleDateString()} ·
                {' '}{usage.data.total_users.toLocaleString()} users
              </>
            )}
          </p>
        </div>
        <div className="flex gap-2 items-center">
          <label className="text-sm">
            Top
            <input
              type="number"
              value={topN}
              onChange={(e) => setTopN(Math.max(5, Math.min(500, parseInt(e.target.value, 10) || 50)))}
              className="ml-1 w-20 border border-gray-300 rounded px-2 py-1 text-sm"
              min={5}
              max={500}
            />
          </label>
          <button
            onClick={() => {
              stats.refetch();
              usage.refetch();
              users.refetch();
            }}
            disabled={stats.isFetching || usage.isFetching || users.isFetching}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50"
          >
            Refresh
          </button>
        </div>
      </header>

      {usage.isError && (
        <div className="rounded border border-amber-300 bg-amber-50 p-3 text-amber-900 text-sm">
          Drive / Gmail data unavailable: {(usage.error as Error).message.slice(0, 300)}
          <div className="mt-1 text-xs">
            Add the <code className="bg-amber-100 px-1 rounded">admin.reports.usage.readonly</code> scope to your DWD
            entry then reload. See <a href="/settings" className="underline">Settings</a>.
          </div>
        </div>
      )}

      {stats.data && usage.data && (
        <section className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <Tile label="Total Drive used" value={`${totalDriveTB.toFixed(2)} TB`} accent="indigo" />
          <Tile label="Users w/ Drive presence" value={countTrue(usage.data, 'has_drive_presence').toLocaleString()} accent="blue" />
          <Tile label="Users w/ Gmail presence" value={countTrue(usage.data, 'has_gmail_presence').toLocaleString()} accent="green" />
          <Tile label="Inactive 90d+" value={stats.data.inactive_90d.toLocaleString()} accent="amber" />
        </section>
      )}

      {/* Top Drive consumers */}
      {usage.data && (
        <section className="space-y-2">
          <h2 className="text-sm font-semibold text-gray-700">Top {topN} Drive consumers</h2>
          <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 text-left">
                <tr>
                  <th className="px-3 py-2 font-medium text-gray-700 w-12">#</th>
                  <th className="px-3 py-2 font-medium text-gray-700">Email</th>
                  <th className="px-3 py-2 font-medium text-gray-700">OU</th>
                  <th className="px-3 py-2 font-medium text-gray-700 text-right">Drive used</th>
                  <th className="px-3 py-2 font-medium text-gray-700 text-right">Items</th>
                  <th className="px-3 py-2 font-medium text-gray-700">Last mail</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {topConsumers.map((u, i) => {
                  const dirUser = userByEmail[u.user_email];
                  return (
                    <tr key={u.user_email} className="hover:bg-gray-50">
                      <td className="px-3 py-1.5 text-xs text-gray-500">{i + 1}</td>
                      <td className="px-3 py-1.5 font-mono text-xs">{u.user_email}</td>
                      <td className="px-3 py-1.5 font-mono text-xs text-gray-500">
                        {dirUser?.org_unit_path ?? '—'}
                      </td>
                      <td className="px-3 py-1.5 text-right">
                        {(u.drive_used_mb / 1024).toFixed(2)} GB
                      </td>
                      <td className="px-3 py-1.5 text-right text-xs">
                        {u.drive_items_owned.toLocaleString()}
                      </td>
                      <td className="px-3 py-1.5 text-xs text-gray-600">
                        {formatDate(u.gmail_last_interaction)}
                      </td>
                    </tr>
                  );
                })}
                {topConsumers.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-3 py-8 text-center text-gray-500">
                      No Drive usage data available yet.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </section>
      )}

      {/* Top Gmail receivers */}
      {usage.data && (
        <section className="space-y-2">
          <h2 className="text-sm font-semibold text-gray-700">Top {topN} Gmail receivers</h2>
          <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 text-left">
                <tr>
                  <th className="px-3 py-2 font-medium text-gray-700 w-12">#</th>
                  <th className="px-3 py-2 font-medium text-gray-700">Email</th>
                  <th className="px-3 py-2 font-medium text-gray-700">OU</th>
                  <th className="px-3 py-2 font-medium text-gray-700 text-right">Received</th>
                  <th className="px-3 py-2 font-medium text-gray-700 text-right">Sent</th>
                  <th className="px-3 py-2 font-medium text-gray-700">Last interaction</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {topGmailReceivers.map((u, i) => {
                  const dirUser = userByEmail[u.user_email];
                  return (
                    <tr key={u.user_email} className="hover:bg-gray-50">
                      <td className="px-3 py-1.5 text-xs text-gray-500">{i + 1}</td>
                      <td className="px-3 py-1.5 font-mono text-xs">{u.user_email}</td>
                      <td className="px-3 py-1.5 font-mono text-xs text-gray-500">
                        {dirUser?.org_unit_path ?? '—'}
                      </td>
                      <td className="px-3 py-1.5 text-right">{u.gmail_num_received.toLocaleString()}</td>
                      <td className="px-3 py-1.5 text-right text-gray-600">
                        {u.gmail_num_sent.toLocaleString()}
                      </td>
                      <td className="px-3 py-1.5 text-xs text-gray-600">
                        {formatDate(u.gmail_last_interaction)}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </section>
      )}

      {/* Per-OU storage breakdown */}
      {ouStorage.length > 0 && (
        <section className="space-y-2">
          <h2 className="text-sm font-semibold text-gray-700">Storage by Organisational Unit (top 25)</h2>
          <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 text-left">
                <tr>
                  <th className="px-3 py-2 font-medium text-gray-700">Org Unit</th>
                  <th className="px-3 py-2 font-medium text-gray-700 text-right">Users</th>
                  <th className="px-3 py-2 font-medium text-gray-700 text-right">Drive used</th>
                  <th className="px-3 py-2 font-medium text-gray-700 text-right">Avg per user</th>
                  <th className="px-3 py-2 font-medium text-gray-700">Share</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {ouStorage.map((b) => {
                  const totalUsedMB = ouStorage.reduce((s, x) => s + x.usedMB, 0);
                  const pct = totalUsedMB > 0 ? (b.usedMB / totalUsedMB) * 100 : 0;
                  const avgGB = b.users > 0 ? b.usedMB / 1024 / b.users : 0;
                  return (
                    <tr key={b.ou}>
                      <td className="px-3 py-1.5 font-mono text-xs">{b.ou}</td>
                      <td className="px-3 py-1.5 text-right">{b.users.toLocaleString()}</td>
                      <td className="px-3 py-1.5 text-right">{(b.usedMB / 1024).toFixed(1)} GB</td>
                      <td className="px-3 py-1.5 text-right text-gray-600">{avgGB.toFixed(2)} GB</td>
                      <td className="px-3 py-1.5">
                        <div className="bg-gray-100 rounded h-2 w-32 overflow-hidden">
                          <div className="bg-indigo-500 h-full" style={{ width: `${Math.max(pct, 1)}%` }} />
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </section>
      )}

      <p className="text-xs text-gray-400">
        Note: Workspace bundles Google Photos into the same Drive quota — no separate "Photos" line is exposed.
        Reports API data lags 24–48 hours; "today's" usage is yesterday's snapshot.
      </p>
    </div>
  );
}

function Tile({
  label,
  value,
  accent,
}: {
  label: string;
  value: string;
  accent: 'blue' | 'green' | 'indigo' | 'amber';
}) {
  const colors: Record<string, string> = {
    blue: 'border-blue-200 bg-blue-50 text-blue-900',
    green: 'border-green-200 bg-green-50 text-green-900',
    indigo: 'border-indigo-200 bg-indigo-50 text-indigo-900',
    amber: 'border-amber-200 bg-amber-50 text-amber-900',
  };
  return (
    <div className={`border rounded-lg px-3 py-2.5 ${colors[accent]}`}>
      <div className="text-xl font-semibold">{value}</div>
      <div className="text-xs uppercase tracking-wide opacity-75 mt-0.5">{label}</div>
    </div>
  );
}

function countTrue(snap: UsageSnapshot, key: 'has_drive_presence' | 'has_gmail_presence'): number {
  let n = 0;
  for (const u of Object.values(snap.users)) {
    if (u[key]) n++;
  }
  return n;
}

function formatDate(iso?: string): string {
  if (!iso) return '—';
  const t = new Date(iso);
  if (t.getFullYear() < 1971 || t.getFullYear() === 1) return 'Never';
  return t.toLocaleDateString();
}
