#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Parse command line arguments
const args = process.argv.slice(2);
const appName = args[0];
const versionType = args[1] || 'patch'; // patch, minor, major

const validApps = ['landing', 'dashboard', 'menu'];
const validVersionTypes = ['patch', 'minor', 'major'];

function showUsage() {
    console.log(`
Usage: node scripts/bump-version.js <app> [version-type]

Apps: ${validApps.join(', ')}
Version types: ${validVersionTypes.join(', ')} (default: patch)

Examples:
  node scripts/bump-version.js landing patch    # 1.0.0 → 1.0.1
  node scripts/bump-version.js dashboard minor # 1.0.0 → 1.1.0
  node scripts/bump-version.js menu major      # 1.0.0 → 2.0.0
  node scripts/bump-version.js landing         # defaults to patch
`);
}

if (!appName || !validApps.includes(appName)) {
    console.error('❌ Invalid or missing app name');
    showUsage();
    process.exit(1);
}

if (!validVersionTypes.includes(versionType)) {
    console.error('❌ Invalid version type');
    showUsage();
    process.exit(1);
}

function bumpVersion(currentVersion, type) {
    const parts = currentVersion.split('.').map(Number);
    const [major, minor, patch] = parts;

    switch (type) {
        case 'major':
            return `${major + 1}.0.0`;
        case 'minor':
            return `${major}.${minor + 1}.0`;
        case 'patch':
            return `${major}.${minor}.${patch + 1}`;
        default:
            throw new Error('Invalid version type');
    }
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
    const newVersion = bumpVersion(currentVersion, versionType);

    packageJson.version = newVersion;

    fs.writeFileSync(packagePath, JSON.stringify(packageJson, null, 2) + '\n');

    console.log(`✅ Successfully bumped ${appName} version: ${currentVersion} → ${newVersion}`);
    console.log(`📁 Updated: ${packagePath}`);
    console.log(`💡 Commit and push this change to trigger deployment for ${appName}`);

} catch (error) {
    console.error('❌ Error bumping version:', error.message);
    process.exit(1);
} 