const API_BASE = import.meta.env.VITE_API_BASE || "/api/v1";

export async function api(path, options = {}) {
  const token = localStorage.getItem("token");
  const headers = { "Content-Type": "application/json", "X-Tenant": localStorage.getItem("tenant") || "demo", ...(options.headers || {}) };
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function login(username, password) {
  const data = await api("/auth/login", { method: "POST", body: JSON.stringify({ username, password }) });
  localStorage.setItem("token", data.token);
  return data;
}
