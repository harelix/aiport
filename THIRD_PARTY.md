# Third-party notices

Vendored dependency: `go.yaml.in/yaml/v3` v3.0.6 from
https://github.com/yaml/go-yaml/tree/v3.0.6.

Original source files, README, module metadata, LICENSE and NOTICE are retained in
`vendor/go.yaml.in/yaml/v3`. Dependency license terms are in that LICENSE; this is
separate from Aiport's own license (Apache-2.0, see LICENSE).

Upstream release archive:
https://codeload.github.com/yaml/go-yaml/zip/refs/tags/v3.0.6

Downloaded archive SHA-256:
`658fbd1b815e823278043995acac22927dd0029a3872b90b2a393afee6d4b06f`.
This records the fetched artifact; it is not a publisher signature verification.
Only non-test root Go files and the metadata/notices listed above are vendored.
The application builds/tests offline using Go's automatic vendor mode. Updating
this dependency requires reviewing upstream changes and refreshing vendored files.

The embedded dashboard bundles React 19.3.0 and React DOM 19.3.0, including their
scheduler dependency. MIT license texts are retained in `licenses/`. Icons come from lucide-react 1.52.0
(ISC license, retained in `licenses/lucide-react-LICENSE`); only the icons used are bundled. The frontend
uses Vite 7.3.6 and TypeScript 5.9.3 as build-time tools. This preview was built
using the corresponding already-installed local packages because package-registry
access was unavailable. Direct package versions are pinned in web/package.json;
a clean registry install and transitive lockfile verification remain release gates.
