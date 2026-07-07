const presetEl = document.getElementById("preset");
const uaEl = document.getElementById("ua");
const ttlEl = document.getElementById("ttl");
const portsEl = document.getElementById("ports");
const logEl = document.getElementById("log");
const statusTextEl = document.getElementById("statusText");
const statusMetaEl = document.getElementById("statusMeta");
const packetCountEl = document.getElementById("packetCount");
const ttlCountEl = document.getElementById("ttlCount");
const uaCountEl = document.getElementById("uaCount");
const logsEl = document.getElementById("logs");
const saveBtn = document.getElementById("saveBtn");
const startBtn = document.getElementById("startBtn");
const stopBtn = document.getElementById("stopBtn");

const presets = {
  wechat: "Mozilla/5.0 (Linux; Android 15; RMX6688 Build/AP3A.240617.008; wv) AppleWebKit/537.36",
  pc: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:149.0) Gecko/20100101 Firefox/149.0",
};

async function fetchJSON(url, options) {
  const response = await fetch(url, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || "request failed");
  }
  return data;
}

function readForm() {
  return {
    preset: presetEl.value,
    ua: presets[presetEl.value] || presets.wechat,
    ttl: Number(ttlEl.value),
    ports: portsEl.value.trim(),
    log: logEl.value,
  };
}

function applyConfig(config) {
  presetEl.value = presets[config.preset] ? config.preset : "wechat";
  uaEl.value = presets[presetEl.value] || presets.wechat;
  ttlEl.value = config.ttl || 64;
  portsEl.value = config.ports || "80";
  logEl.value = config.log || "info";
}

function applyStatus(snapshot) {
  statusTextEl.textContent = snapshot.status;
  statusMetaEl.textContent = snapshot.lastError || "运行正常";
  packetCountEl.textContent = snapshot.packetCount || 0;
  ttlCountEl.textContent = snapshot.ttlCount || 0;
  uaCountEl.textContent = snapshot.uaCount || 0;
  logsEl.textContent = (snapshot.logs || []).join("\n");
  logsEl.scrollTop = logsEl.scrollHeight;
}

async function loadConfig() {
  const config = await fetchJSON("/api/config");
  applyConfig(config);
}

async function refreshStatus() {
  const snapshot = await fetchJSON("/api/status");
  applyStatus(snapshot);
}

async function saveConfig() {
  const config = await fetchJSON("/api/config", {
    method: "POST",
    body: JSON.stringify(readForm()),
  });
  applyConfig(config);
}

async function startRunner() {
  await saveConfig();
  const snapshot = await fetchJSON("/api/start", { method: "POST" });
  applyStatus(snapshot);
}

async function stopRunner() {
  const snapshot = await fetchJSON("/api/stop", { method: "POST" });
  applyStatus(snapshot);
}

presetEl.addEventListener("change", () => {
  const preset = presetEl.value;
  uaEl.value = presets[preset] || presets.wechat;
});

saveBtn.addEventListener("click", () => saveConfig().catch(showError));
startBtn.addEventListener("click", () => startRunner().catch(showError));
stopBtn.addEventListener("click", () => stopRunner().catch(showError));

function showError(error) {
  statusTextEl.textContent = "error";
  statusMetaEl.textContent = error.message;
}

async function bootstrap() {
  await loadConfig();
  await refreshStatus();
  window.setInterval(() => refreshStatus().catch(showError), 1500);
}

bootstrap().catch(showError);
