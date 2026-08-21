import { execFileSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import ts from "typescript";

const trackedSourceFiles = new Set(
  execFileSync("git", ["ls-files", "-z", "apps"], { encoding: "utf8" })
    .split("\0")
    .filter((path) => /\.(tsx|jsx)$/.test(path)),
);
const cache = new Map();

function countImplicitButtons(relativePath) {
  if (!trackedSourceFiles.has(relativePath)) {
    throw new Error(`audit case references an untracked source file: ${relativePath}`);
  }
  if (cache.has(relativePath)) return cache.get(relativePath);

  const source = readFileSync(resolve(relativePath), "utf8");
  const sourceFile = ts.createSourceFile(
    relativePath,
    source,
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TSX,
  );
  let count = 0;

  function visit(node) {
    if (ts.isJsxElement(node) || ts.isJsxSelfClosingElement(node)) {
      const tagName = ts.isJsxElement(node) ? node.openingElement.tagName : node.tagName;
      if (tagName.getText(sourceFile) === "button") {
        const attributes = ts.isJsxElement(node)
          ? node.openingElement.attributes
          : node.attributes;
        if (
          !attributes.properties.some(
            (property) => ts.isJsxAttribute(property) && property.name.text === "type",
          )
        ) {
          count += 1;
        }
      }
    }
    ts.forEachChild(node, visit);
  }

  visit(sourceFile);
  cache.set(relativePath, count);
  return count;
}

const suites = await Promise.all(
  Array.from({ length: 25 }, (_, index) =>
    import(`./suite-${String(index + 1).padStart(2, "0")}.audit.mjs`),
  ),
);
const cases = suites.flatMap(({ cases: suiteCases }) => suiteCases);
const failures = [];

for (const [relativePath, label] of cases) {
  const count = countImplicitButtons(relativePath);
  if (count > 0) failures.push(`${label}: ${relativePath} has ${count} implicit button(s)`);
}

if (failures.length > 0) {
  throw new Error(`button audit failed:\n${failures.join("\n")}`);
}

console.log(`verified ${cases.length} button audit cases across ${cache.size} source files`);
