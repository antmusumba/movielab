#!/bin/bash

echo "🎬 MovieLab Setup Script"
echo "========================"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

echo "✅ Go is installed"

# Check if environment variables are set
if [ -z "$TMDB_API_KEY" ]; then
    echo "❌ TMDB_API_KEY is not set"
    echo "Please get your API key from: https://www.themoviedb.org/settings/api"
    echo "Then run: export TMDB_API_KEY=your_api_key_here"
    exit 1
fi

if [ -z "$OMDB_API_KEY" ]; then
    echo "❌ OMDB_API_KEY is not set"
    echo "Please get your API key from: http://www.omdbapi.com/"
    echo "Then run: export OMDB_API_KEY=your_api_key_here"
    exit 1
fi

echo "✅ API keys are configured"

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod tidy

if [ $? -eq 0 ]; then
    echo "✅ Dependencies installed successfully"
else
    echo "❌ Failed to install dependencies"
    exit 1
fi

# Create static directory if it doesn't exist
if [ ! -d "static" ]; then
    echo "📁 Creating static directory..."
    mkdir -p static
fi

echo ""
echo "🚀 Setup complete! You can now run the application with:"
echo "   go run main.go"
echo ""
echo "The application will be available at: http://localhost:8080"
echo ""
echo "📝 Note: Make sure to keep your API keys secure and never commit them to version control." 