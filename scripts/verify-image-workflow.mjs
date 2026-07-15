#!/usr/bin/env node

import fs from 'node:fs'
import path from 'node:path'
import process from 'node:process'
import { fileURLToPath, pathToFileURL } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const jsYamlEntry = path.join(root, 'apps', 'web', 'node_modules', 'js-yaml', 'dist', 'js-yaml.mjs')
const { load } = await import(pathToFileURL(jsYamlEntry).href)
const workflowPath = path.join(root, '.github', 'workflows', 'images.yml')
const webIgnorePath = path.join(root, 'apps', 'web', 'Dockerfile.dockerignore')
const serverIgnorePath = path.join(root, 'apps', 'server', 'Dockerfile.dockerignore')
const fullSha = /^[0-9a-f]{40}$/
const semverTrigger = 'v[0-9]+.[0-9]+.[0-9]+'

function fail(message) {
  console.error(`image workflow validation failed: ${message}`)
  process.exit(1)
}

function mapping(value, label) {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    fail(`${label} must be a mapping`)
  }
  return value
}

function list(value, label) {
  if (!Array.isArray(value)) {
    fail(`${label} must be a list`)
  }
  return value
}

function verifyActionPins(node) {
  if (Array.isArray(node)) {
    node.forEach(verifyActionPins)
    return
  }
  if (node === null || typeof node !== 'object') {
    return
  }
  for (const [key, value] of Object.entries(node)) {
    if (key === 'uses' && typeof value === 'string' && !value.startsWith('./')) {
      const separator = value.lastIndexOf('@')
      if (separator === -1 || !fullSha.test(value.slice(separator + 1))) {
        fail(`Action must use a full commit SHA: ${value}`)
      }
    }
    verifyActionPins(value)
  }
}

function requirePermissions(job, expected, name) {
  const permissions = mapping(job.permissions, `${name}.permissions`)
  const entries = Object.entries(permissions).sort(([left], [right]) => left.localeCompare(right))
  const expectedEntries = Object.entries(expected).sort(([left], [right]) => left.localeCompare(right))
  if (JSON.stringify(entries) !== JSON.stringify(expectedEntries)) {
    fail(`${name}.permissions must be exactly ${JSON.stringify(expected)}`)
  }
}

function requireMatrix(job, name) {
  const include = list(
    mapping(mapping(job.strategy, `${name}.strategy`).matrix, `${name}.strategy.matrix`).include,
    `${name}.strategy.matrix.include`,
  )
  const actual = include
    .map((entry) => {
      const item = mapping(entry, `${name}.strategy.matrix.include entry`)
      return `${item.component}|${item.image}|${item.dockerfile}`
    })
    .sort()
  const expected = [
    'server|acmhot100-server|apps/server/Dockerfile',
    'web|acmhot100-web|apps/web/Dockerfile',
  ]
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    fail(`${name} matrix must build exactly Web and Server`)
  }
}

function jobSteps(job, name) {
  return list(job.steps, `${name}.steps`).map((entry) =>
    mapping(entry, `${name}.steps entry`),
  )
}

function findAction(steps, action) {
  const step = steps.find(
    (entry) => typeof entry.uses === 'string' && entry.uses.startsWith(`${action}@`),
  )
  if (!step) {
    fail(`missing ${action} step`)
  }
  return step
}

function verifyBuildJob(job) {
  if (job.if !== "${{ github.event_name == 'pull_request' }}") {
    fail('build job must run only for pull requests')
  }
  requirePermissions(job, { contents: 'read' }, 'build')
  requireMatrix(job, 'build')
  const steps = jobSteps(job, 'build')
  if (
    steps.some(
      (step) =>
        typeof step.uses === 'string' &&
        (step.uses.startsWith('docker/login-action@') || step.uses.startsWith('actions/attest@')),
    )
  ) {
    fail('pull request build job must not log in or create registry attestations')
  }
  const build = findAction(steps, 'docker/build-push-action')
  if (mapping(build.with, 'build build-push-action.with').push !== false) {
    fail('pull request build must set push: false')
  }
}

function verifyPublishJob(job) {
  requirePermissions(
    job,
    {
      attestations: 'write',
      contents: 'read',
      'id-token': 'write',
      packages: 'write',
    },
    'publish',
  )
  requireMatrix(job, 'publish')
  const steps = jobSteps(job, 'publish')
  findAction(steps, 'docker/login-action')
  const metadata = findAction(steps, 'docker/metadata-action')
  const build = findAction(steps, 'docker/build-push-action')
  const attest = findAction(steps, 'actions/attest')

  const metadataInputs = mapping(metadata.with, 'metadata-action.with')
  const tags = String(metadataInputs.tags ?? '')
    .split('\n')
    .map((tag) => tag.trim())
    .filter(Boolean)
    .sort()
  const expectedTags = [
    'type=raw,value=${{ needs.prepare.outputs.version }}',
    'type=sha,prefix=sha-,format=long',
  ].sort()
  if (JSON.stringify(tags) !== JSON.stringify(expectedTags)) {
    fail('publish metadata must include only the exact version and full Git SHA tags')
  }
  if (String(metadataInputs.flavor ?? '').trim() !== 'latest=false') {
    fail('publish metadata must disable latest')
  }

  const labels = String(metadataInputs.labels ?? '')
  for (const label of [
    'org.opencontainers.image.source=',
    'org.opencontainers.image.revision=',
    'org.opencontainers.image.version=',
  ]) {
    if (!labels.includes(label)) {
      fail(`publish metadata is missing ${label}`)
    }
  }

  const buildInputs = mapping(build.with, 'publish build-push-action.with')
  if (buildInputs.push !== true || buildInputs.sbom !== true) {
    fail('publish build must push and enable SBOM')
  }
  if (!String(buildInputs.provenance ?? '').startsWith('mode=max')) {
    fail('publish build must enable max provenance')
  }
  const attestInputs = mapping(attest.with, 'attest.with')
  if (attestInputs['push-to-registry'] !== true) {
    fail('artifact attestation must be pushed to the registry')
  }

  const commands = steps.map((step) => String(step.run ?? '')).join('\n')
  if (
    !commands.includes('docker buildx imagetools inspect') ||
    !commands.includes('${IMAGE_NAME}:${RELEASE_VERSION}') ||
    !commands.includes('${IMAGE_NAME}:sha-${GITHUB_SHA}') ||
    !commands.includes("published=true") ||
    !commands.includes("published=false")
  ) {
    fail('publish job must detect complete, new, and inconsistent immutable-tag states')
  }
}

function verifyContextAllowlist(filePath, expected, label) {
  if (!fs.existsSync(filePath)) {
    fail(`${label} Dockerfile.dockerignore is missing`)
  }
  const patterns = fs
    .readFileSync(filePath, 'utf8')
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line && !line.startsWith('#'))
  if (JSON.stringify(patterns) !== JSON.stringify(expected)) {
    fail(`${label} Docker context must match its build-input allowlist exactly`)
  }
}

function verifyBuildContexts() {
  verifyContextAllowlist(
    webIgnorePath,
    [
      '**',
      '!apps/',
      '!apps/web/',
      '!apps/web/Dockerfile',
      '!apps/web/index.html',
      '!apps/web/nginx.conf',
      '!apps/web/package.json',
      '!apps/web/package-lock.json',
      '!apps/web/src/',
      '!apps/web/src/**',
      '!apps/web/tsconfig.json',
      '!apps/web/vite.config.ts',
    ],
    'Web',
  )
  verifyContextAllowlist(
    serverIgnorePath,
    [
      '**',
      '!apps/',
      '!apps/server/',
      '!apps/server/**',
      '!seed/',
      '!seed/problems/',
      '!seed/problems/**',
    ],
    'Server',
  )
}

function main() {
  if (!fs.existsSync(workflowPath)) {
    fail('.github/workflows/images.yml is missing')
  }
  let workflow
  try {
    workflow = mapping(
      load(fs.readFileSync(workflowPath, 'utf8'), { filename: workflowPath }),
      'workflow',
    )
  } catch (error) {
    fail(`workflow YAML is invalid: ${error instanceof Error ? error.message : String(error)}`)
  }

  const triggers = mapping(workflow.on, 'on')
  const allowedTriggers = ['pull_request', 'push', 'workflow_dispatch']
  if (JSON.stringify(Object.keys(triggers).sort()) !== JSON.stringify(allowedTriggers.sort())) {
    fail('image workflow must use only pull_request, version tag, and manual triggers')
  }
  const pullRequest = mapping(triggers.pull_request, 'on.pull_request')
  if (JSON.stringify(list(pullRequest.branches, 'on.pull_request.branches')) !== '["main"]') {
    fail('image pull request builds must target only main')
  }
  const push = mapping(triggers.push, 'on.push')
  if (JSON.stringify(Object.keys(push).sort()) !== '["tags"]') {
    fail('image publication push trigger must not include branches')
  }
  if (JSON.stringify(list(push.tags, 'on.push.tags')) !== JSON.stringify([semverTrigger])) {
    fail(`publish tag trigger must be exactly ${semverTrigger}`)
  }
  if (!Object.hasOwn(triggers, 'workflow_dispatch')) {
    fail('controlled workflow_dispatch publishing is missing')
  }
  if (Object.keys(mapping(workflow.permissions, 'permissions')).length !== 0) {
    fail('top-level permissions must be empty; grant per job')
  }

  const jobs = mapping(workflow.jobs, 'jobs')
  if (JSON.stringify(Object.keys(jobs).sort()) !== '["build","prepare","publish"]') {
    fail('workflow must contain only build, prepare, and publish jobs')
  }
  verifyBuildJob(mapping(jobs.build, 'jobs.build'))
  const prepare = mapping(jobs.prepare, 'jobs.prepare')
  requirePermissions(prepare, { contents: 'read' }, 'prepare')
  const prepareSteps = jobSteps(prepare, 'prepare')
  const prepareCommands = prepareSteps.map((step) => String(step.run ?? '')).join('\n')
  const prepareEnvironment = prepareSteps
    .map((step) => JSON.stringify(step.env ?? {}))
    .join('\n')
  if (
    !prepareCommands.includes('SEMVER_PATTERN') ||
    !prepareCommands.includes('refs/heads/main') ||
    !prepareCommands.includes('refs/tags/') ||
    !prepareCommands.includes('git merge-base --is-ancestor') ||
    !prepareEnvironment.includes('github.ref_name') ||
    !prepareEnvironment.includes('inputs.version')
  ) {
    fail('prepare job must validate tag and manual semantic versions')
  }
  verifyPublishJob(mapping(jobs.publish, 'jobs.publish'))
  verifyActionPins(workflow)
  verifyBuildContexts()
  console.log('image workflow validation passed')
}

main()
