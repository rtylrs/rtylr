# Vbitez Multi-App Deployment System

This project uses a sophisticated per-app versioning and deployment system that allows each application to be versioned and deployed independently using simple `version.json` files.

## 📁 App Structure

The project contains three main applications:

- **Landing** (`apps/landing/`) - Main marketing/public website
- **Dashboard** (`apps/dashboard/`) - Admin dashboard application
- **Menu** (`apps/menu/`) - Digital menu application

Each app has its own `version.json` file with independent versioning.

## 🔄 How Deployment Works

### Automatic Deployment Triggers

The GitHub workflow automatically deploys apps when:

1. **Version Changes**: When you update the `version` field in an app's `version.json` file
2. **File Changes**: The workflow monitors changes to `apps/*/version.json` files
3. **Selective Deployment**: Only apps with version changes are built and deployed

### Current App Versions

- **Landing**: v0.0.2
- **Dashboard**: v0.0.2
- **Menu**: v0.0.1

## 🚀 Deploying Apps

### Method 1: Using the Bump Version Script (Recommended)

Use the provided utility script to bump versions easily:

```bash
# Automatically bump patch version with rollover logic
node scripts/bump-version.js landing     # 0.0.1 → 0.0.2
node scripts/bump-version.js dashboard   # 0.0.99 → 0.1.0 (rollover)
node scripts/bump-version.js menu        # 0.99.99 → 1.0.0 (rollover)
```

The script automatically:

1. **Checks dependencies** and installs if needed (runs `pnpm install`)
2. **Builds the app first** to verify it works (prevents broken deployments)
3. **Bumps the version** only if build succeeds (rollover logic: patch → minor → major at 100)
4. **Commits changes intelligently**:
   - Other uncommitted changes → separate commit
   - Version bump → clean isolated commit
5. **Pushes automatically** to trigger deployment

That's it! The deployment starts immediately.

### Method 2: Manual Version Update

1. Edit the app's `version.json` file:

   ```json
   {
     "version": "0.0.2", // ← Update this
     "app": "landing",
     "description": "Vbitez Landing Page Application"
   }
   ```

2. Commit and push the change:
   ```bash
   git add apps/landing/version.json
   git commit -m "🚀 Deploy landing v0.0.2"
   git push origin main
   ```

### Method 3: Deploy Multiple Apps

To deploy multiple apps simultaneously, update their versions in the same commit:

```bash
# Update multiple app versions (each pushes automatically)
node scripts/bump-version.js landing
node scripts/bump-version.js dashboard
```

## 📋 Deployment Process

When you run the bump script, it:

1. **🔍 Verifies Build**: Ensures your code works before any changes
2. **📈 Bumps Version**: Updates version.json with rollover logic
3. **💾 Commits Smart**: Separates other changes from version bump
4. **🚀 Pushes Auto**: Triggers GitHub Actions deployment
5. **📱 Notifies**: Sends Telegram notification when complete

## 🎯 Deployment Targets

| App       | Build Command        | Deploy Location                            |
| --------- | -------------------- | ------------------------------------------ |
| Landing   | `nx build landing`   | `/subdomains/vbitez.voxire.com/`           |
| Dashboard | `nx build dashboard` | `/subdomains/dashboard.vbitez.voxire.com/` |
| Menu      | `nx build menu`      | `/subdomains/menu.vbitez.voxire.com/`      |

## 📊 Monitoring Deployments

### GitHub Actions

- View deployment status in the **Actions** tab
- Check logs for build and deployment details
- Monitor workflow runs for any failures

### Telegram Notifications

Successful deployments send rich notifications including:

- 📱 List of deployed apps and their versions
- ⏰ Deployment timestamp
- 📝 Recent changes since last deployment
- 🔗 Commit information

## 🛠️ Development Workflow

### For New Features

1. Develop your feature in the relevant app
2. Test locally using `nx serve <app-name>`
3. When ready to deploy:
   ```bash
   node scripts/bump-version.js <app-name>
   # Script builds, commits, pushes, and triggers deployment automatically!
   ```

### For Hotfixes

```bash
# Script handles everything automatically - build, commit, push, deploy
node scripts/bump-version.js <app-name>
```

### For Any Updates

```bash
# Single command deployment - couldn't be simpler!
node scripts/bump-version.js <app-name>
```

### What the Script Does

- ✅ **Builds first** - ensures your code works before committing
- ✅ **Smart commits** - separates other changes from version bumps
- ✅ **Clean history** - each version bump is an isolated commit
- ✅ **Auto-deploys** - pushes automatically to trigger deployment
- ✅ **Clear feedback** - shows exactly what happened and deployment status

## 🔧 Configuration

### Environment Variables

The workflow uses these secrets:

- `SFTP_HOST` - FTP server hostname
- `SFTP_USERNAME` - FTP username
- `SFTP_PASSWORD` - FTP password

### Workflow Triggers

The deployment workflow triggers on:

- Push to `main` branch
- Changes to `apps/*/version.json` or root `package.json`

## 🚨 Troubleshooting

### Deployment Not Triggered

- Ensure you've updated the `version` field in the app's `version.json`
- Check that the file was committed and pushed to `main` branch
- Verify the workflow file path patterns match your changes

### Build Failures

- Check the GitHub Actions logs for specific error details
- Ensure all dependencies are properly listed in the root `package.json`
- Verify the build command works locally: `nx build <app-name>`

### Deploy Failures

- Check FTP credentials and server connectivity
- Verify the target directory paths are correct
- Review server-side logs if available

## 📝 Best Practices

1. **Semantic Versioning**: Use semantic versioning (MAJOR.MINOR.PATCH)

   - PATCH: Bug fixes, small updates
   - MINOR: New features, backward compatible
   - MAJOR: Breaking changes

2. **Clear Commit Messages**: Use descriptive commit messages

   ```bash
   git commit -m "🚀 Deploy landing v0.0.3: Add contact form validation"
   ```

3. **Test Before Deploy**: Always test changes locally before bumping versions

4. **Single App Changes**: Deploy one app at a time unless changes are related

5. **Version Coordination**: Keep related apps in sync when they share dependencies

## 🔄 Migration from package.json

This system uses simple `version.json` files instead of separate `package.json` files:

| Old Approach                  | New Approach                    |
| ----------------------------- | ------------------------------- |
| Separate package.json per app | Simple version.json per app     |
| Build conflicts with nx       | Clean nx builds                 |
| Dependency management issues  | Uses root package.json for deps |
| Complex structure             | Minimal version tracking        |

**The system maintains backward compatibility** - if no app version changes are detected, it will fall back to checking for deployment metadata in the root `package.json`.

---

_Last updated: January 2025_
