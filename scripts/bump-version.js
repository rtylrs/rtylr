#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// Parse command line arguments
const args = process.argv.slice(2);
const appName = args[0];

const validApps = ['landing', 'dashboard', 'menu', 'auth', 'all'];

function showUsage() {
    console.log(`
Usage: node scripts/bump-version.js <app>

Apps: ${validApps.join(', ')}

This script:
1. Ensures dependencies are installed (runs pnpm install if needed)
2. Builds the app first to verify it works
3. Bumps the app version only if build succeeds (with rollover logic)
4. Commits only the version.json change
5. Handles other uncommitted changes separately
6. Creates app-specific git tag
7. Automatically pushes to trigger deployment
8. GitHub Actions will send Telegram notification with commit messages after deployment completes

Version rollover logic:
- Patch increments by 1 (0.0.1 → 0.0.2)
- At patch 100, rolls to next minor (0.0.99 → 0.1.0)  
- At minor 100, rolls to next major (0.99.99 → 1.0.0)

Examples:
  node scripts/bump-version.js landing     # Deploy landing app with change log
  node scripts/bump-version.js dashboard   # Deploy dashboard app with change log
  node scripts/bump-version.js menu        # Deploy menu app with change log
  node scripts/bump-version.js all         # Deploy all apps with individual notifications
`);
}

if (!appName || !validApps.includes(appName)) {
    console.error('❌ Invalid or missing app name');
    showUsage();
    process.exit(1);
}

// Handle "all" option
if (appName === 'all') {
    console.log('🚀 Deploying all apps...');
    const appsToProcess = ['landing', 'dashboard', 'menu', 'auth'];
    const deploymentResults = [];

    for (const app of appsToProcess) {
        console.log(`\n═══════════════════════════════════════`);
        console.log(`🎯 Processing ${app.toUpperCase()}`);
        console.log(`═══════════════════════════════════════`);

        try {
            // Run the script recursively for each app
            execCommand(`node scripts/bump-version.js ${app}`, `Deploying ${app}`);
            console.log(`✅ ${app} deployment completed successfully`);

            deploymentResults.push({
                app,
                success: true
            });
        } catch (error) {
            console.error(`❌ Failed to deploy ${app}:`, error.message);
            deploymentResults.push({
                app,
                success: false,
                error: error.message
            });
        }
    }

    const successCount = deploymentResults.filter(r => r.success).length;

    console.log(`\n🎉 All apps deployment process completed!`);
    console.log(`📦 GitHub Actions will deploy: ${appsToProcess.join(', ')}`);
    console.log('📱 Telegram notifications will be sent after each deployment completes');
    process.exit(successCount === appsToProcess.length ? 0 : 1);
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

function execCommand(command, description) {
    try {
        console.log(`🔄 ${description}...`);
        const output = execSync(command, { encoding: 'utf8', stdio: 'pipe' });
        return output.trim();
    } catch (error) {
        console.error(`❌ Failed to ${description.toLowerCase()}`);
        console.error(error.message);
        process.exit(1);
    }
}



function checkGitStatus() {
    try {
        const status = execSync('git status --porcelain', { encoding: 'utf8' });
        return status.trim();
    } catch (error) {
        console.error('❌ Failed to check git status');
        process.exit(1);
    }
}

try {
    const versionPath = path.join(__dirname, '..', 'apps', appName, 'version.json');
    const relativeVersionPath = `apps/${appName}/version.json`;

    if (!fs.existsSync(versionPath)) {
        console.error(`❌ version.json not found for app: ${appName}`);
        console.error(`Expected path: ${versionPath}`);
        process.exit(1);
    }

    // Check for existing uncommitted changes
    const initialStatus = checkGitStatus();
    const hasOtherChanges = initialStatus.length > 0;

    if (hasOtherChanges) {
        console.log('📝 Found existing uncommitted changes - will handle separately');
    }



    // 1. Ensure dependencies are installed
    const nodeModulesPath = path.join(__dirname, '..', 'node_modules');
    if (!fs.existsSync(nodeModulesPath)) {
        execCommand('pnpm install', 'Installing dependencies');
        console.log('✅ Dependencies installed');
    }

    // 2. Build the app FIRST to verify it works before changing anything
    execCommand(`npx nx build ${appName}`, `Building ${appName} to verify it works`);
    console.log(`✅ ${appName} built successfully - safe to bump version`);

    // 3. Only now bump version since we know the build works
    const versionJson = JSON.parse(fs.readFileSync(versionPath, 'utf8'));
    const currentVersion = versionJson.version;
    const newVersion = bumpVersion(currentVersion);

    versionJson.version = newVersion;

    fs.writeFileSync(versionPath, JSON.stringify(versionJson, null, 2) + '\n');

    console.log(`✅ Bumped ${appName} version: ${currentVersion} → ${newVersion}`);
    console.log(`📁 Updated: ${versionPath}`);

    // 4. Handle git commits
    if (hasOtherChanges) {
        // First commit other changes
        execCommand('git add .', 'Staging all changes');
        execCommand(`git reset HEAD ${relativeVersionPath}`, `Unstaging ${relativeVersionPath}`);

        const stagedFiles = execSync('git diff --cached --name-only', { encoding: 'utf8' }).trim();
        if (stagedFiles) {
            execCommand(`git commit -m "📝 Commit changes before ${appName} version bump"`, 'Committing other changes');
            console.log('✅ Other changes committed separately');
        }
    }

    // 5. Commit only the version.json version change
    execCommand(`git add ${relativeVersionPath}`, `Staging ${relativeVersionPath}`);
    execCommand(`git commit -m "🚀 Bump ${appName} to v${newVersion}"`, `Committing version bump`);
    console.log(`✅ Version bump committed: ${appName} v${newVersion}`);

    // 6. Create app-specific tag
    const tagName = `${appName}-v${newVersion}`;
    execCommand(`git tag ${tagName}`, `Creating tag ${tagName}`);
    console.log(`✅ Created tag: ${tagName}`);

    // 7. Push to trigger deployment automatically
    execCommand(`git push origin main`, `Pushing to trigger deployment`);
    execCommand(`git push origin ${tagName}`, `Pushing tag ${tagName}`);
    console.log(`✅ Pushed to origin/main - deployment triggered!`);

    // 8. Success message
    console.log('');
    console.log('🎉 Deployment process started successfully!');
    console.log(`📦 GitHub Actions will now build and deploy ${appName} v${newVersion}`);
    console.log(`🏷️  Tag created: ${tagName}`);
    console.log('📱 Telegram notification will be sent after deployment completes');

} catch (error) {
    console.error('❌ Error during version bump process:', error.message);
    process.exit(1);
} 