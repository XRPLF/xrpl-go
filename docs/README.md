# Documentation website

The SDK documentation uses [Docusaurus](https://docusaurus.io/). Run commands below from `docs/`.

## Local setup

Use Node.js 20 or later and Yarn Classic (1.x).

```bash
yarn install --frozen-lockfile
yarn start
```

Open `http://localhost:3000/xrpl-go/`. Most content changes reload automatically. Restart the server after changing configuration if needed.

## Check changes

```bash
yarn typecheck
yarn build
```

`yarn build` checks internal page links and generates `build/`. Preview it with `yarn serve`. A successful site build does not validate Go examples.

## Content ownership

- `docs/`: application guides and migration instructions.
- `sidebars.ts`: explicit guide order and task groups. Add new pages here without changing existing document IDs or URLs unnecessarily.
- `changelog/`: historical release notes. Follow the repository's changelog guidance before changing them.
- `src/pages/index.tsx`: site homepage.
- `docusaurus.config.ts`: navigation, footer, site metadata, and deployment paths.

Teach Go workflows and SDK-specific contracts here. Link protocol fields and server behavior to XRPL documentation. Link full signatures, structs, constants, and error inventories to the Go API reference. Keep contributor and release procedures in the repository guides.

Each guide should identify its audience, show a useful example early, label fragments and prerequisites, and explain expected results. Do not copy a second protocol or API catalogue into a guide.

## Deployment

`.github/workflows/docs-deploy.yml` installs dependencies, builds the site, and deploys the build artifact to GitHub Pages when a push to `main` changes `docs/**` or the workflow file. `.github/workflows/docs-test-deploy.yml` runs the same build, without deploying, on pull requests to `main` that change `docs/**` or either docs workflow file. Neither workflow runs `yarn typecheck` or checks Go examples. Run the checks above locally before review and verify changed Go examples separately.

Do not use the generic `yarn deploy` command for the normal repository workflow. The production base path is `/xrpl-go`.
