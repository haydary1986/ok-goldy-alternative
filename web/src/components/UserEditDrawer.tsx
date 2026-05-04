import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import Drawer from './Drawer';
import { useToast } from './Toast';
import { api, apiDelete, apiPatch, apiPost } from '../lib/api';
import type { Alias, UpdateUserRequest, User } from '../lib/types';

interface Props {
  user: User | null;
  onClose: () => void;
  onDeleted?: () => void;
}

export default function UserEditDrawer({ user, onClose, onDeleted }: Props) {
  const qc = useQueryClient();
  const toast = useToast();
  const [givenName, setGivenName] = useState('');
  const [familyName, setFamilyName] = useState('');
  const [orgUnit, setOrgUnit] = useState('');
  const [suspended, setSuspended] = useState(false);

  useEffect(() => {
    if (!user) return;
    setGivenName(user.given_name ?? '');
    setFamilyName(user.family_name ?? '');
    setOrgUnit(user.org_unit_path ?? '');
    setSuspended(user.suspended);
  }, [user]);

  const aliases = useQuery({
    queryKey: ['user-aliases', user?.id],
    queryFn: () => api<{ aliases: Alias[] }>(`/users/${user!.id}/aliases`),
    enabled: !!user,
  });

  const save = useMutation({
    mutationFn: () => {
      if (!user) throw new Error('no user');
      const patch: UpdateUserRequest = {};
      if (givenName !== (user.given_name ?? '')) patch.given_name = givenName;
      if (familyName !== (user.family_name ?? '')) patch.family_name = familyName;
      if (orgUnit !== (user.org_unit_path ?? '')) patch.org_unit_path = orgUnit;
      if (suspended !== user.suspended) patch.suspended = suspended;
      if (Object.keys(patch).length === 0) {
        return Promise.resolve(user);
      }
      return apiPatch<User>(`/users/${user.id}`, patch);
    },
    onSuccess: () => {
      toast.success('User updated.');
      qc.invalidateQueries({ queryKey: ['users'] });
      onClose();
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const del = useMutation({
    mutationFn: () => apiDelete(`/users/${user!.id}`),
    onSuccess: () => {
      toast.success('User deleted.');
      qc.invalidateQueries({ queryKey: ['users'] });
      onDeleted?.();
      onClose();
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const toggleSuspend = useMutation({
    mutationFn: () => {
      const next = !user!.suspended;
      return apiPatch<User>(`/users/${user!.id}`, { suspended: next } as UpdateUserRequest);
    },
    onSuccess: (_data, _vars) => {
      toast.success(user!.suspended ? 'User restored.' : 'User suspended.');
      qc.invalidateQueries({ queryKey: ['users'] });
      onClose();
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const [newAlias, setNewAlias] = useState('');
  const addAlias = useMutation({
    mutationFn: () => apiPost<Alias>(`/users/${user!.id}/aliases`, { alias: newAlias }),
    onSuccess: () => {
      setNewAlias('');
      qc.invalidateQueries({ queryKey: ['user-aliases', user?.id] });
      toast.success('Alias added.');
    },
    onError: (err: Error) => toast.error(err.message),
  });
  const removeAlias = useMutation({
    mutationFn: (alias: string) =>
      apiDelete(`/users/${user!.id}/aliases/${encodeURIComponent(alias)}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['user-aliases', user?.id] });
      toast.success('Alias removed.');
    },
    onError: (err: Error) => toast.error(err.message),
  });

  return (
    <Drawer
      open={!!user}
      onClose={onClose}
      title={user?.primary_email ?? ''}
      subtitle={user ? `${user.id} • ${user.org_unit_path ?? '/'}` : ''}
      footer={
        <div className="flex flex-wrap items-center gap-2">
          <button
            onClick={() => save.mutate()}
            disabled={save.isPending}
            className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {save.isPending ? 'Saving…' : 'Save changes'}
          </button>
          <button
            onClick={() => toggleSuspend.mutate()}
            disabled={toggleSuspend.isPending}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50"
          >
            {user?.suspended ? 'Restore' : 'Suspend'}
          </button>
          <div className="ml-auto" />
          <button
            onClick={() => {
              if (
                user &&
                confirm(
                  `Permanently delete ${user.primary_email}? Workspace will release the username after a delay.`,
                )
              ) {
                del.mutate();
              }
            }}
            disabled={del.isPending}
            className="px-3 py-1.5 text-sm border border-red-300 text-red-700 rounded hover:bg-red-50 disabled:opacity-50"
          >
            {del.isPending ? 'Deleting…' : 'Delete user'}
          </button>
        </div>
      }
    >
      {user && (
        <div className="space-y-5">
          <Section title="Profile">
            <Field label="Given name">
              <input
                value={givenName}
                onChange={(e) => setGivenName(e.target.value)}
                className="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-blue-500"
              />
            </Field>
            <Field label="Family name">
              <input
                value={familyName}
                onChange={(e) => setFamilyName(e.target.value)}
                className="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-blue-500"
              />
            </Field>
            <Field label="Organisational unit">
              <input
                value={orgUnit}
                onChange={(e) => setOrgUnit(e.target.value)}
                placeholder="/"
                className="w-full font-mono text-xs border border-gray-300 rounded px-3 py-1.5 focus:outline-none focus:border-blue-500"
              />
            </Field>
            <Field label="Suspended">
              <label className="inline-flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={suspended}
                  onChange={(e) => setSuspended(e.target.checked)}
                  className="rounded"
                />
                <span>{suspended ? 'User is currently suspended' : 'User is active'}</span>
              </label>
            </Field>
            <Field label="Admin role">
              <span className={`text-xs ${user.is_admin ? 'text-green-700' : 'text-gray-500'}`}>
                {user.is_admin ? 'Super-admin (managed in Workspace Admin)' : 'Standard user'}
              </span>
            </Field>
          </Section>

          <Section title="Aliases">
            {aliases.isLoading && <div className="text-xs text-gray-500">Loading aliases…</div>}
            {aliases.isError && (
              <div className="text-xs text-red-600">Failed to load aliases.</div>
            )}
            {aliases.data && (
              <ul className="space-y-1">
                {aliases.data.aliases.length === 0 ? (
                  <li className="text-xs text-gray-500">No aliases.</li>
                ) : (
                  aliases.data.aliases.map((a) => (
                    <li key={a.alias} className="flex items-center justify-between text-sm">
                      <code className="font-mono text-xs">{a.alias}</code>
                      <button
                        onClick={() => {
                          if (confirm(`Remove alias ${a.alias}?`)) removeAlias.mutate(a.alias);
                        }}
                        disabled={removeAlias.isPending}
                        className="text-red-600 hover:underline text-xs"
                      >
                        Remove
                      </button>
                    </li>
                  ))
                )}
              </ul>
            )}
            <div className="mt-2 flex gap-2">
              <input
                value={newAlias}
                onChange={(e) => setNewAlias(e.target.value)}
                placeholder="new.alias@yourdomain.com"
                className="flex-1 border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-blue-500"
              />
              <button
                onClick={() => addAlias.mutate()}
                disabled={!newAlias || addAlias.isPending}
                className="px-3 py-1.5 text-sm bg-gray-800 text-white rounded hover:bg-gray-900 disabled:opacity-50"
              >
                Add
              </button>
            </div>
          </Section>
        </div>
      )}
    </Drawer>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section>
      <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">
        {title}
      </h3>
      <div className="space-y-3">{children}</div>
    </section>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">{label}</label>
      {children}
    </div>
  );
}
