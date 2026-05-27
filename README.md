# Retrospector

CDK stack to inspect past malicious activity in security logs by retrocognition

## Standalone Usage

```bash
npm install
make
npx tsc
npx cdk deploy
```

## Usage as a Git Submodule

When this repo is consumed as a submodule:

- Do **not** run `npm install` inside this directory. The parent repo's `node_modules` provides all `@aws-cdk/*` packages and `@types/node`.
- The `tsconfig.json` includes both `./node_modules/@types` and `../node_modules/@types` in `typeRoots` to support both standalone and submodule usage.
- Run `make` here to build Go binaries, then `npx tsc` (which resolves from the parent's `node_modules`).

If you see type conflicts (e.g. "Types have separate declarations of a private property"), it usually means this directory has its own `node_modules` installed that conflicts with the parent's CDK version. Remove it:

```bash
rm -rf node_modules
```
