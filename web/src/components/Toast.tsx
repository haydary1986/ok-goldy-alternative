import { createContext, useCallback, useContext, useState, type ReactNode } from 'react';

type ToastKind = 'success' | 'error' | 'info';
interface Toast {
  id: number;
  kind: ToastKind;
  msg: string;
}

const ToastCtx = createContext<{
  push: (kind: ToastKind, msg: string) => void;
}>({ push: () => {} });

export function useToast() {
  const ctx = useContext(ToastCtx);
  return {
    success: (msg: string) => ctx.push('success', msg),
    error: (msg: string) => ctx.push('error', msg),
    info: (msg: string) => ctx.push('info', msg),
  };
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const push = useCallback((kind: ToastKind, msg: string) => {
    const id = Date.now() + Math.random();
    setToasts((cur) => [...cur, { id, kind, msg }]);
    setTimeout(() => {
      setToasts((cur) => cur.filter((t) => t.id !== id));
    }, 5000);
  }, []);

  return (
    <ToastCtx.Provider value={{ push }}>
      {children}
      <div className="fixed bottom-4 right-4 z-[60] flex flex-col gap-2 max-w-sm">
        {toasts.map((t) => (
          <div
            key={t.id}
            className={`px-4 py-2.5 rounded shadow-lg text-sm border flex items-start gap-2 transition-all bg-white ${
              t.kind === 'success'
                ? 'border-green-300 text-green-900'
                : t.kind === 'error'
                ? 'border-red-300 text-red-900'
                : 'border-blue-300 text-blue-900'
            }`}
          >
            <span className="mt-0.5">
              {t.kind === 'success' ? '✓' : t.kind === 'error' ? '⚠' : 'ℹ'}
            </span>
            <span className="flex-1 break-words">{t.msg}</span>
            <button
              onClick={() => setToasts((cur) => cur.filter((x) => x.id !== t.id))}
              className="text-gray-400 hover:text-gray-600 text-xs"
              aria-label="Dismiss"
            >
              ×
            </button>
          </div>
        ))}
      </div>
    </ToastCtx.Provider>
  );
}
