import { Link } from 'react-router-dom';

export default function Dashboard() {
  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-semibold">Dashboard</h1>
        <p className="text-gray-600 mt-1">
          Welcome to Ok Goldy Alternative — bulk Google Workspace administration without the spreadsheet.
        </p>
      </header>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card title="Users" description="Bulk create / update / suspend / delete" to="/users" />
        <Card title="Groups" description="Manage groups and members in bulk" to="/groups" />
        <Card title="Audit log" description="Every mutation, recorded" to="/audit" />
        <Card title="Settings" description="Upload service-account JSON to enable Workspace API" to="/settings" />
      </div>

      <section className="bg-white border border-gray-200 rounded-lg p-4 text-sm text-gray-700 space-y-2">
        <h2 className="font-medium text-gray-900">Setup checklist</h2>
        <ol className="list-decimal list-inside space-y-1 text-gray-600">
          <li>Create a Google Cloud Service Account with Domain-Wide Delegation.</li>
          <li>Authorize the required Admin SDK scopes (see README).</li>
          <li>
            Mount the service-account JSON at <code className="bg-gray-100 px-1 rounded">/secrets/service-account.json</code>.
          </li>
          <li>
            Set <code className="bg-gray-100 px-1 rounded">GOLDY_GOOGLE_DELEGATED_ADMIN</code> to a super-admin email.
          </li>
          <li>Restart the server — the Workspace endpoints will come online.</li>
        </ol>
      </section>
    </div>
  );
}

function Card({ title, description, to }: { title: string; description: string; to: string }) {
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
