#!/bin/bash
# Batch compile programs for multiple platforms: armv8-linux, armv7-linux, x86-Windows, amd64-Windows, amd64-Linux

# Define build function (Parameter 1: Target OS GOOS, Parameter 2: Target architecture GOARCH, Parameter 3: Output file name)
build() {
    local GOOS=$1
    local GOARCH=$2
    local OUTPUT=$3
    echo -e "\n=== Start compiling $GOOS/$GOARCH target ==="
    # Temporarily set environment variables and execute compilation (CGO_ENABLED=0 disables CGO to ensure cross-platform compatibility)
    GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -o $OUTPUT main.go
    # Check compilation result
    if [ $? -eq 0 ]; then
        echo "✅ Compilation successful: $OUTPUT (File size: $(du -sh $OUTPUT | cut -f1))"
    else
        echo "❌ Compilation failed: $GOOS/$GOARCH"
    fi
}

# 1. armv8-linux (64-bit ARM architecture Linux, such as Raspberry Pi 4 64-bit, ARM servers)
build "linux" "arm64" "phone-generator_armv8-linux"

# 2. armv7-linux (32-bit ARM architecture Linux, such as Raspberry Pi 2/3 32-bit, embedded devices)
build "linux" "arm"   "phone-generator_armv7-linux"

# 3. x86-Windows (32-bit x86 architecture Windows, compatible with older 32-bit systems)
build "windows" "386" "phone-generator_x86-windows.exe"

# 4. amd64-Windows (64-bit x86_64 architecture Windows, mainstream Windows 10/11 systems)
build "windows" "amd64" "phone-generator_amd64-windows.exe"

# 5. amd64-Linux (64-bit x86_64 architecture Linux, mainstream Linux servers/desktop systems, such as Ubuntu 64-bit, CentOS 64-bit)
build "linux" "amd64" "phone-generator_amd64-linux"

echo -e "\n====================================="
echo "All platform compilation tasks completed!"
echo "List of generated files:"
ls -l phone-generator_*  # List all compilation products