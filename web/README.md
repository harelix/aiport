# Dashboard development

React/TypeScript with Vite. `pnpm install` then `pnpm build` from this directory
writes production assets to `../internal/dashboard/dist`, embedded by Go.
No frontend service or package install is required by end users.

During the initial offline build, `node_modules` is a local ignored symlink to
existing packages on the development machine. This is not part of the distributable
or a portable dependency installation. Registry availability is needed for a clean
contributor install and transitive lockfile verification. Build-time versions and
runtime MIT notices are recorded in the root THIRD_PARTY.md and licenses/.

No dummy data or bypass token is built into the application. It always requests
the local authenticated management API. Browser visual QA remains outstanding.
