#!/bin/bash

# K6 Installation Script
# Installs K6 load testing tool

echo "🔧 K6 Installation Script"
echo "========================="

# Check if K6 is already installed
if command -v k6 &> /dev/null; then
    echo "✅ K6 is already installed"
    k6 version
    exit 0
fi

echo "📦 Installing K6..."

# Detect OS and install accordingly
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    echo "🍎 Detected macOS"
    if command -v brew &> /dev/null; then
        echo "Installing via Homebrew..."
        brew install k6
    else
        echo "❌ Homebrew not found. Please install Homebrew first:"
        echo "   /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        exit 1
    fi
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    echo "🐧 Detected Linux"
    if command -v apt-get &> /dev/null; then
        echo "Installing via apt..."
        sudo gpg -k
        sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
        echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
        sudo apt-get update
        sudo apt-get install k6
    elif command -v yum &> /dev/null; then
        echo "Installing via yum..."
        sudo yum install https://dl.k6.io/rpm/repo.rpm
        sudo yum install k6
    else
        echo "❌ Unsupported package manager. Please install K6 manually:"
        echo "   Visit: https://k6.io/docs/getting-started/installation/"
        exit 1
    fi
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
    # Windows
    echo "🪟 Detected Windows"
    if command -v winget &> /dev/null; then
        echo "Installing via winget..."
        winget install k6
    elif command -v chocolatey &> /dev/null; then
        echo "Installing via Chocolatey..."
        choco install k6
    else
        echo "❌ Please install K6 manually:"
        echo "   Visit: https://k6.io/docs/getting-started/installation/"
        exit 1
    fi
else
    echo "❌ Unsupported OS: $OSTYPE"
    echo "Please install K6 manually:"
    echo "   Visit: https://k6.io/docs/getting-started/installation/"
    exit 1
fi

# Verify installation
if command -v k6 &> /dev/null; then
    echo ""
    echo "✅ K6 installed successfully!"
    k6 version
    echo ""
    echo "🚀 You can now run load tests:"
    echo "   ./k6-test.sh"
    echo "   ./run-all-tests.sh"
else
    echo "❌ Installation failed. Please install K6 manually:"
    echo "   Visit: https://k6.io/docs/getting-started/installation/"
    exit 1
fi

