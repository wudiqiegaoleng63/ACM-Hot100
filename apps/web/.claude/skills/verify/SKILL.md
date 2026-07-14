---
name: verify
summary: Drive the ACM Hot 100 web/API/worker flow against isolated local services.
---

1. Start isolated MySQL on `127.0.0.1:3307`, Redis on `127.0.0.1:6380`, and Mailpit on `127.0.0.1:8025`/`:1025`.
2. Run API migrations and seed with `APP_BASE_URL=http://127.0.0.1:5174`, API `:8081`, mock judge, and an isolated Redis prefix.
3. Start `cmd/api`, `cmd/judge-worker`, then Vite from `apps/web` with `VITE_API_PROXY_TARGET=http://127.0.0.1:8081` on `:5174`.
4. Drive registration → Mailpit link → login → problem → draft → sample → submission → profile → detail with Playwright Chrome. Use `E2E_EXTERNAL_SERVERS=1 E2E_BASE_URL=http://127.0.0.1:5174 E2E_MAILPIT_URL=http://127.0.0.1:8025`.
5. Use a fresh `E2E_RUN_ID`; the test deletes Mailpit messages. Stop processes and remove isolated containers afterward.

Gotchas: system ports 3306/6379 may already be occupied; `npx` must run from `apps/web`; Chrome channel works when Playwright's bundled browser is absent; Monaco CDN can load slowly, but the flow can seed the authenticated local draft and reload.
