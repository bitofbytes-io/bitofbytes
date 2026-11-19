// Compiles each file given on the command line as a classic browser script,
// without running it. `node --check` parses CommonJS, which allows a top-level
// `return` that a browser rejects; vm.Script uses the plain script goal.
"use strict";

const fs = require("node:fs");
const vm = require("node:vm");

let failed = false;
for (const file of process.argv.slice(2)) {
  try {
    new vm.Script(fs.readFileSync(file, "utf8"), { filename: file });
  } catch (err) {
    failed = true;
    console.error(`${file}: ${err.name}: ${err.message}`);
  }
}
process.exit(failed ? 1 : 0);
