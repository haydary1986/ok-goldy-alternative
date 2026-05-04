import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import Modal from './Modal';
import { useToast } from './Toast';
import { apiPost } from '../lib/api';
import type { CreateUserRequest, User } from '../lib/types';

interface Props {
  open: boolean;
  onClose: () => void;
}

export default function CreateUserModal({ open, onClose }: Props) {
  const qc = useQueryClient();
  const toast = useToast();
  const [form, setForm] = useState<CreateUserRequest>({
    primary_email: '',
    given_name: '',
    family_name: '',
    password: '',
    org_unit_path: '',
  });
  const set = (k: keyof CreateUserRequest) => (v: string) => setForm((s) => ({ ...s, [k]: v }));

  const create = useMutation({
    mutationFn: () => apiPost<User>('/users', form),
    onSuccess: (u) => {
      toast.success(`Created ${u.primary_email}.`);
      qc.invalidateQueries({ queryKey: ['users'] });
      setForm({ primary_email: '', given_name: '', family_name: '', password: '', org_unit_path: '' });
      onClose();
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const valid =
    form.primary_email && form.given_name && form.family_name && form.password.length >= 8;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Create user"
      footer={
        <div className="flex justify-end gap-2">
          <button
            onClick={onClose}
            disabled={create.isPending}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded text-gray-700 hover:bg-gray-50 disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            onClick={() => create.mutate()}
            disabled={!valid || create.isPending}
            className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {create.isPending ? 'Creating…' : 'Create user'}
          </button>
        </div>
      }
    >
      <div className="space-y-3">
        <Field label="Primary email" required>
          <input
            type="email"
            value={form.primary_email}
            onChange={(e) => set('primary_email')(e.target.value)}
            placeholder="firstname.lastname@yourdomain.com"
            className="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-blue-500"
            required
          />
        </Field>
        <div className="grid grid-cols-2 gap-3">
          <Field label="Given name" required>
            <input
              value={form.given_name}
              onChange={(e) => set('given_name')(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-blue-500"
              required
            />
          </Field>
          <Field label="Family name" required>
            <input
              value={form.family_name}
              onChange={(e) => set('family_name')(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-blue-500"
              required
            />
          </Field>
        </div>
        <Field label="Initial password" required hint="Min 8 chars. User will be required to change on first sign-in.">
          <input
            type="text"
            value={form.password}
            onChange={(e) => set('password')(e.target.value)}
            className="w-full font-mono text-sm border border-gray-300 rounded px-3 py-1.5 focus:outline-none focus:border-blue-500"
            required
            minLength={8}
          />
          <button
            type="button"
            onClick={() => set('password')(genPassword())}
            className="mt-1 text-xs text-blue-600 hover:underline"
          >
            Generate strong password
          </button>
        </Field>
        <Field label="Organisational unit" hint="Defaults to / (root)">
          <input
            value={form.org_unit_path ?? ''}
            onChange={(e) => set('org_unit_path')(e.target.value)}
            placeholder="/"
            className="w-full font-mono text-xs border border-gray-300 rounded px-3 py-1.5 focus:outline-none focus:border-blue-500"
          />
        </Field>
      </div>
    </Modal>
  );
}

function Field({
  label,
  required,
  hint,
  children,
}: {
  label: string;
  required?: boolean;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">
        {label} {required && <span className="text-red-500">*</span>}
      </label>
      {children}
      {hint && <div className="text-xs text-gray-500 mt-0.5">{hint}</div>}
    </div>
  );
}

function genPassword(): string {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789!@#$%';
  return Array.from(crypto.getRandomValues(new Uint32Array(16)))
    .map((n) => chars[n % chars.length])
    .join('');
}
