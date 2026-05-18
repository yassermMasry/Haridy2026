const tbody = document.querySelector("#invoice-table tbody");
const tpl = document.querySelector("#row-template");
const addBtn = document.querySelector("#add-row");
const barcodeInput = document.querySelector("#barcode-input");

function addRow(matchBarcode = "") {
  const row = tpl.content.cloneNode(true);
  tbody.appendChild(row);
  const tr = tbody.lastElementChild;
  const select = tr.querySelector(".item-select");
  if (matchBarcode) {
    [...select.options].forEach((option) => {
      if (option.dataset.barcode === matchBarcode) select.value = option.value;
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
}

addBtn?.addEventListener("click", () => addRow());
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
    addRow(barcodeInput.value.trim());
    barcodeInput.value = "";
  }
});
document.addEventListener("input", (event) => {
  if (event.target.id === "paid_cash") event.target.dataset.touched = "1";
  if (event.target.classList.contains("item-select")) {
    const tr = event.target.closest("tr");
    tr.querySelector(".price").value = selectedPrice(event.target);
  }
  calculate();
});
document.addEventListener("click", (event) => {
  if (event.target.classList.contains("remove-row")) {
    event.target.closest("tr").remove();
    calculate();
  }
});
addRow();
