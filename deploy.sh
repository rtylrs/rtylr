#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'    

echo -e "${BLUE}🚀 Vbitez Multi-App Deployment${NC}"
echo -e "${CYAN}================================${NC}"

# Check if we're in the right directory
if [ ! -f "package.json" ]; then
    echo -e "${RED}❌ Error: package.json not found. Run this script from the workspace root.${NC}"
    exit 1
fi

# Available apps
declare -a APPS=("landing" "dashboard" "menu")
declare -a APP_DESCRIPTIONS=("Landing Page & Main App" "Admin Dashboard" "Menu Display App")

echo -e "${YELLOW}📱 Available apps to deploy:${NC}"
for i in "${!APPS[@]}"; do
    echo -e "${CYAN}  $((i+1)). ${APPS[i]} - ${APP_DESCRIPTIONS[i]}${NC}"
done
echo -e "${CYAN}  $((${#APPS[@]}+1)). all - Deploy all apps${NC}"

echo ""
read -p "$(echo -e ${PURPLE}Please select an app to deploy [1-$((${#APPS[@]}+1))]:${NC} )" choice

# Validate input
if ! [[ "$choice" =~ ^[0-9]+$ ]] || [ "$choice" -lt 1 ] || [ "$choice" -gt $((${#APPS[@]}+1)) ]; then
    echo -e "${RED}❌ Invalid selection. Please choose a number between 1 and $((${#APPS[@]}+1)).${NC}"
    exit 1
fi

# Determine what to deploy
if [ "$choice" -eq $((${#APPS[@]}+1)) ]; then
    DEPLOY_TARGET="all"
    DEPLOY_APPS=("${APPS[@]}")
    echo -e "${GREEN}📦 Selected: Deploy all apps${NC}"
else
    DEPLOY_TARGET="${APPS[$((choice-1))]}"
    DEPLOY_APPS=("$DEPLOY_TARGET")
    echo -e "${GREEN}📦 Selected: ${DEPLOY_TARGET} - ${APP_DESCRIPTIONS[$((choice-1))]}${NC}"
fi

# Get current version
CURRENT_VERSION=$(node -p "require('./package.json').version")
echo -e "${YELLOW}📦 Current version: ${CURRENT_VERSION}${NC}"

# Calculate new version
IFS='.' read -ra VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR=${VERSION_PARTS[0]}
MINOR=${VERSION_PARTS[1]}
PATCH=${VERSION_PARTS[2]}

NEW_PATCH=$((PATCH + 1))

if [ $NEW_PATCH -gt 99 ]; then
    NEW_PATCH=0
    NEW_MINOR=$((MINOR + 1))
    
    if [ $NEW_MINOR -gt 99 ]; then
        NEW_MINOR=0
        NEW_MAJOR=$((MAJOR + 1))
    else
        NEW_MAJOR=$MAJOR
    fi
else
    NEW_MINOR=$MINOR
    NEW_MAJOR=$MAJOR
fi

NEW_VERSION="${NEW_MAJOR}.${NEW_MINOR}.${NEW_PATCH}"
echo -e "${GREEN}📈 New version: ${NEW_VERSION}${NC}"

# Update package.json with version and deployment info
echo -e "${BLUE}📝 Updating package.json...${NC}"
node -e "
const fs = require('fs');
const packageJson = JSON.parse(fs.readFileSync('package.json', 'utf8'));
packageJson.version = '${NEW_VERSION}';
packageJson.deployment = {
  version: '${NEW_VERSION}',
  timestamp: new Date().toISOString(),
  deployTarget: '${DEPLOY_TARGET}',
  apps: [$(printf '\"%s\",' "${DEPLOY_APPS[@]}" | sed 's/,$//')]
};
fs.writeFileSync('package.json', JSON.stringify(packageJson, null, 2) + '\n');
"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Successfully updated package.json to version ${NEW_VERSION}${NC}"
else
    echo -e "${RED}❌ Failed to update package.json${NC}"
    exit 1
fi

# Build selected apps
echo -e "${BLUE}🔨 Building selected apps...${NC}"
BUILD_SUCCESS=true

for app in "${DEPLOY_APPS[@]}"; do
    echo -e "${CYAN}Building ${app}...${NC}"
    npx nx build "$app"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ ${app} built successfully${NC}"
    else
        echo -e "${RED}❌ Failed to build ${app}${NC}"
        BUILD_SUCCESS=false
    fi
done

if [ "$BUILD_SUCCESS" = false ]; then
    echo -e "${RED}❌ Build process failed for one or more apps${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All selected apps built successfully${NC}"

# Git operations
echo -e "${BLUE}📝 Checking for existing changes...${NC}"

if git diff --name-only --cached | grep -v "package.json" | grep -q . || git diff --name-only | grep -v "package.json" | grep -q .; then
    echo -e "${YELLOW}⚠️  Found existing uncommitted changes. Committing them first...${NC}"
    
    git add .
    git commit -m "📝 Commit changes before version ${NEW_VERSION}"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Existing changes committed successfully${NC}"
    else
        echo -e "${RED}❌ Failed to commit existing changes${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✅ No existing changes to commit${NC}"
fi

echo -e "${BLUE}📝 Committing version bump and deployment metadata...${NC}"
git add package.json
git commit -m "🚀 Release version ${NEW_VERSION} - Deploy: ${DEPLOY_TARGET}"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Changes committed successfully${NC}"
else
    echo -e "${RED}❌ Failed to commit changes${NC}"
    exit 1
fi

echo -e "${BLUE}🏷️  Creating git tag...${NC}"
git tag -a "v${NEW_VERSION}" -m "Release version ${NEW_VERSION} - Deploy: ${DEPLOY_TARGET}"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Git tag created successfully${NC}"
else
    echo -e "${RED}❌ Failed to create git tag${NC}"
    exit 1
fi

echo -e "${BLUE}📤 Pushing to remote repository...${NC}"
git push origin main
git push origin "v${NEW_VERSION}"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Successfully pushed to remote repository${NC}"
else
    echo -e "${RED}❌ Failed to push to remote repository${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 Deployment process completed successfully!${NC}"
echo -e "${GREEN}📦 Version ${NEW_VERSION} has been prepared for deployment${NC}"
echo -e "${CYAN}🚀 Target: ${DEPLOY_TARGET}${NC}"
echo -e "${YELLOW}⏳ GitHub Actions will now handle the deployment...${NC}" 