export default function Audit() {
  return (
    <div className="space-y-4">
      <header>
        <h1 className="text-2xl font-semibold">Audit log</h1>
        <p className="text-gray-600 mt-1">
          Every mutation Goldy performs is recorded with actor, action, target, and a JSON
          before/after snapshot.
        </p>
      </header>

      <div className="rounded border border-amber-300 bg-amber-50 p-4 text-sm text-amber-900">
        <p className="font-medium">Audit endpoint not yet wired</p>
        <p className="mt-1 text-amber-800">
          The <code className="bg-amber-100 px-1 rounded">audit_log</code> table is being populated by
          every Users / Groups mutation, but the read API
          (<code className="bg-amber-100 px-1 rounded">GET /api/v1/audit</code>) currently returns 501.
          Wiring it is part of the next milestone.
        </p>
      </div>
    </div>
  );
}
