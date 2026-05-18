export default function KpiCard({ label, value }) {
  return <div className="rounded-lg border bg-white p-5 shadow-sm"><p className="text-sm text-slate-500">{label}</p><strong className="mt-2 block text-3xl">{value}</strong></div>;
}
