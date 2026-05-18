document.querySelector("[data-toggle-dark]")?.addEventListener("click", () => {
  document.documentElement.classList.toggle("dark");
  localStorage.setItem("dark", document.documentElement.classList.contains("dark") ? "1" : "0");
});
if (localStorage.getItem("dark") === "1") document.documentElement.classList.add("dark");
document.querySelector("[data-toggle-sidebar]")?.addEventListener("click", () => {
  document.querySelector("#sidebar")?.classList.toggle("hidden");
});
document.querySelectorAll(".btn-primary").forEach((button) => {
  button.addEventListener("click", () => {
    button.classList.add("opacity-80");
    if (button.closest("form")) button.setAttribute("aria-busy", "true");
  });
});

const offlineBadge = document.createElement("div");
offlineBadge.className = "offline-indicator";
offlineBadge.textContent = "Offline";
document.body.appendChild(offlineBadge);
function syncOfflineState() {
  offlineBadge.classList.toggle("is-visible", !navigator.onLine);
}
window.addEventListener("online", syncOfflineState);
window.addEventListener("offline", syncOfflineState);
syncOfflineState();

document.addEventListener("keydown", (event) => {
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "k") {
    const search = document.querySelector("[data-quick-search], input[type='search'], input[name='q']");
    if (search) {
      event.preventDefault();
      search.focus();
      search.select?.();
    }
  }
});
