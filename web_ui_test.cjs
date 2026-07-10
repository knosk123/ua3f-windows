const test = require("node:test");
const assert = require("node:assert/strict");

function loadUI() {
  try {
    return require("./web/ui.js");
  } catch {
    return {};
  }
}

test("presents runner statuses in Chinese with uptime", () => {
  const ui = loadUI();
  assert.equal(typeof ui.statusPresentation, "function");

  const presentation = ui.statusPresentation(
    { status: "running", startedAt: "2026-07-10T10:00:00Z" },
    new Date("2026-07-10T10:01:05Z"),
  );

  assert.deepEqual(presentation, {
    label: "运行中",
    meta: "运行正常 · 已运行 1 分 5 秒",
    tone: "running",
  });
});

test("derives state-aware controls for a dirty running config", () => {
  const ui = loadUI();
  assert.equal(typeof ui.controlState, "function");

  assert.deepEqual(ui.controlState("running", true, false), {
    fieldsDisabled: false,
    saveDisabled: false,
    saveLabel: "应用并重启",
    startDisabled: true,
    stopDisabled: false,
    exitDisabled: false,
  });
});

test("validates TTL", () => {
  const ui = loadUI();
  assert.equal(typeof ui.validateConfig, "function");

  assert.deepEqual(ui.validateConfig({ ttl: 0 }), {
    ttl: "TTL 必须是 1 到 255 之间的整数",
  });
  assert.deepEqual(ui.validateConfig({ ttl: 64 }), {});
});

test("ignores legacy ports because HTTP is detected on every TCP port", () => {
  const ui = loadUI();
  assert.deepEqual(ui.validateConfig({ ttl: 64, ports: "not-a-port" }), {});
});

test("applies a running config by saving before restart", async () => {
  const ui = loadUI();
  assert.equal(typeof ui.applyAndRestart, "function");

  const calls = [];
  const result = await ui.applyAndRestart({
    save: async () => calls.push("save"),
    stop: async () => calls.push("stop"),
    refresh: async () => {
      calls.push("refresh");
      return { status: "idle" };
    },
    pause: async () => calls.push("pause"),
    start: async () => {
      calls.push("start");
      return "started";
    },
  });

  assert.equal(result, "started");
  assert.deepEqual(calls, ["save", "stop", "refresh", "start"]);
});
