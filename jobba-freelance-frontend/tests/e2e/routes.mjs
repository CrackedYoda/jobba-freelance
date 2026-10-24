// End-to-end route sweep against a running server and its LOCAL database.
// Creates temporary wallet-session users directly in the database, exercises every page and API route
// a client, freelancer, outsider and admin can reach, then removes what it created.
// Usage: E2E_BASE_URL=http://localhost:3000 npm run test:e2e (server must use the same DATABASE_URL).
import { createHash, randomBytes } from "node:crypto";
import bs58 from "bs58";
import pg from "pg";

const BASE = process.env.E2E_BASE_URL ?? "http://localhost:3000";
const db = new pg.Client({ connectionString: process.env.E2E_DATABASE_URL ?? "postgres://jobba:jobba@localhost:5434/jobba_freelance" });
await db.connect();
const results = [];
const check = (name, ok, detail = "") => { results.push({ name, ok, detail }); };

const wallet = () => bs58.encode(randomBytes(32)); // a valid, empty Solana public key
async function user(name, roles, completed = true, fixedWallet) {
  const w = fixedWallet ?? wallet();
  const { rows } = await db.query("insert into profiles (wallet_public_key, display_name, headline, location, bio, hourly_rate_atomic, onboarding_status, wallet_verified_at) values ($1,$2,'Smoke tester','Remote','Smoke test bio',50000000,$3,now()) returning id", [w, name, completed ? "completed" : "in_progress"]);
  for (const r of roles) await db.query("insert into profile_roles (profile_id, role) values ($1,$2)", [rows[0].id, r]);
  const token = randomBytes(32).toString("base64url");
  await db.query("insert into auth_sessions (profile_id, token_hash, wallet_public_key, expires_at) values ($1,$2,$3, now() + interval '1 hour')", [rows[0].id, createHash("sha256").update(token).digest(), w]);
  return { id: rows[0].id, cookie: `jobba_session=${token}` };
}
const client = await user("Smoke Client", ["client"]);
const free = await user("Smoke Freelancer", ["freelancer"]);
const other = await user("Smoke Outsider", ["freelancer"]);
const admin = await user("Smoke Admin", ["admin", "client"]);
const newcomer = await user("Smoke Newcomer", [], false);

async function req(path, { as, method = "GET", body, headers = {} } = {}) {
  const res = await fetch(BASE + path, { method, redirect: "manual", headers: { ...(as ? { cookie: as.cookie } : {}), ...(method !== "GET" ? { origin: BASE, "content-type": "application/json" } : {}), ...headers }, body: body ? JSON.stringify(body) : undefined });
  const text = await res.text();
  let json; try { json = JSON.parse(text); } catch { /* html */ }
  return { status: res.status, json, text, location: res.headers.get("location") };
}
const expect = async (name, path, status, opts) => { const r = await req(path, opts); check(`${name} ${opts?.method ?? "GET"} ${path} → ${status}`, r.status === status, `got ${r.status} ${r.json?.error?.code ?? r.location ?? ""}`); return r; };

// Public pages
for (const p of ["/", "/projects", "/freelancers", "/help", "/faq", "/terms", "/privacy", "/connect", "/icon.svg", "/opengraph-image", "/illustrations/escrow-flow.svg", "/illustrations/wallet.svg", "/illustrations/empty-projects.svg", "/illustrations/empty-contracts.svg", "/illustrations/not-found.svg"]) await expect("public", p, 200);
for (const slug of ["getting-started", "wallets-and-sign-in", "for-clients", "for-freelancers", "escrow", "fees", "disputes", "security"]) await expect("doc", `/help/${slug}`, 200);
await expect("missing doc", "/help/nope", 404);
await expect("missing page", "/does-not-exist", 404);
await expect("health", "/api/v1/health", 200);
await expect("session anon", "/api/v1/auth/session", 200);

// Guards when signed out
for (const p of ["/dashboard", "/dashboard/client", "/dashboard/freelancer", "/contracts", "/admin/disputes"]) { const r = await req(p); check(`guard ${p} → /connect`, r.status === 307 && r.location?.includes("/connect"), `got ${r.status} ${r.location}`); }
await expect("anon create project", "/api/v1/projects", 401, { method: "POST", body: { title: "x" } });

// Client posts a draft, edits, publishes
// Resumable sign-up: an unfinished profile can save progress but cannot use the marketplace
await expect("onboarding state", "/api/v1/onboarding", 200, { as: newcomer });
await expect("unfinished cannot post", "/api/v1/projects", 403, { as: newcomer, method: "POST", body: { title: "x" } });
await expect("unfinished dashboard redirects", "/dashboard", 307, { as: newcomer });
const saved = await expect("save step", "/api/v1/onboarding", 200, { as: newcomer, method: "PATCH", body: { step: 1, roles: ["freelancer"], displayName: "Smoke Newcomer", location: "Remote" } });
check("progress persisted", saved.json?.data?.step === 1 && saved.json?.data?.roles?.includes("freelancer"));
const resumed = await req("/api/v1/onboarding", { as: newcomer });
check("resume returns saved step", resumed.json?.data?.step === 1 && resumed.json?.data?.displayName === "Smoke Newcomer");
const incomplete = await expect("complete blocked while details missing", "/api/v1/onboarding/complete", 422, { as: newcomer, method: "POST" });
check("lists missing details", Array.isArray(incomplete.json?.error?.details?.missing) && incomplete.json.error.details.missing.length > 0);
const png = Buffer.from("89504e470d0a1a0a0000000d49484452000000010000000108060000001f15c4890000000d49444154789c6360000002000154a24f5d0000000049454e44ae426082", "hex");
const form = new FormData(); form.set("file", new Blob([png], { type: "image/png" }), "a.png");
const up = await fetch(BASE + "/api/v1/onboarding/avatar", { method: "POST", headers: { cookie: newcomer.cookie, origin: BASE }, body: form });
check("avatar upload → 200", up.status === 200, `got ${up.status}`);
const bad = new FormData(); bad.set("file", new Blob(["<svg/>"], { type: "image/png" }), "a.png");
const upBad = await fetch(BASE + "/api/v1/onboarding/avatar", { method: "POST", headers: { cookie: newcomer.cookie, origin: BASE }, body: bad });
check("non-image upload rejected → 400", upBad.status === 400, `got ${upBad.status}`);
await expect("avatar served", `/api/v1/profiles/${newcomer.id}/avatar`, 200);
const allSkills = (await req("/api/v1/onboarding", { as: newcomer })).json.data.availableSkills;
await expect("fill freelancer details", "/api/v1/onboarding", 200, { as: newcomer, method: "PATCH", body: { headline: "Smoke engineer", bio: "Long enough biography for the completeness check to pass.", hourlyRateAtomic: "40000000", yearsExperience: 3, skillIds: [allSkills[0].id], experience: [{ title: "Engineer", organization: "Acme", startYear: 2021, endYear: null }] } });
const poor = await req("/api/v1/onboarding/complete", { as: newcomer, method: "POST" });
check("empty wallet cannot finish → 402 INSUFFICIENT_BALANCE", poor.status === 402 && poor.json?.error?.code === "INSUFFICIENT_BALANCE", `got ${poor.status} ${poor.json?.error?.code}`);
// A funded wallet (the public devnet deployer address, > $1 of SOL) can finish sign-up as a client.
const FUNDED = process.env.E2E_FUNDED_WALLET ?? "BMPZFSJ4oiSgmMvbXKfyYyBCvtvR4vx637ZvooRio7c8";
await db.query("delete from profiles where wallet_public_key = $1 and display_name like 'Smoke %'", [FUNDED]);
const funded = await user("Smoke Funded", [], false, FUNDED);
const form2 = new FormData(); form2.set("file", new Blob([png], { type: "image/png" }), "a.png");
await fetch(BASE + "/api/v1/onboarding/avatar", { method: "POST", headers: { cookie: funded.cookie, origin: BASE }, body: form2 });
await expect("fill client details", "/api/v1/onboarding", 200, { as: funded, method: "PATCH", body: { roles: ["client"], displayName: "Smoke Funded", location: "Remote", companyName: "Smoke Labs", industry: "Software & SaaS", companySize: "2-10", hiringNeeds: "We hire designers and Solana engineers regularly." } });
const done = await req("/api/v1/onboarding/complete", { as: funded, method: "POST" });
check("funded wallet completes sign-up → 200", done.status === 200 && done.json?.data?.status === "completed", `got ${done.status} ${done.json?.error?.code ?? ""}`);
await expect("completed client can open dashboard", "/dashboard/client", 200, { as: funded });
const skills = (await req("/api/v1/onboarding", { as: client })).json.data.availableSkills;
const created = await expect("client draft", "/api/v1/projects", 201, { as: client, method: "POST", body: { title: "Smoke project draft", description: "A".repeat(60), category: "Development", budgetAtomic: "2000000000", deliveryDays: 10, skillIds: [skills[0].id], status: "draft" } });
const pid = created.json.data.id;
await expect("draft hidden from others", `/projects/${pid}`, 404, { as: free });
await expect("draft visible to owner", `/projects/${pid}`, 200, { as: client });
await expect("owner edit page", `/projects/${pid}/edit`, 200, { as: client });
await expect("non-owner edit page", `/projects/${pid}/edit`, 404, { as: free });
await expect("edit draft", `/api/v1/projects/${pid}`, 200, { as: client, method: "PATCH", body: { title: "Smoke project", deliveryDays: 12 } });
await expect("freelancer cannot patch", `/api/v1/projects/${pid}`, 403, { as: free, method: "PATCH", body: { title: "Hijack" } });
await expect("publish", `/api/v1/projects/${pid}`, 200, { as: client, method: "PATCH", body: { status: "open" } });
await expect("invalid transition", `/api/v1/projects/${pid}`, 409, { as: client, method: "PATCH", body: { status: "draft" } });
await expect("public detail", `/projects/${pid}`, 200);
const list = await req("/projects?q=Smoke"); check("listed in find work", list.text.includes("Smoke project"));

// Freelancer proposes, client declines, freelancer re-proposes
const prop = await expect("propose", `/api/v1/projects/${pid}/proposals`, 201, { as: free, method: "POST", body: { coverLetter: "I will deliver this cleanly with tests.", amountAtomic: "1900000000", estimatedDays: 9 } });
const propId = prop.json.data.id;
await expect("client cannot propose", `/api/v1/projects/${pid}/proposals`, 403, { as: client, method: "POST", body: { coverLetter: "I will deliver this cleanly with tests.", amountAtomic: "1", estimatedDays: 1 } });
await expect("terms locked with proposals", `/api/v1/projects/${pid}`, 409, { as: client, method: "PATCH", body: { budgetAtomic: "3000000000" } });
await expect("client lists proposals", `/api/v1/projects/${pid}/proposals`, 200, { as: client });
await expect("outsider cannot list proposals", `/api/v1/projects/${pid}/proposals`, 403, { as: other });
await expect("freelancer cannot reject", `/api/v1/proposals/${propId}/reject`, 403, { as: free, method: "POST" });
await expect("client rejects", `/api/v1/proposals/${propId}/reject`, 200, { as: client, method: "POST" });
await expect("withdrawn/rejected cannot be withdrawn", `/api/v1/proposals/${propId}/withdraw`, 409, { as: free, method: "POST" });
await expect("second freelancer proposes", `/api/v1/projects/${pid}/proposals`, 201, { as: other, method: "POST", body: { coverLetter: "Another clear approach for this work.", amountAtomic: "1800000000", estimatedDays: 8 } });
const p2 = (await req(`/api/v1/projects/${pid}/proposals`, { as: client })).json.data.items.find((p) => p.status === "pending").id;
await expect("withdraw own", `/api/v1/proposals/${p2}/withdraw`, 200, { as: other, method: "POST" });
const again = await expect("re-propose after withdraw", `/api/v1/projects/${pid}/proposals`, 201, { as: other, method: "POST", body: { coverLetter: "Updated approach after a second look.", amountAtomic: "1850000000", estimatedDays: 8 } });
check("revived same proposal", again.json.data.id === p2);

// Select & fund → contract
await expect("select needs idempotency key", `/api/v1/proposals/${p2}/fund`, 400, { as: client, method: "POST" });
const contract = await expect("select proposal", `/api/v1/proposals/${p2}/fund`, 200, { as: client, method: "POST", headers: { "idempotency-key": "smoke-select-1" } });
const cid = contract.json.data.id;
const again2 = await req(`/api/v1/proposals/${p2}/fund`, { as: client, method: "POST", headers: { "idempotency-key": "smoke-select-2" } });
check("select is idempotent", again2.json?.data?.id === cid, JSON.stringify(again2.json?.error ?? ""));
await expect("contract page client", `/contracts/${cid}`, 200, { as: client });
await expect("contract page freelancer", `/contracts/${cid}`, 200, { as: other });
await expect("contract page outsider", `/contracts/${cid}`, 404, { as: free });
await expect("contracts list", "/contracts", 200, { as: client });
await expect("client dashboard", "/dashboard/client", 200, { as: client });
await expect("freelancer dashboard", "/dashboard/freelancer", 200, { as: other });
await expect("dashboard router", "/dashboard", 307, { as: other });
await expect("onboarding", "/onboarding", 200, { as: other });
await expect("profile page", `/freelancers/${other.id}`, 200);
await expect("submission list", `/api/v1/contracts/${cid}/submission`, 200, { as: other });
await expect("conversations", "/api/v1/conversations", 200, { as: client });
await expect("cancel project with contract", `/api/v1/projects/${pid}`, 409, { as: client, method: "PATCH", body: { status: "cancelled" } });
await expect("dispute before funding", "/api/v1/disputes", 409, { as: client, method: "POST", body: { contractId: cid, category: "other", statement: "Testing a dispute before funding happens.", evidence: [], idempotencyKey: "smoke-dispute-1" } });
const chain = await req(`/api/v1/contracts/${cid}/chain-state`, { as: client });
check("chain-state responds (needs deployment manifest)", [200, 503].includes(chain.status), `got ${chain.status} ${chain.json?.error?.code ?? ""}`);
const fund = await req(`/api/v1/contracts/${cid}/transactions/fund`, { as: client, method: "POST", body: { idempotencyKey: "smoke-fund-1" } });
check("fund tx endpoint responds (needs deployment manifest)", [200, 503].includes(fund.status), `got ${fund.status} ${fund.json?.error?.code ?? ""}`);
await expect("freelancer cannot abandon", `/api/v1/contracts/${cid}/abandon`, 403, { as: other, method: "POST" });
await expect("client abandons unfunded", `/api/v1/contracts/${cid}/abandon`, 200, { as: client, method: "POST" });
await expect("abandoned contract gone", `/contracts/${cid}`, 404, { as: client });
await expect("cancel project", `/api/v1/projects/${pid}`, 200, { as: client, method: "PATCH", body: { status: "cancelled" } });
const after = await req(`/api/v1/projects/${pid}/proposals`, { as: client });
check("pending proposals declined on cancel", after.json.data.items.every((p) => p.status !== "pending"));

// Admin
await expect("admin queue page", "/admin/disputes", 200, { as: admin });
await expect("admin queue page non-admin", "/admin/disputes", 307, { as: client });
await expect("admin api", "/api/v1/admin/disputes", 200, { as: admin });
await expect("admin api non-admin", "/api/v1/admin/disputes", 403, { as: client });

// Sign-in challenge
await expect("challenge", "/api/v1/auth/challenge", 200, { method: "POST", body: { address: "7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU" } });
await expect("logout", "/api/v1/auth/logout", 200, { method: "POST", as: free });

const failed = results.filter((r) => !r.ok);
for (const r of results) console.log(`${r.ok ? "PASS" : "FAIL"} ${r.name}${r.ok ? "" : `  (${r.detail})`}`);
console.log(`\n${results.length - failed.length}/${results.length} passed`);

// Cleanup
const ids = [client.id, free.id, other.id, admin.id, newcomer.id, funded.id];
await db.query("delete from chain_actions where actor_id = any($1)", [ids]);
await db.query("delete from contracts where client_id = any($1)", [ids]);
await db.query("delete from proposals where freelancer_id = any($1)", [ids]);
await db.query("delete from projects where client_id = any($1)", [ids]);
await db.query("delete from auth_nonces where wallet_public_key in (select wallet_public_key from profiles where id = any($1)) or wallet_public_key = '7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU'", [ids]);
await db.query("delete from auth_sessions where profile_id = any($1)", [ids]);
await db.query("delete from profile_roles where profile_id = any($1)", [ids]);
await db.query("delete from profiles p where id = any($1) and not exists (select 1 from admin_audit_log a where a.admin_id = p.id)", [ids]);
await db.end();

process.exit(failed.length ? 1 : 0);
