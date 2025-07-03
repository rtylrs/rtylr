#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Starting deployment process...${NC}"

# Check if package.json exists
if [ ! -f "package.json" ]; then
    echo -e "${RED}❌ Error: package.json not found${NC}"
    exit 1
fi

# Get current version from package.json
CURRENT_VERSION=$(node -p "require('./package.json').version")
echo -e "${YELLOW}📦 Current version: ${CURRENT_VERSION}${NC}"

# Parse version components
IFS='.' read -ra VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR=${VERSION_PARTS[0]}
MINOR=${VERSION_PARTS[1]}
PATCH=${VERSION_PARTS[2]}

# Increment patch version
NEW_PATCH=$((PATCH + 1))

# Check if patch version needs to roll over
if [ $NEW_PATCH -gt 99 ]; then
    NEW_PATCH=0
    NEW_MINOR=$((MINOR + 1))
    
    # Check if minor version needs to roll over
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

# Create new version string
NEW_VERSION="${NEW_MAJOR}.${NEW_MINOR}.${NEW_PATCH}"

echo -e "${GREEN}📈 New version: ${NEW_VERSION}${NC}"

# Update package.json with new version
node -e "
const fs = require('fs');
const packageJson = JSON.parse(fs.readFileSync('package.json', 'utf8'));
packageJson.version = '${NEW_VERSION}';
fs.writeFileSync('package.json', JSON.stringify(packageJson, null, 2) + '\n');
"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Successfully updated package.json to version ${NEW_VERSION}${NC}"
else
    echo -e "${RED}❌ Failed to update package.json${NC}"
    exit 1
fi

# Build the project
echo -e "${BLUE}🔨 Building project...${NC}"
npm run build

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Build completed successfully${NC}"
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

# Git operations
echo -e "${BLUE}📝 Committing changes...${NC}"

# Add all changes
git add .

# Commit with version bump message
git commit -m "🚀 Release version ${NEW_VERSION}"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Changes committed successfully${NC}"
else
    echo -e "${RED}❌ Failed to commit changes${NC}"
    exit 1
fi

# Create and push tag
echo -e "${BLUE}🏷️  Creating git tag...${NC}"
git tag -a "v${NEW_VERSION}" -m "Release version ${NEW_VERSION}"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Git tag created successfully${NC}"
else
    echo -e "${RED}❌ Failed to create git tag${NC}"
    exit 1
fi

# Push changes and tags
echo -e "${BLUE}📤 Pushing to remote repository...${NC}"
git push origin main
git push origin "v${NEW_VERSION}"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Successfully pushed to remote repository${NC}"
else
    echo -e "${RED}❌ Failed to push to remote repository${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
echo -e "${GREEN}📦 Version ${NEW_VERSION} has been deployed${NC}"