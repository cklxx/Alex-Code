#!/bin/bash

# Alex Website Deployment Script
# This script helps deploy the Alex website to various platforms

set -e

echo "🚀 Alex Website Deployment Script"
echo "================================="

# Check if we're in the right directory
if [ ! -f "index.html" ]; then
    echo "❌ Error: index.html not found. Please run this script from the docs/ directory."
    exit 1
fi

echo "📁 Current directory: $(pwd)"
echo "✅ Found index.html"

# Function to deploy to GitHub Pages
deploy_github_pages() {
    echo ""
    echo "🌐 Deploying to GitHub Pages..."
    echo "--------------------------------"
    
    # Check if git is available
    if ! command -v git &> /dev/null; then
        echo "❌ Git is not installed. Please install Git first."
        exit 1
    fi
    
    # Check if we're in a git repository
    if [ ! -d "../.git" ]; then
        echo "❌ Not in a Git repository. Please initialize Git first:"
        echo "   cd .. && git init && git remote add origin <your-repo-url>"
        exit 1
    fi
    
    echo "📤 Committing website files..."
    cd ..
    git add docs/
    git commit -m "🌐 Update Alex website" || echo "No changes to commit"
    
    echo "📤 Pushing to GitHub..."
    git push origin main
    
    echo ""
    echo "✅ Deployment complete!"
    echo "🔗 Your website should be available at:"
    echo "   https://yourusername.github.io/Alex-Code"
    echo ""
    echo "📝 To enable GitHub Pages:"
    echo "   1. Go to your repository settings"
    echo "   2. Scroll to 'Pages' section"
    echo "   3. Set source to 'Deploy from a branch'"
    echo "   4. Select 'main' branch and '/docs' folder"
}

# Function to deploy to Netlify
deploy_netlify() {
    echo ""
    echo "🌐 Deploying to Netlify..."
    echo "-------------------------"
    
    # Check if netlify CLI is available
    if ! command -v netlify &> /dev/null; then
        echo "📦 Installing Netlify CLI..."
        npm install -g netlify-cli
    fi
    
    echo "🚀 Deploying to Netlify..."
    netlify deploy --prod --dir .
    
    echo "✅ Deployment complete!"
}

# Function to deploy to Vercel
deploy_vercel() {
    echo ""
    echo "🌐 Deploying to Vercel..."
    echo "------------------------"
    
    # Check if vercel CLI is available
    if ! command -v vercel &> /dev/null; then
        echo "📦 Installing Vercel CLI..."
        npm install -g vercel
    fi
    
    echo "🚀 Deploying to Vercel..."
    vercel --prod
    
    echo "✅ Deployment complete!"
}

# Function to start local server
start_local() {
    echo ""
    echo "🖥️  Starting local development server..."
    echo "---------------------------------------"
    
    # Try different methods to serve the site locally
    if command -v python3 &> /dev/null; then
        echo "🐍 Using Python 3 server..."
        echo "🌐 Website available at: http://localhost:8000"
        echo "⏹️  Press Ctrl+C to stop"
        python3 -m http.server 8000
    elif command -v python &> /dev/null; then
        echo "🐍 Using Python 2 server..."
        echo "🌐 Website available at: http://localhost:8000"
        echo "⏹️  Press Ctrl+C to stop"
        python -m SimpleHTTPServer 8000
    elif command -v npx &> /dev/null; then
        echo "📦 Using npx serve..."
        echo "🌐 Website will open automatically"
        npx serve .
    else
        echo "❌ No suitable server found. Please install Python or Node.js"
        echo "   Python: https://python.org"
        echo "   Node.js: https://nodejs.org"
        exit 1
    fi
}

# Function to run validation checks
validate_website() {
    echo ""
    echo "🔍 Validating website..."
    echo "----------------------"
    
    # Check HTML structure
    if grep -q "<title>" index.html; then
        echo "✅ Title tag found"
    else
        echo "❌ Missing title tag"
    fi
    
    if grep -q "viewport" index.html; then
        echo "✅ Viewport meta tag found"
    else
        echo "❌ Missing viewport meta tag"
    fi
    
    if grep -q "description" index.html; then
        echo "✅ Meta description found"
    else
        echo "⚠️  Consider adding meta description for SEO"
    fi
    
    # Check for CSS
    if grep -q "<style>" index.html || grep -q "\.css" index.html; then
        echo "✅ CSS found"
    else
        echo "❌ No CSS found"
    fi
    
    # Check file size
    file_size=$(wc -c < index.html)
    if [ $file_size -lt 1000000 ]; then  # 1MB
        echo "✅ File size OK (${file_size} bytes)"
    else
        echo "⚠️  Large file size (${file_size} bytes) - consider optimization"
    fi
    
    echo ""
    echo "✅ Validation complete!"
}

# Main menu
echo ""
echo "What would you like to do?"
echo ""
echo "1) 🖥️  Start local development server"
echo "2) 🌐 Deploy to GitHub Pages"
echo "3) 🟢 Deploy to Netlify"
echo "4) ▲  Deploy to Vercel"
echo "5) 🔍 Validate website"
echo "6) ❌ Exit"
echo ""

read -p "Enter your choice (1-6): " choice

case $choice in
    1)
        start_local
        ;;
    2)
        deploy_github_pages
        ;;
    3)
        deploy_netlify
        ;;
    4)
        deploy_vercel
        ;;
    5)
        validate_website
        ;;
    6)
        echo "👋 Goodbye!"
        exit 0
        ;;
    *)
        echo "❌ Invalid choice. Please run the script again."
        exit 1
        ;;
esac