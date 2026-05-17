document.querySelector("[data-toggle-dark]")?.addEventListener("click", () => {
  document.documentElement.classList.toggle("dark");
  localStorage.setItem("dark", document.documentElement.classList.contains("dark") ? "1" : "0");
});
if (localStorage.getItem("dark") === "1") document.documentElement.classList.add("dark");
document.querySelector("[data-toggle-sidebar]")?.addEventListener("click", () => {
  document.querySelector("#sidebar")?.classList.toggle("hidden");
});
document.querySelectorAll(".btn-primary").forEach((button) => {
  button.addEventListener("click", () => button.classList.add("opacity-80"));
});
