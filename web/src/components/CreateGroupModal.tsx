import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import Modal from './Modal';
import { useToast } from './Toast';
import { apiPost } from '../lib/api';
import type { CreateGroupRequest, Group } from '../lib/types';

interface Props {
  open: boolean;
  onClose: () => void;
}

export default function CreateGroupModal({ open, onClose }: Props) {
  const qc = useQueryClient();
  const toast = useToast();
  const [form, setForm] = useState<CreateGroupRequest>({ email: '', name: '', description: '' });
  const set = (k: keyof CreateGroupRequest) => (v: string) => setForm((s) => ({ ...s, [k]: v }));

  const create = useMutation({
    mutationFn: () => apiPost<Group>('/groups', form),
    onSuccess: (g) => {
      toast.success(`Created ${g.email}.`);
      qc.invalidateQueries({ queryKey: ['groups'] });
      setForm({ email: '', name: '', description: '' });
      onClose();
    },
    onError: (err: Error) => toast.error(err.message),
  });

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Create group"
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
            disabled={!form.email || create.isPending}
            className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {create.isPending ? 'Creating…' : 'Create group'}
          </button>
        </div>
      }
    >
      <div className="space-y-3">
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">
            Group email <span className="text-red-500">*</span>
          </label>
          <input
            type="email"
            value={form.email}
            onChange={(e) => set('email')(e.target.value)}
            placeholder="team@yourdomain.com"
            className="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-blue-500"
            required
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">Display name</label>
          <input
            value={form.name ?? ''}
            onChange={(e) => set('name')(e.target.value)}
            placeholder="Team name (optional)"
            className="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-blue-500"
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">Description</label>
          <textarea
            rows={3}
            value={form.description ?? ''}
            onChange={(e) => set('description')(e.target.value)}
            className="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:border-blue-500"
          />
        </div>
      </div>
    </Modal>
  );
}
