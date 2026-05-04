import { Outlet, Link, useLocation } from 'react-router-dom';
import { useState, useEffect } from 'react';
import { getActor, setActor } from '../lib/api';

const navItems = [
  { to: '/', label: 'Dashboard' },
  { to: '/users', label: 'Users' },
  { to: '/groups', label: 'Groups' },
  { to: '/audit', label: 'Audit log' },
  { to: '/settings', label: 'Settings' },
];

export default function Layout() {
  const { pathname } = useLocation();
  const [actor, setActorState] = useState(() => getActor());

  useEffect(() => {
    setActor(actor);
  }, [actor]);

  return (
    <div className="min-h-screen flex flex-col">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between gap-4">
          <Link to="/" className="text-xl font-semibold text-gray-900 whitespace-nowrap">
            Ok Goldy <span className="text-gray-400">Alternative</span>
          </Link>
          <input
            type="email"
            placeholder="acting-as: admin@example.com"
            value={actor}
            onChange={(e) => setActorState(e.target.value)}
            className="border border-gray-300 rounded px-3 py-1.5 text-sm w-72 focus:outline-none focus:border-blue-500"
            aria-label="Acting as"
          />
        </div>
        <nav className="max-w-7xl mx-auto px-4 flex gap-2 -mb-px">
          {navItems.map((item) => {
            const active = pathname === item.to || (item.to !== '/' && pathname.startsWith(item.to));
            return (
              <Link
                key={item.to}
                to={item.to}
                className={`px-3 py-2 text-sm border-b-2 transition ${
                  active
                    ? 'border-blue-600 text-blue-700'
                    : 'border-transparent text-gray-600 hover:text-gray-900'
                }`}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>
      </header>
      <main className="flex-1 max-w-7xl mx-auto w-full px-4 py-6">
        <Outlet />
      </main>
      <footer className="border-t border-gray-200 bg-white">
        <div className="max-w-7xl mx-auto px-4 py-3 text-xs text-gray-500 flex justify-between">
          <span>Ok Goldy Alternative</span>
          <a
            href="https://github.com/haydary1986/ok-goldy-alternative"
            className="hover:text-gray-700"
          >
            GitHub
          </a>
        </div>
      </footer>
    </div>
  );
}
