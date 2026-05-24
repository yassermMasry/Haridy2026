const tbody = document.querySelector("#invoice-table tbody");
const tpl = document.querySelector("#row-template");
const addBtn = document.querySelector("#add-row");
const barcodeInput = document.querySelector("#barcode-input");
const itemSearchResults = document.querySelector("#item-search-results");
const warehouseFilter = document.querySelector("#warehouse-filter");

function itemOptions() {
  const select = tpl.content.querySelector(".item-select");
  return [...(select?.options || [])];
}

function optionMatches(option, query) {
  const q = query.toLowerCase();
  return (option.dataset.name || "").toLowerCase().includes(q) ||
    (option.dataset.code || "").toLowerCase().includes(q) ||
    (option.dataset.barcode || "").toLowerCase().includes(q);
}

function exactOption(query) {
  const q = query.toLowerCase();
  return itemOptions().find((option) =>
    (option.dataset.code || "").toLowerCase() === q ||
    (option.dataset.barcode || "").toLowerCase() === q
  );
}

function addRow(matchValue = "") {
  const row = tpl.content.cloneNode(true);
  tbody.appendChild(row);
  const tr = tbody.lastElementChild;
  const select = tr.querySelector(".item-select");
  if (matchValue) {
    const query = matchValue.toLowerCase();
    [...select.options].forEach((option) => {
      if ((option.dataset.barcode || "").toLowerCase() === query || (option.dataset.code || "").toLowerCase() === query || option.value === matchValue) {
        select.value = option.value;
      }
    });
  }
  tr.querySelector(".price").value = selectedPrice(select);
  calculate();
  tr.querySelector(".qty")?.focus();
}

function selectedPrice(select) {
  return select.selectedOptions[0]?.dataset.price || "0";
}

function calculate() {
  let subtotal = 0;
  tbody.querySelectorAll("tr").forEach((tr) => {
    const qty = parseFloat(tr.querySelector(".qty").value || "0");
    const price = parseFloat(tr.querySelector(".price").value || "0");
    const total = qty * price;
    subtotal += total;
    tr.querySelector(".line-total").textContent = total.toFixed(2);
  });
  const discount = parseFloat(document.querySelector("#discount").value || "0");
  const tax = parseFloat(document.querySelector("#tax").value || "0");
  const grand = subtotal - discount + tax;
  document.querySelector("#grand-total").textContent = grand.toFixed(2);
  const paid = document.querySelector("#paid_cash");
  if (!paid.dataset.touched) paid.value = grand.toFixed(2);
  const paidValue = parseFloat(paid.value || "0");
  const remaining = grand - paidValue;
  document.querySelector("#remaining-total").textContent = remaining.toFixed(2);
  const status = document.querySelector("#invoice-status");
  if (remaining === 0) status.textContent = "مسددة بالكامل";
  else if (remaining > 0) status.textContent = "آجل/باقي على العميل";
  else status.textContent = "زيادة مدفوعة/رصيد دائن للعميل";
}

addBtn?.addEventListener("click", () => addRow());
warehouseFilter?.addEventListener("change", () => {
  const url = new URL(window.location.href);
  if (warehouseFilter.value) url.searchParams.set("warehouse_id", warehouseFilter.value);
  else url.searchParams.delete("warehouse_id");
  window.location.href = url.toString();
});
function renderItemSearch(query) {
  if (!itemSearchResults) return;
  query = query.trim();
  if (!query) {
    itemSearchResults.classList.add("hidden");
    itemSearchResults.innerHTML = "";
    return;
  }
  const matches = itemOptions().filter((option) => optionMatches(option, query)).slice(0, 8);
  if (matches.length === 0) {
    itemSearchResults.innerHTML = '<div class="p-3 text-sm text-slate-500">لا توجد نتائج</div>';
    itemSearchResults.classList.remove("hidden");
    return;
  }
  itemSearchResults.innerHTML = matches.map((option) => `
    <button type="button" class="item-result block w-full border-b p-3 text-right text-sm hover:bg-slate-50" data-value="${option.value}">
      <div class="font-bold">${option.dataset.name || ""}</div>
      <div class="text-xs text-slate-500">الكود: ${option.dataset.code || "-"} | الباركود: ${option.dataset.barcode || "-"} | المتاح: ${option.dataset.quantity || "0"} | سعر البيع: ${option.dataset.price || "0"}</div>
    </button>
  `).join("");
  itemSearchResults.classList.remove("hidden");
}
document.addEventListener("keydown", (event) => {
  if (event.key === "F2") {
    event.preventDefault();
    addRow();
  }
  if (event.key === "F4") {
    event.preventDefault();
    barcodeInput?.focus();
  }
  if ((event.ctrlKey || event.metaKey) && event.key === "Enter") {
    event.preventDefault();
    event.target.closest("form")?.requestSubmit();
  }
});
barcodeInput?.addEventListener("keydown", (event) => {
  if (event.key === "Enter") {
    event.preventDefault();
    const value = barcodeInput.value.trim();
    const exact = exactOption(value);
    if (exact) addRow(exact.value);
    else addRow(value);
    barcodeInput.value = "";
    renderItemSearch("");
  }
});
document.addEventListener("input", (event) => {
  if (event.target.id === "paid_cash") event.target.dataset.touched = "1";
  if (event.target.id === "barcode-input") renderItemSearch(event.target.value);
  if (event.target.classList.contains("item-select")) {
    const tr = event.target.closest("tr");
    tr.querySelector(".price").value = selectedPrice(event.target);
  }
  calculate();
});
document.addEventListener("click", (event) => {
  if (event.target.closest(".item-result")) {
    addRow(event.target.closest(".item-result").dataset.value);
    barcodeInput.value = "";
    renderItemSearch("");
  }
  if (event.target.classList.contains("remove-row")) {
    event.target.closest("tr").remove();
    calculate();
  }
});
addRow();

const customerModal = document.querySelector("#customer-modal");
const openCustomerModal = document.querySelector("#open-customer-modal");
const closeCustomerModal = document.querySelector("#close-customer-modal");
const saveCustomerModal = document.querySelector("#save-customer-modal");
const customerSelect = document.querySelector("#customer-select");

function closeCustomer() {
  customerModal?.classList.add("hidden");
  customerModal?.classList.remove("flex");
  document.querySelector("#quick-customer-error")?.classList.add("hidden");
}

openCustomerModal?.addEventListener("click", () => {
  customerModal.classList.remove("hidden");
  customerModal.classList.add("flex");
  document.querySelector("#quick-customer-name")?.focus();
});
closeCustomerModal?.addEventListener("click", closeCustomer);
saveCustomerModal?.addEventListener("click", async () => {
  const error = document.querySelector("#quick-customer-error");
  error.classList.add("hidden");
  const csrf = document.querySelector('input[name="_csrf"]').value;
  const body = new URLSearchParams();
  body.set("_csrf", csrf);
  body.set("name", document.querySelector("#quick-customer-name").value);
  body.set("phone", document.querySelector("#quick-customer-phone").value);
  body.set("address", document.querySelector("#quick-customer-address").value);
  const response = await fetch("/customers/quick-create", {
    method: "POST",
    headers: {"Content-Type": "application/x-www-form-urlencoded", "X-CSRF-Token": csrf},
    body,
  });
  const data = await response.json();
  if (!response.ok) {
    error.textContent = data.error || "تعذر حفظ العميل";
    error.classList.remove("hidden");
    return;
  }
  customerSelect.add(new Option(data.name, data.id, true, true));
  closeCustomer();
});
