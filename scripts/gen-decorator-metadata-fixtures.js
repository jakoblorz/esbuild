const fs = require('fs')
const os = require('os')
const path = require('path')
const child_process = require('child_process')

const repoRoot = path.join(__dirname, '..')
const ts = require(path.join(repoRoot, '.fixtureenv/node_modules/typescript'))
const tscPath = path.join(repoRoot, '.fixtureenv/node_modules/.bin/tsc')

const casesDir = path.join(repoRoot, 'internal/js_parser/testdata/emit_decorator_metadata/cases')
const fixturesDir = path.join(repoRoot, 'internal/js_parser/testdata/emit_decorator_metadata/fixtures')

const printer = ts.createPrinter({ removeComments: true, newLine: ts.NewLineKind.LineFeed })

function normalizeText(text) {
  return text.replace(/\s+/g, ' ').trim()
}

function stripParens(node) {
  while (ts.isParenthesizedExpression(node)) node = node.expression
  return node
}

function printNode(node, sourceFile) {
  return normalizeText(printer.printNode(ts.EmitHint.Unspecified, node, sourceFile))
}

function serializeExpr(node, sourceFile) {
  node = stripParens(node)

  if (ts.isIdentifier(node)) return node.text
  if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) return JSON.stringify(node.text)
  if (ts.isNumericLiteral(node) || ts.isBigIntLiteral(node)) return node.getText(sourceFile)
  if (node.kind === ts.SyntaxKind.TrueKeyword) return 'true'
  if (node.kind === ts.SyntaxKind.FalseKeyword) return 'false'
  if (node.kind === ts.SyntaxKind.NullKeyword) return 'null'

  if (ts.isPropertyAccessExpression(node)) {
    return `${serializeExpr(node.expression, sourceFile)}.${node.name.text}`
  }

  if (ts.isArrayLiteralExpression(node)) {
    return `[${node.elements.map(item => serializeExpr(item, sourceFile)).join(', ')}]`
  }

  if (ts.isVoidExpression(node)) {
    return `void ${serializeExpr(node.expression, sourceFile)}`
  }

  if (ts.isPrefixUnaryExpression(node)) {
    return `${ts.tokenToString(node.operator)}${serializeExpr(node.operand, sourceFile)}`
  }

  return printNode(node, sourceFile)
}

function extractMetadataCall(node, sourceFile) {
  if (!ts.isCallExpression(node) || node.arguments.length < 2) return null

  const callee = stripParens(node.expression)
  let isMetadataCall = false

  if (ts.isIdentifier(callee)) {
    isMetadataCall = callee.text === '__metadata' || callee.text === '__legacyMetadata'
  } else if (ts.isPropertyAccessExpression(callee)) {
    isMetadataCall = ts.isIdentifier(callee.expression) && callee.expression.text === 'Reflect' && callee.name.text === 'metadata'
  }

  if (!isMetadataCall) return null

  const keyNode = stripParens(node.arguments[0])
  if (!ts.isStringLiteral(keyNode) && !ts.isNoSubstitutionTemplateLiteral(keyNode)) return null

  return {
    key: keyNode.text,
    value: serializeExpr(node.arguments[1], sourceFile),
  }
}

function extractMetadataRecords(js, sourcePathForDiagnostics) {
  const sourceFile = ts.createSourceFile(sourcePathForDiagnostics, js, ts.ScriptTarget.ESNext, true, ts.ScriptKind.JS)
  const records = []

  const visit = (node) => {
    if (ts.isCallExpression(node)) {
      const callee = stripParens(node.expression)
      const helper = ts.isIdentifier(callee) ? callee.text : ''
      if (helper === '__decorate' || helper === '__decorateClass') {
        const args = node.arguments
        if (args.length >= 2) {
          const decoratorsArg = stripParens(args[0])
          if (ts.isArrayLiteralExpression(decoratorsArg)) {
            const metadata = []
            for (const element of decoratorsArg.elements) {
              const entry = extractMetadataCall(element, sourceFile)
              if (entry) metadata.push(entry)
            }

            if (metadata.length > 0) {
              records.push({
                kind: args.length >= 3 ? 'member' : 'class',
                target: serializeExpr(args[1], sourceFile),
                key: args.length >= 3 ? serializeExpr(args[2], sourceFile) : null,
                metadata,
              })
            }
          }
        }
      }
    }

    ts.forEachChild(node, visit)
  }

  visit(sourceFile)
  return records
}

function compileCase(casePath, outDir) {
  const args = [
    tscPath,
    '--target', 'ES2020',
    '--module', 'ESNext',
    '--useDefineForClassFields', 'false',
    '--experimentalDecorators',
    '--emitDecoratorMetadata',
    '--pretty', 'false',
    '--outDir', outDir,
    casePath,
  ]

  child_process.execFileSync(process.execPath, args, {
    cwd: repoRoot,
    stdio: ['ignore', 'pipe', 'pipe'],
  })
}

function main() {
  if (!fs.existsSync(casesDir)) throw new Error(`Missing cases directory: ${casesDir}`)
  fs.mkdirSync(fixturesDir, { recursive: true })

  const files = fs.readdirSync(casesDir)
    .filter(name => name.endsWith('.ts'))
    .sort((a, b) => a.localeCompare(b))

  const tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'esbuild-md-fixtures-'))

  try {
    for (const file of files) {
      const casePath = path.join(casesDir, file)
      const outDir = path.join(tempRoot, path.basename(file, '.ts'))
      fs.mkdirSync(outDir, { recursive: true })

      compileCase(casePath, outDir)

      const emittedPath = path.join(outDir, file.replace(/\.ts$/, '.js'))
      const js = fs.readFileSync(emittedPath, 'utf8')
      const records = extractMetadataRecords(js, emittedPath)

      const fixture = {
        typescriptVersion: ts.version,
        source: file,
        records,
      }

      const fixturePath = path.join(fixturesDir, file.replace(/\.ts$/, '.json'))
      fs.writeFileSync(fixturePath, JSON.stringify(fixture, null, 2) + '\n')
      process.stdout.write(`wrote ${path.relative(repoRoot, fixturePath)}\n`)
    }
  } finally {
    fs.rmSync(tempRoot, { recursive: true, force: true })
  }
}

main()
