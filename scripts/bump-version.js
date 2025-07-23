#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Parse command line arguments
const args = process.argv.slice(2);
const appName = args[0];

const validApps = ['landing', 'dashboard', 'menu'];

function showUsage() {
    console.log(`
Usage: node scripts/bump-version.js <app>

Apps: ${validApps.join(', ')}

This script automatically increments the patch version, with rollover logic:
- Patch increments by 1 (0.0.1 → 0.0.2)
- At patch 100, rolls to next minor (0.0.99 → 0.1.0)  
- At minor 100, rolls to next major (0.99.99 → 1.0.0)

Examples:
  node scripts/bump-version.js landing     # 0.0.1 → 0.0.2
  node scripts/bump-version.js dashboard   # 0.0.99 → 0.1.0 (rollover)
  node scripts/bump-version.js menu        # 0.99.99 → 1.0.0 (rollover)
`);
}

if (!appName || !validApps.includes(appName)) {
    console.error('❌ Invalid or missing app name');
    showUsage();
    process.exit(1);
}

function bumpVersion(currentVersion) {
    const parts = currentVersion.split('.').map(Number);
    let [major, minor, patch] = parts;

    // Increment patch
    let newPatch = patch + 1;
    let newMinor = minor;
    let newMajor = major;

    // Rollover logic like deploy.sh
    if (newPatch > 99) {
        newPatch = 0;
        newMinor = minor + 1;

        if (newMinor > 99) {
            newMinor = 0;
            newMajor = major + 1;
        }
    }

    return `${newMajor}.${newMinor}.${newPatch}`;
}

try {
    const packagePath = path.join(__dirname, '..', 'apps', appName, 'package.json');

    if (!fs.existsSync(packagePath)) {
        console.error(`❌ Package.json not found for app: ${appName}`);
        console.error(`Expected path: ${packagePath}`);
        process.exit(1);
    }

    const packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
    const currentVersion = packageJson.version;
    const newVersion = bumpVersion(currentVersion);

    packageJson.version = newVersion;

    fs.writeFileSync(packagePath, JSON.stringify(packageJson, null, 2) + '\n');

    console.log(`✅ Successfully bumped ${appName} version: ${currentVersion} → ${newVersion}`);
    console.log(`📁 Updated: ${packagePath}`);
    console.log(`💡 Commit and push this change to trigger deployment for ${appName}`);

} catch (error) {
    console.error('❌ Error bumping version:', error.message);
    process.exit(1);
} 