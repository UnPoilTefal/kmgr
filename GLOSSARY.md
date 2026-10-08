# kmgr — domain glossary

## Identity (`normalize.ContextName`)

The canonical `{user}@{cluster}` identity kmgr assigns to a managed
kubeconfig. It drives three things at once:

- the source filename (`kubeconfig_{user}@{cluster}.yaml`)
- the cluster name inside the file
- the AuthInfo (user) name inside the file — namespaced by cluster, i.e.
  always equal to the identity itself, to avoid AuthInfo collisions when
  merging multiple source files into one kubeconfig

It is distinct from the **context key** actually written inside a
kubeconfig file, which normally equals the identity but can be overridden
with `kmgr import --ctx <name>`. When overridden, the context key
diverges from the identity while cluster/AuthInfo still follow the
identity — `kmgr check`/`kmgr fix` treat that divergence as a naming
issue by design (they don't know about custom context keys).

Two ways to obtain an identity:

- `normalize.New(user, cluster)` — builds one from raw, untrusted parts
  (sanitizes via `normalize.Name`). Used when importing a new kubeconfig.
- `normalize.Parse(s)` / `normalize.FromFilename(path)` — reads an
  already-conforming identity as-is, no sanitizing. Used for values a
  caller expects to already be valid (a CLI argument, an existing
  filename).

## Fixable source file (`config.SourceCheck.UnfixableReason`)

A source kubeconfig is *fixable* when `kmgr fix` can repair it in place
(rewrite its context/cluster/AuthInfo to match its filename). It is
*unfixable* — and gets quarantined instead — when its content isn't
parseable, or its filename doesn't follow the `kubeconfig_{user}@{cluster}.yaml`
convention. This is a single decision, computed once in
`internal/config`, and shared by `kmgr check` (which hints), `kmgr fix`
(which acts), and `kmgr merge` (which quarantines silently on write).

## Render (`config.OutputMode`, `config.Palette`)

kmgr's structured check/status/list results (`SourceCheck`, `ContextCheck`,
`StatusResult`, `ManagedContext`, `MirrorStatusEntry`) each expose a
`Render(mode, palette) string` method producing either the verbose human
output or the compact machine-readable output used by `kmgr --ai` /
`KMGR_AI` (auto-detected for AI agents via `CLAUDECODE`). Procedural
narration (import/merge/sync steps) is not part of this — it stays on
the plain `info`/`ok`/`warn`/`hint` primitives in `cmd/root.go`, which
are already mode-aware.
