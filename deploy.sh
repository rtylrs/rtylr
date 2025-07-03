#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'    

echo -e "${BLUE}🚀 Starting deployment process...${NC}"

if [ ! -f "package.json" ]; then
    echo -e "${RED}❌ Error: package.json not found${NC}"
    exit 1
fi

CURRENT_VERSION=$(node -p "require('./package.json').version")
echo -e "${YELLOW}📦 Current version: ${CURRENT_VERSION}${NC}"

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

echo -e "${BLUE}🔨 Building project...${NC}"
npm run build

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Build completed successfully${NC}"
else
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

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

echo -e "${BLUE}📝 Committing version bump...${NC}"
git add package.json
git commit -m "🚀 Release version ${NEW_VERSION}"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Changes committed successfully${NC}"
else
    echo -e "${RED}❌ Failed to commit changes${NC}"
    exit 1
fi

echo -e "${BLUE}🏷️  Creating git tag...${NC}"
git tag -a "v${NEW_VERSION}" -m "Release version ${NEW_VERSION}"

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

echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
echo -e "${GREEN}📦 Version ${NEW_VERSION} has been deployed${NC}"