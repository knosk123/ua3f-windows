const presetEl = document.getElementById("preset");
const uaEl = document.getElementById("ua");
const ttlEl = document.getElementById("ttl");
const logEl = document.getElementById("log");
const statusCardEl = document.getElementById("statusCard");
const statusTextEl = document.getElementById("statusText");
const statusMetaEl = document.getElementById("statusMeta");
const packetCountEl = document.getElementById("packetCount");
const ttlCountEl = document.getElementById("ttlCount");
const uaCountEl = document.getElementById("uaCount");
const logsEl = document.getElementById("logs");
const ttlErrorEl = document.getElementById("ttlError");
const actionNoticeEl = document.getElementById("actionNotice");
const saveBtn = document.getElementById("saveBtn");
const startBtn = document.getElementById("startBtn");
const stopBtn = document.getElementById("stopBtn");
const exitBtn = document.getElementById("exitBtn");

const editableFields = [presetEl, ttlEl, logEl];
const presets = {
  wechat: "Mozilla/5.0 (Linux; Android 15; RMX6688 Build/AP3A.240617.008; wv) AppleWebKit/537.36",
  pc: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:149.0) Gecko/20100101 Firefox/149.0",
};

let currentStatus = "idle";
let formDirty = false;
let actionPending = false;
let exiting = false;
let pollTimer;

async function fetchJSON(url, options) {
  const response = await fetch(url, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || "请求失败");
  }
  return data;
}

function readForm() {
  return {
    preset: presetEl.value,
    ua: presets[presetEl.value] || presets.wechat,
    ttl: Number(ttlEl.value),
    log: logEl.value,
  };
}

function renderValidation(errors) {
  ttlErrorEl.textContent = errors.ttl || "";
  ttlEl.setAttribute("aria-invalid", errors.ttl ? "true" : "false");
}

function validateForm() {
  const errors = UA3FUI.validateConfig(readForm());
  renderValidation(errors);
  if (Object.keys(errors).length > 0) {
    throw new Error("请先修正标出的配置项");
  }
}

function applyConfig(config) {
  presetEl.value = presets[config.preset] ? config.preset : "wechat";
  uaEl.value = presets[presetEl.value] || presets.wechat;
  ttlEl.value = config.ttl || 64;
  logEl.value = config.log || "info";
  formDirty = false;
  renderValidation({});
  updateControls();
}

function applyStatus(snapshot) {
  currentStatus = snapshot.status;
  const presentation = UA3FUI.statusPresentation(snapshot);
  statusTextEl.textContent = presentation.label;
  statusMetaEl.textContent = presentation.meta;
  statusCardEl.dataset.status = presentation.tone;
  packetCountEl.textContent = snapshot.packetCount || 0;
  ttlCountEl.textContent = snapshot.ttlCount || 0;
  uaCountEl.textContent = snapshot.uaCount || 0;
  logsEl.textContent = (snapshot.logs || []).join("\n");
  logsEl.scrollTop = logsEl.scrollHeight;
  updateControls();
}

function updateControls() {
  const state = UA3FUI.controlState(currentStatus, formDirty, actionPending);
  editableFields.forEach((field) => {
    field.disabled = state.fieldsDisabled;
  });
  uaEl.disabled = state.fieldsDisabled;
  saveBtn.disabled = state.saveDisabled;
  saveBtn.textContent = state.saveLabel;
  startBtn.disabled = state.startDisabled;
  stopBtn.disabled = state.stopDisabled;
  exitBtn.disabled = state.exitDisabled;
}

function showNotice(message, tone = "") {
  actionNoticeEl.textContent = message;
  actionNoticeEl.dataset.tone = tone;
}

async function withPending(action) {
  actionPending = true;
  updateControls();
  try {
    return await action();
  } finally {
    actionPending = false;
    updateControls();
  }
}

async function loadConfig() {
  const config = await fetchJSON("/api/config");
  applyConfig(config);
}

async function refreshStatus() {
  if (exiting) return { status: "idle" };
  const snapshot = await fetchJSON("/api/status");
  applyStatus(snapshot);
  return snapshot;
}

async function saveConfig() {
  validateForm();
  const config = await fetchJSON("/api/config", {
    method: "POST",
    body: JSON.stringify(readForm()),
  });
  applyConfig(config);
  return config;
}

async function postStart() {
  const snapshot = await fetchJSON("/api/start", { method: "POST" });
  applyStatus(snapshot);
  return snapshot;
}

async function startRunner() {
  await withPending(async () => {
    await saveConfig();
    await postStart();
  });
  showNotice("配置已保存，改写已启动", "success");
}

async function stopRunner() {
  await withPending(async () => {
    const snapshot = await fetchJSON("/api/stop", { method: "POST" });
    applyStatus(snapshot);
  });
  showNotice("正在停止改写", "success");
}

async function saveOrApply() {
  if (currentStatus !== "running") {
    await withPending(saveConfig);
    showNotice("配置已保存", "success");
    return;
  }

  await withPending(() => UA3FUI.applyAndRestart({
    save: saveConfig,
    stop: async () => {
      const snapshot = await fetchJSON("/api/stop", { method: "POST" });
      applyStatus(snapshot);
    },
    refresh: refreshStatus,
    pause: () => new Promise((resolve) => window.setTimeout(resolve, 100)),
    start: postStart,
  }));
  showNotice("新配置已应用，改写已重新启动", "success");
}

function showError(error) {
  showNotice(error.message, "error");
}

editableFields.forEach((field) => {
  field.addEventListener("input", () => {
    formDirty = true;
    if (field === ttlEl) {
      renderValidation(UA3FUI.validateConfig(readForm()));
    }
    updateControls();
  });
});

presetEl.addEventListener("change", () => {
  uaEl.value = presets[presetEl.value] || presets.wechat;
});

saveBtn.addEventListener("click", () => saveOrApply().catch(showError));
startBtn.addEventListener("click", () => startRunner().catch(showError));
stopBtn.addEventListener("click", () => stopRunner().catch(showError));
exitBtn.addEventListener("click", async () => {
  if (!window.confirm("退出后将停止流量改写并关闭后台程序，确定退出吗？")) return;

  actionPending = true;
  updateControls();
  window.clearInterval(pollTimer);
  try {
    await fetchJSON("/api/quit", { method: "POST" });
  } catch (error) {
    console.debug("application server closed", error);
  }
  exiting = true;
  statusTextEl.textContent = "程序已退出";
  statusMetaEl.textContent = "现在可以关闭此页面";
  statusCardEl.dataset.status = "idle";
  showNotice("后台程序已退出", "success");
});

async function bootstrap() {
  await loadConfig();
  await refreshStatus();
  pollTimer = window.setInterval(() => refreshStatus().catch(showError), 1500);
}

bootstrap().catch(showError);
