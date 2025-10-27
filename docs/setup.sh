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

# Ensure bundler isn't in frozen mode
echo "🔓 Ensuring bundler is not in frozen mode..."
bundle config unset frozen 2>/dev/null || true

# Clean up any existing bundle cache (optional)
if [ -f "Gemfile.lock" ]; then
    echo "🧹 Removing existing Gemfile.lock for fresh dependency resolution..."
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

echo "✅ Setup complete! You can now run:"
echo "   bundle exec jekyll serve"
echo ""
echo "🌐 Your site will be available at: http://localhost:4000"