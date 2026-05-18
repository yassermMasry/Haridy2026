import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import { BarChart, Bar, ResponsiveContainer, XAxis, YAxis } from "recharts";
import { api, login } from "./api/client";
import KpiCard from "./components/KpiCard";
import "./style.css";

function App() {
  const [summary, setSummary] = useState(null);
  const [error, setError] = useState("");

  async function load() {
    try {
      if (!localStorage.getItem("token")) await login("admin", "admin123");
      const res = await api("/financial/statements");
      setSummary(res.data);
    } catch (err) {
      setError(err.message);
    }
  }

  useEffect(() => { load(); }, []);
  const chart = summary ? [{ name: "Sales", value: summary.sales }, { name: "Purchases", value: summary.purchases }] : [];

  return <main className="min-h-screen bg-slate-100 p-6">
    <header className="mb-6 flex items-center justify-between"><h1 className="text-3xl font-black">Haridy SaaS ERP</h1><span className="rounded-full bg-emerald-100 px-3 py-1 text-sm text-emerald-700">Flutter-ready API</span></header>
    {error && <div className="mb-4 rounded bg-red-50 p-3 text-red-700">{error}</div>}
    <section className="grid gap-4 md:grid-cols-4">
      <KpiCard label="Sales" value={summary?.sales?.toFixed?.(2) || "..."} />
      <KpiCard label="Purchases" value={summary?.purchases?.toFixed?.(2) || "..."} />
      <KpiCard label="Gross Profit" value={summary?.gross_profit?.toFixed?.(2) || "..."} />
      <KpiCard label="Cash" value={summary?.cash?.toFixed?.(2) || "..."} />
    </section>
    <section className="mt-6 rounded-lg border bg-white p-5">
      <h2 className="mb-4 text-xl font-bold">Financial Snapshot</h2>
      <ResponsiveContainer width="100%" height={260}><BarChart data={chart}><XAxis dataKey="name"/><YAxis/><Bar dataKey="value" fill="#059669"/></BarChart></ResponsiveContainer>
    </section>
  </main>;
}

createRoot(document.getElementById("root")).render(<App />);
