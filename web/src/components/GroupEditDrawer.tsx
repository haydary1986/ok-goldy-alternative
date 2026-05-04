import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import Drawer from './Drawer';
import { useToast } from './Toast';
import { api, apiDelete, apiPost } from '../lib/api';
import type { Group, Member, MembersListResponse } from '../lib/types';

interface Props {
  group: Group | null;
  onClose: () => void;
}

export default function GroupEditDrawer({ group, onClose }: Props) {
  const qc = useQueryClient();
  const toast = useToast();

  useEffect(() => {
    if (!group) return;
  }, [group]);

  const members = useQuery({
    queryKey: ['group-members', group?.id],
    queryFn: () =>
      api<MembersListResponse>(`/groups/${encodeURIComponent(group!.id)}/members?page_size=200`),
    enabled: !!group,
  });

  const [newEmail, setNewEmail] = useState('');
  const [newRole, setNewRole] = useState<'OWNER' | 'MANAGER' | 'MEMBER'>('MEMBER');

  const addMember = useMutation({
    mutationFn: () =>
      apiPost<Member>(`/groups/${encodeURIComponent(group!.id)}/members`, {
        email: newEmail,
        role: newRole,
      }),
    onSuccess: () => {
      setNewEmail('');
      qc.invalidateQueries({ queryKey: ['group-members', group?.id] });
      toast.success('Member added.');
    },
    onError: (err: Error) => toast.error(err.message),
  });
  const removeMember = useMutation({
    mutationFn: (key: string) =>
      apiDelete(`/groups/${encodeURIComponent(group!.id)}/members/${encodeURIComponent(key)}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['group-members', group?.id] });
      toast.success('Member removed.');
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const del = useMutation({
    mutationFn: () => apiDelete(`/groups/${encodeURIComponent(group!.id)}`),
    onSuccess: () => {
      toast.success('Group deleted.');
      qc.invalidateQueries({ queryKey: ['groups'] });
      onClose();
    },
    onError: (err: Error) => toast.error(err.message),
  });

  return (
    <Drawer
      open={!!group}
      onClose={onClose}
      title={group?.email ?? ''}
      subtitle={group ? group.name || group.id : ''}
      width="max-w-2xl"
      footer={
        <div className="flex items-center gap-2">
          <button
            onClick={() => {
              if (
                group &&
                confirm(
                  `Permanently delete group ${group.email}? Members are not affected — only the group itself.`,
                )
              ) {
                del.mutate();
              }
            }}
            disabled={del.isPending}
            className="ml-auto px-3 py-1.5 text-sm border border-red-300 text-red-700 rounded hover:bg-red-50 disabled:opacity-50"
          >
            {del.isPending ? 'Deleting…' : 'Delete group'}
          </button>
        </div>
      }
    >
      {group && (
        <div className="space-y-5">
          <section>
            <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">
              Group
            </h3>
            <dl className="grid grid-cols-1 sm:grid-cols-2 gap-y-1 text-sm text-gray-700">
              <div>
                <dt className="inline font-medium">Email: </dt>
                <dd className="inline font-mono text-xs">{group.email}</dd>
              </div>
              {group.name && (
                <div>
                  <dt className="inline font-medium">Name: </dt>
                  <dd className="inline">{group.name}</dd>
                </div>
              )}
              {group.description && (
                <div className="col-span-full">
                  <dt className="inline font-medium">Description: </dt>
                  <dd className="inline">{group.description}</dd>
                </div>
              )}
              <div>
                <dt className="inline font-medium">Direct members: </dt>
                <dd className="inline">{group.direct_members_count ?? 0}</dd>
              </div>
            </dl>
          </section>

          <section>
            <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">
              Members
            </h3>

            <div className="flex gap-2 mb-3">
              <input
                value={newEmail}
                onChange={(e) => setNewEmail(e.target.value)}
                placeholder="email@yourdomain.com"
                className="flex-1 border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-blue-500"
              />
              <select
                value={newRole}
                onChange={(e) => setNewRole(e.target.value as 'OWNER' | 'MANAGER' | 'MEMBER')}
                className="border border-gray-300 rounded px-3 py-1.5 text-sm bg-white"
              >
                <option value="MEMBER">Member</option>
                <option value="MANAGER">Manager</option>
                <option value="OWNER">Owner</option>
              </select>
              <button
                onClick={() => addMember.mutate()}
                disabled={!newEmail || addMember.isPending}
                className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
              >
                Add
              </button>
            </div>

            {members.isLoading && <div className="text-xs text-gray-500">Loading members…</div>}
            {members.isError && (
              <div className="text-xs text-red-600">Failed to load members.</div>
            )}
            {members.data && (
              <div className="border border-gray-200 rounded">
                <table className="w-full text-sm">
                  <thead className="bg-gray-50 text-left">
                    <tr>
                      <th className="px-3 py-1.5 font-medium text-gray-700">Email</th>
                      <th className="px-3 py-1.5 font-medium text-gray-700">Role</th>
                      <th className="px-3 py-1.5"></th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    {members.data.members.length === 0 ? (
                      <tr>
                        <td colSpan={3} className="px-3 py-4 text-center text-gray-500">
                          No members.
                        </td>
                      </tr>
                    ) : (
                      members.data.members.map((m) => (
                        <tr key={m.id ?? m.email}>
                          <td className="px-3 py-1.5 font-mono text-xs">{m.email}</td>
                          <td className="px-3 py-1.5 text-xs text-gray-600">{m.role ?? '—'}</td>
                          <td className="px-3 py-1.5 text-right">
                            <button
                              onClick={() => {
                                if (confirm(`Remove ${m.email} from ${group.email}?`)) {
                                  removeMember.mutate(m.id ?? m.email);
                                }
                              }}
                              disabled={removeMember.isPending}
                              className="text-red-600 hover:underline text-xs"
                            >
                              Remove
                            </button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            )}
          </section>
        </div>
      )}
    </Drawer>
  );
}
