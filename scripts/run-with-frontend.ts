import { existsSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import crypto from "node:crypto";

const PLAN_FILE = "railpack-plan.json";
const FRONTEND_IMAGE = "ghcr.io/railwayapp/railpack:railpack-frontend";

// Parse command line arguments
const args = process.argv.slice(2);
if (args.length === 0) {
  console.error("Please provide a directory path");
  process.exit(1);
}

const dir = args[0];
const envArgs = args.slice(1).filter((arg) => arg.startsWith("--env"));

// Create temp directory to save the build plan to
const randId = Math.random().toString(36).slice(2);
const tmpDir = join("/tmp", "railpack-" + randId);
const planDir = join(tmpDir, "plan");
mkdirSync(planDir, { recursive: true });

const cleanup = () => {
  if (existsSync(tmpDir)) {
    rmSync(tmpDir, { recursive: true, force: true });
  }
};

process.on("exit", cleanup);
process.on("SIGINT", () => {
  cleanup();
  process.exit();
});

// Generate build plan
console.log(`Generating build plan for ${dir}`);
const planResult = spawnSync(
  "go",
  ["run", "cmd/cli/main.go", "plan", dir, "--format", "json"],
  {
    stdio: ["inherit", "pipe", "inherit"],
  }
);

if (planResult.status !== 0) {
  console.error("Failed to generate build plan");
  process.exit(1);
}

const planPath = join(planDir, PLAN_FILE);
writeFileSync(planPath, planResult.stdout);

// Parse all env vars so that we can use them as secrets
const envVars: Record<string, string> = {};
const secretArgs: string[] = [];

envArgs.forEach((arg) => {
  const [_, nameValue] = arg.split("--env ");
  const [name, value] = nameValue.split("=");
  if (name && value) {
    envVars[name] = value;
    secretArgs.push(`--secret id=${name},env=${name}`);
  }
});

// Options that are passed to our custom frontend
const cacheKey = dir;
const secretsHash = crypto
  .createHash("sha256")
  .update(Object.values(envVars).sort().join(""))
  .digest("hex");

// Pipe buildctl and docker load together
const buildctlArgs = [
  "build",
  `--local`,
  `context=${dir}`,
  `--local`,
  `dockerfile=${planDir}`,
  "--frontend=gateway.v0",
  "--opt",
  `source=${FRONTEND_IMAGE}`,
  "--opt",
  `cache-key=${cacheKey}`,
  "--opt",
  `secrets-hash=${secretsHash}`,
  "--output",
  "type=docker,name=test",
  ...secretArgs,
];

console.log("Executing buildctl", buildctlArgs.join(" "));

const buildctl = spawnSync("buildctl", buildctlArgs, {
  stdio: ["inherit", "pipe", "inherit"],
  env: { ...process.env, ...envVars },
});

if (buildctl.status !== 0) {
  console.error("buildctl command failed");
  process.exit(1);
}

const dockerLoad = spawnSync("docker", ["load"], {
  input: buildctl.stdout,
  stdio: ["pipe", "inherit", "inherit"],
});

if (dockerLoad.status !== 0) {
  console.error("docker load failed");
  process.exit(1);
}
