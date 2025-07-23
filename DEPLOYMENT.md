# Vbitez Multi-App Deployment System

This project uses a sophisticated per-app versioning and deployment system that allows each application to be versioned and deployed independently.

## 📁 App Structure

The project contains three main applications:
- **Landing** (`apps/landing/`) - Main marketing/public website
- **Dashboard** (`apps/dashboard/`) - Admin dashboard application  
- **Menu** (`apps/menu/`) - Digital menu application

Each app has its own `package.json` file with independent versioning.

## 🔄 How Deployment Works

### Automatic Deployment Triggers

The GitHub workflow automatically deploys apps when:

1. **Version Changes**: When you update the `version` field in an app's `package.json` file
2. **File Changes**: The workflow monitors changes to `apps/*/package.json` files
3. **Selective Deployment**: Only apps with version changes are built and deployed

### Current App Versions

- **Landing**: v1.0.0
- **Dashboard**: v1.0.0  
- **Menu**: v1.0.0

## 🚀 Deploying Apps

### Method 1: Using the Bump Version Script (Recommended)

Use the provided utility script to bump versions easily:

```bash
# Automatically bump patch version with rollover logic
node scripts/bump-version.js landing     # 0.0.1 → 0.0.2
node scripts/bump-version.js dashboard   # 0.0.99 → 0.1.0 (rollover)  
node scripts/bump-version.js menu        # 0.99.99 → 1.0.0 (rollover)
```

The script uses the same auto-increment logic as deploy.sh:
- Patch increments by 1 (0.0.1 → 0.0.2)
- At patch 100, rolls to next minor (0.0.99 → 0.1.0)
- At minor 100, rolls to next major (0.99.99 → 1.0.0)

Then commit and push:
```bash
git add apps/landing/package.json
git commit -m "🚀 Bump landing version to v1.0.1"
git push origin main
```

### Method 2: Manual Version Update

1. Edit the app's `package.json` file:
   ```json
   {
     "name": "@vbitez/landing",
     "version": "1.0.1",  // ← Update this
     ...
   }
   ```

2. Commit and push the change:
   ```bash
   git add apps/landing/package.json
   git commit -m "🚀 Deploy landing v1.0.1"
   git push origin main
   ```

### Method 3: Deploy Multiple Apps

To deploy multiple apps simultaneously, update their versions in the same commit:

```bash
# Update multiple app versions
node scripts/bump-version.js landing patch
node scripts/bump-version.js dashboard patch

# Commit all changes
git add apps/*/package.json
git commit -m "🚀 Deploy landing v1.0.1 and dashboard v1.0.1"
git push origin main
```

## 📋 Deployment Process

When you push version changes, the workflow:

1. **🔍 Detects Changes**: Scans for version changes in app `package.json` files
2. **🏗️ Builds Apps**: Only builds apps that have version changes
3. **🚀 Deploys**: Deploys each app to its respective subdomain:
   - Landing → `vbitez.voxire.com`
   - Dashboard → `dashboard.vbitez.voxire.com`
   - Menu → `menu.vbitez.voxire.com`
4. **📱 Notifies**: Sends success notification to Telegram with deployment details

## 🎯 Deployment Targets

| App | Build Command | Deploy Location |
|-----|---------------|-----------------|
| Landing | `nx build landing` | `/subdomains/vbitez.voxire.com/` |
| Dashboard | `nx build dashboard` | `/subdomains/dashboard.vbitez.voxire.com/` |
| Menu | `nx build menu` | `/subdomains/menu.vbitez.voxire.com/` |

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
   git add apps/<app-name>/package.json
   git commit -m "✨ Add new feature to <app-name>"
   git push origin main
   ```

### For Hotfixes
```bash
# Quick deployment (auto-increments version)
node scripts/bump-version.js <app-name>
git add apps/<app-name>/package.json
git commit -m "🐛 Hotfix for <app-name>"
git push origin main
```

### For Any Updates
```bash
# All deployments use the same simple command
node scripts/bump-version.js <app-name>
git add apps/<app-name>/package.json
git commit -m "🚀 Deploy <app-name> updates"
git push origin main
```

## 🔧 Configuration

### Environment Variables
The workflow uses these secrets:
- `SFTP_HOST` - FTP server hostname
- `SFTP_USERNAME` - FTP username
- `SFTP_PASSWORD` - FTP password

### Workflow Triggers
The deployment workflow triggers on:
- Push to `main` branch
- Changes to `apps/*/package.json` or root `package.json`

## 🚨 Troubleshooting

### Deployment Not Triggered
- Ensure you've updated the `version` field in the app's `package.json`
- Check that the file was committed and pushed to `main` branch
- Verify the workflow file path patterns match your changes

### Build Failures
- Check the GitHub Actions logs for specific error details
- Ensure all dependencies are properly listed in the app's `package.json`
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
   git commit -m "🚀 Deploy landing v1.2.3: Add contact form validation"
   ```

3. **Test Before Deploy**: Always test changes locally before bumping versions

4. **Single App Changes**: Deploy one app at a time unless changes are related

5. **Version Coordination**: Keep related apps in sync when they share dependencies

## 🔄 Migration from deploy.sh

This new system **replaces** the old `deploy.sh` script. Key differences:

| Old System (deploy.sh) | New System (per-app) |
|------------------------|----------------------|
| Single main package.json version | Individual app versions |
| Deployment metadata in root | Version changes trigger deployment |
| Manual script execution | Simple npm scripts |
| All apps deploy together | Selective app deployment |

**The system maintains backward compatibility** - if no app version changes are detected, it will fall back to checking for deployment metadata in the root `package.json`.

---

*Last updated: $(date)* 