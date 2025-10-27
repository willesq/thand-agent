#!/bin/bash

# Setup script for Jekyll documentation

set -e

echo "🚀 Setting up Jekyll documentation environment..."

# Initialize rbenv if available
if command -v rbenv &> /dev/null; then
    echo "📋 Initializing rbenv..."
    export PATH="$HOME/.rbenv/bin:$PATH"
    eval "$(rbenv init -)"
    
    # Set local Ruby version
    if [ -f ".ruby-version" ]; then
        RUBY_VERSION=$(cat .ruby-version)
        echo "📌 Setting Ruby version to $RUBY_VERSION..."
        rbenv local $RUBY_VERSION
        rbenv rehash
    fi
fi

# Check Ruby version
RUBY_VERSION=$(ruby --version)
echo "📌 Using Ruby: $RUBY_VERSION"

# Navigate to docs directory
cd "$(dirname "$0")"

# Clean up any existing bundle
if [ -f "Gemfile.lock" ]; then
    echo "🧹 Cleaning existing Gemfile.lock..."
    rm -f Gemfile.lock
fi

# Install bundler if not present
if ! command -v bundle &> /dev/null; then
    echo "💎 Installing bundler..."
    gem install bundler
fi

# Update bundler to latest
echo "🔄 Updating bundler..."
bundle update --bundler

# Install dependencies
echo "📦 Installing Jekyll dependencies..."
bundle install

# Add Linux platform for GitHub Actions
echo "🐧 Adding Linux platform support..."
bundle lock --add-platform x86_64-linux

echo "✅ Setup complete! You can now run:"
echo "   bundle exec jekyll serve"
echo ""
echo "🌐 Your site will be available at: http://localhost:4000"