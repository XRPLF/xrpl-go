// This runs only in the trusted reporting job, not in the build container.
const marker = '<!-- xrpl-go-dev-portal-check -->';
const portal = 'https://github.com/XRPLF/xrpl-dev-portal';
const safePath = /^(?!.*\.\.)[A-Za-z0-9_][A-Za-z0-9_./-]{0,200}$/;

// The summary comes from compiler output, so accept only validated paths and inert text.
function parseSummary(encoded) {
  let raw;
  try {
    raw = JSON.parse(Buffer.from(encoded || '', 'base64').toString('utf8'));
  } catch {
    return null;
  }
  if (!raw || !Number.isInteger(raw.checked) || !Array.isArray(raw.failed)) return null;
  const failed = raw.failed.slice(0, 100).flatMap(entry => {
    if (typeof entry?.package !== 'string' || !safePath.test(entry.package)) return [];
    const errors = (Array.isArray(entry.errors) ? entry.errors : []).slice(0, 10).map(error => ({
      file: typeof error?.file === 'string' && safePath.test(error.file) ? error.file : null,
      line: Number.isInteger(error?.line) && error.line > 0 ? error.line : null,
      message: cleanMessage(error?.message),
    }));
    return [{ package: entry.package, errors }];
  });
  return { checked: raw.checked, failed };
}

function cleanMessage(message) {
  return String(message ?? '')
    // Shorten "github.com/Peersyst/xrpl-go/xrpl/transaction/types".MPTAmount to types.MPTAmount.
    .replace(/"[A-Za-z0-9_.\/-]*\/([A-Za-z0-9_-]+)"\./g, '$1.')
    .replace(/[\u0000-\u001f\u007f`]/g, ' ')
    .slice(0, 300);
}

function renderFailures(summary, ref) {
  const tree = `${portal}/tree/${ref}/_code-samples`;
  const blob = `${portal}/blob/${ref}/_code-samples`;
  return summary.failed.map(({ package: pkg, errors }) => {
    const lines = errors.map(({ file, line, message }) => {
      const where = file
        ? `[\`${file.slice(pkg.length + 1) || file}${line ? `:${line}` : ''}\`](${blob}/${file}${line ? `#L${line}` : ''})`
        : '';
      return `- ${where}${where ? ': ' : ''}\`${message}\``;
    });
    return `### [\`${pkg}\`](${tree}/${pkg})\n\n${lines.join('\n') || '- No compiler errors were captured. Check the build report.'}`;
  }).join('\n\n');
}

function renderBody({ release, commit, run, summary }) {
  const ref = commit === 'unresolved' ? 'master' : commit;
  const commitLink = commit === 'unresolved' ? 'unresolved' : `[\`${commit.slice(0, 7)}\`](${portal}/commit/${commit})`;
  const rows = [
    `| xrpl-go release | ${release === 'unresolved' ? release : `[\`${release}\`](https://github.com/XRPLF/xrpl-go/releases/tag/${release})`} |`,
    `| Developer portal commit | ${commitLink} |`,
    `| Workflow run | [Logs and build report](${run}) |`,
  ];
  let intro, details;
  if (summary?.failed.length) {
    intro = 'Some developer portal Go examples do not compile against the latest stable xrpl-go release.';
    rows.splice(2, 0, `| Failed packages | ${summary.failed.length} of ${summary.checked} |`);
    details = `## Failed examples

Update each example below to match the xrpl-go API. Paths are relative to \`_code-samples\`.

${renderFailures(summary, ref)}`;
  } else {
    intro = 'The developer portal Go example check failed before it reported any compiler error.';
    details = `## Setup failure

The check likely failed during setup or dependency download. Check the [workflow logs](${run}) before changing the documentation.`;
  }
  return `${marker}
${intro}

| | |
| --- | --- |
${rows.join('\n')}

${details}

---

<sub>This is a compile-only Linux check. It does not run examples or check ledger behavior. The full compiler output is in the <b>dev-portal-build-report</b> artifact of the workflow run. The next successful check closes this issue.</sub>
`;
}

module.exports = async function report({ github, context, env }) {
  const result = env.CHECK_RESULT;
  if (!['success', 'failure'].includes(result)) return;
  const repo = context.repo;
  const run = `${context.serverUrl}/${repo.owner}/${repo.repo}/actions/runs/${context.runId}`;
  const release = /^v\d+\.\d+\.\d+$/.test(env.RELEASE || '') ? env.RELEASE : 'unresolved';
  const commit = /^[a-f0-9]{40}$/.test(env.PORTAL_SHA || '') ? env.PORTAL_SHA : 'unresolved';
  const issues = await github.paginate(github.rest.issues.listForRepo, {
    ...repo, state: 'open', creator: 'github-actions[bot]', per_page: 100,
  });
  // Update the newest open tracked issue. A closed issue stays closed, and a new failure opens a new one.
  const issue = issues
    .filter(issue => !issue.pull_request && issue.body?.startsWith(marker))
    .sort((a, b) => b.number - a.number)[0];
  if (result === 'success') {
    if (issue) {
      await github.rest.issues.createComment({
        ...repo, issue_number: issue.number,
        body: `All developer portal Go examples compile against ${release}. [Workflow run](${run}).`,
      });
      await github.rest.issues.update({
        ...repo, issue_number: issue.number, state: 'closed', state_reason: 'completed',
      });
    }
    return;
  }
  const summary = parseSummary(env.SUMMARY);
  const count = summary?.failed.length;
  const title = count
    ? `Developer portal: ${count} Go example${count === 1 ? ' fails' : 's fail'} to compile against xrpl-go ${release}`
    : 'Developer portal Go example check failed';
  const body = renderBody({ release, commit, run, summary });
  if (issue) {
    await github.rest.issues.update({ ...repo, issue_number: issue.number, title, body });
  } else {
    await github.rest.issues.create({ ...repo, title, body });
  }
};
