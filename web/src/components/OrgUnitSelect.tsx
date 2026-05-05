import { useEffect, useMemo, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, apiPost } from '../lib/api';
import { useToast } from './Toast';
import type { CreateOrgUnitRequest, OrgUnit, OrgUnitsListResponse } from '../lib/types';

interface Props {
  value: string;
  onChange: (path: string) => void;
  placeholder?: string;
}

// OrgUnitSelect renders a combobox-like dropdown for choosing a Workspace
// organizational unit by path, plus an inline "+ Create new OU" flow that
// posts to /api/v1/orgunits and selects the freshly-created path.
export default function OrgUnitSelect({ value, onChange, placeholder = '/' }: Props) {
  const qc = useQueryClient();
  const toast = useToast();
  const [open, setOpen] = useState(false);
  const [filter, setFilter] = useState('');
  const [creating, setCreating] = useState(false);
  const wrapperRef = useRef<HTMLDivElement | null>(null);

  const list = useQuery({
    queryKey: ['orgunits'],
    queryFn: () => api<OrgUnitsListResponse>('/orgunits'),
    staleTime: 60_000,
  });

  // Close on outside click.
  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        setOpen(false);
        setCreating(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open]);

  const filtered = useMemo<OrgUnit[]>(() => {
    if (!list.data) return [];
    const f = filter.trim().toLowerCase();
    if (!f) return list.data.org_units;
    return list.data.org_units.filter((ou) =>
      [ou.org_unit_path, ou.name, ou.description].filter(Boolean).join(' ').toLowerCase().includes(f),
    );
  }, [list.data, filter]);

  const exactMatch = useMemo(() => {
    if (!list.data || !filter) return true;
    return list.data.org_units.some((ou) => ou.org_unit_path === filter);
  }, [list.data, filter]);

  return (
    <div ref={wrapperRef} className="relative">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="w-full text-left border border-gray-300 rounded px-3 py-1.5 text-sm bg-white hover:border-gray-400 focus:outline-none focus:border-blue-500 flex items-center justify-between"
      >
        <span className="font-mono text-xs truncate">{value || placeholder}</span>
        <span className="text-gray-400 text-xs ml-2">▼</span>
      </button>

      {open && (
        <div className="absolute z-30 mt-1 w-full bg-white border border-gray-300 rounded shadow-lg max-h-72 flex flex-col">
          <input
            autoFocus
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            placeholder="Type to filter or enter a new OU path…"
            className="px-3 py-1.5 border-b border-gray-200 text-sm font-mono focus:outline-none"
          />

          <div className="overflow-y-auto flex-1">
            {list.isLoading && <div className="px-3 py-2 text-xs text-gray-500">Loading OUs…</div>}
            {list.isError && (
              <div className="px-3 py-2 text-xs text-red-600 break-words">
                {(list.error as Error).message}
              </div>
            )}
            {filtered.map((ou) => (
              <button
                key={ou.org_unit_path}
                type="button"
                onClick={() => {
                  onChange(ou.org_unit_path);
                  setOpen(false);
                  setFilter('');
                }}
                className={`w-full text-left px-3 py-1.5 text-sm hover:bg-blue-50 flex items-center justify-between ${
                  ou.org_unit_path === value ? 'bg-blue-100' : ''
                }`}
              >
                <span className="font-mono text-xs truncate">{ou.org_unit_path}</span>
                {ou.name && ou.org_unit_path !== '/' && (
                  <span className="text-xs text-gray-500 ml-2 truncate">{ou.name}</span>
                )}
              </button>
            ))}
            {filtered.length === 0 && !list.isLoading && (
              <div className="px-3 py-2 text-xs text-gray-500">No matching OUs.</div>
            )}
          </div>

          <div className="border-t border-gray-200 p-2 bg-gray-50">
            {!creating ? (
              <button
                type="button"
                onClick={() => setCreating(true)}
                className="w-full text-left text-sm text-blue-700 hover:bg-blue-50 rounded px-2 py-1"
              >
                + Create new organizational unit{!exactMatch && filter ? ` "${filter}"` : ''}
              </button>
            ) : (
              <CreateOUInline
                initialPath={!exactMatch ? filter : ''}
                existing={list.data?.org_units ?? []}
                onCancel={() => setCreating(false)}
                onCreated={(path) => {
                  qc.invalidateQueries({ queryKey: ['orgunits'] });
                  toast.success(`Created ${path}.`);
                  onChange(path);
                  setOpen(false);
                  setCreating(false);
                  setFilter('');
                }}
                onError={(msg) => toast.error(msg)}
              />
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function CreateOUInline({
  initialPath,
  existing,
  onCancel,
  onCreated,
  onError,
}: {
  initialPath: string;
  existing: OrgUnit[];
  onCancel: () => void;
  onCreated: (path: string) => void;
  onError: (msg: string) => void;
}) {
  // Pre-split the initial path into name + parent.
  const initial = splitPath(initialPath || '/New');
  const [name, setName] = useState(initial.name);
  const [parent, setParent] = useState(initial.parent);
  const [description, setDescription] = useState('');

  const create = useMutation({
    mutationFn: () =>
      apiPost<OrgUnit>('/orgunits', {
        name,
        parent_org_unit_path: parent,
        description: description || undefined,
      } as CreateOrgUnitRequest),
    onSuccess: (ou) => onCreated(ou.org_unit_path),
    onError: (err: Error) => onError(err.message),
  });

  const parents = existing
    .map((o) => o.org_unit_path)
    .sort((a, b) => a.localeCompare(b));

  return (
    <div className="space-y-2">
      <div className="grid grid-cols-2 gap-2">
        <label className="block text-xs">
          <span className="text-gray-600 font-medium">Name</span>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="leaf-name"
            className="mt-0.5 w-full border border-gray-300 rounded px-2 py-1 text-sm focus:outline-none focus:border-blue-500"
          />
        </label>
        <label className="block text-xs">
          <span className="text-gray-600 font-medium">Parent</span>
          <select
            value={parent}
            onChange={(e) => setParent(e.target.value)}
            className="mt-0.5 w-full border border-gray-300 rounded px-2 py-1 text-sm bg-white focus:outline-none focus:border-blue-500 font-mono text-xs"
          >
            {parents.length === 0 && <option value="/">/</option>}
            {parents.map((p) => (
              <option key={p} value={p}>
                {p}
              </option>
            ))}
          </select>
        </label>
      </div>
      <label className="block text-xs">
        <span className="text-gray-600 font-medium">Description (optional)</span>
        <input
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          className="mt-0.5 w-full border border-gray-300 rounded px-2 py-1 text-sm focus:outline-none focus:border-blue-500"
        />
      </label>
      <div className="flex justify-between items-center pt-1">
        <span className="text-xs text-gray-500 font-mono">
          → {parent === '/' ? '' : parent}/{name}
        </span>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={onCancel}
            disabled={create.isPending}
            className="px-2 py-1 text-xs border border-gray-300 rounded hover:bg-gray-100"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={() => create.mutate()}
            disabled={!name || create.isPending}
            className="px-2 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {create.isPending ? 'Creating…' : 'Create OU'}
          </button>
        </div>
      </div>
    </div>
  );
}

function splitPath(path: string): { parent: string; name: string } {
  if (!path || path === '/') return { parent: '/', name: '' };
  const idx = path.lastIndexOf('/');
  if (idx <= 0) return { parent: '/', name: path.replace(/^\//, '') };
  return { parent: path.slice(0, idx) || '/', name: path.slice(idx + 1) };
}
